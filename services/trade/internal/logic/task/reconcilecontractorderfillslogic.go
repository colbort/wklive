package tasklogic

import (
	"fmt"
	"strings"

	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/trade"
	"wklive/services/trade/models"

	"github.com/shopspring/decimal"
)

const (
	orderFillCheck       = "ORDER_FILL"
	orderFillScanLimit   = 500
	orderFillStableDelay = int64(60 * 1000)
)

type contractOrderFillAudit struct {
	TenantId          int64           `db:"tenant_id"`
	OrderId           int64           `db:"order_id"`
	OrderNo           string          `db:"order_no"`
	Status            int64           `db:"status"`
	ContractValueType int64           `db:"contract_value_type"`
	PriceScale        int64           `db:"price_scale"`
	OrderQty          decimal.Decimal `db:"order_qty"`
	OrderFilledQty    decimal.Decimal `db:"order_filled_qty"`
	OrderFilledAmount decimal.Decimal `db:"order_filled_amount"`
	OrderCanceledQty  decimal.Decimal `db:"order_canceled_qty"`
	OrderAvgPrice     decimal.Decimal `db:"order_avg_price"`
	OrderFee          decimal.Decimal `db:"order_fee"`
	FillCount         int64           `db:"fill_count"`
	FillQty           decimal.Decimal `db:"fill_qty"`
	FillAmount        decimal.Decimal `db:"fill_amount"`
	FillFee           decimal.Decimal `db:"fill_fee"`
	FillAvgPrice      decimal.Decimal `db:"fill_avg_price"`
}

func (l *ReconcileContractAssetFlowsLogic) reconcileOrderFills(tenantID int64) error {
	now := utils.NowMillis()
	cursor, err := l.loadReconciliationCursor(tenantID, orderFillCheck, now)
	if err != nil {
		return err
	}
	rows, err := l.findContractOrderFillAudits(tenantID, cursor, now-orderFillStableDelay, orderFillScanLimit)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return l.completeReconciliationCycle(tenantID, orderFillCheck, now)
	}
	for _, row := range rows {
		if err = l.persistOrderFillAudit(row, now); err != nil {
			return err
		}
	}
	return l.advanceReconciliationCursor(tenantID, orderFillCheck, rows[len(rows)-1].OrderId, now)
}

func (l *ReconcileContractAssetFlowsLogic) findContractOrderFillAudits(tenantID, cursor, cutoff int64, limit int) ([]*contractOrderFillAudit, error) {
	tenantClause := ""
	args := []any{int64(common.ProductType_PRODUCT_TYPE_DERIVATIVE), cursor, cutoff}
	if tenantID > 0 {
		tenantClause = " AND o.tenant_id=?"
		args = append(args, tenantID)
	}
	args = append(args, limit)
	query := `
SELECT
  o.tenant_id,
  o.id AS order_id,
  o.order_no,
  o.status,
  o.contract_value_type,
  s.price_scale,
  o.qty AS order_qty,
  o.filled_qty AS order_filled_qty,
  o.filled_amount AS order_filled_amount,
  o.canceled_qty AS order_canceled_qty,
  o.avg_price AS order_avg_price,
  o.fee AS order_fee,
  COUNT(f.id) AS fill_count,
  COALESCE(SUM(f.qty),0) AS fill_qty,
  COALESCE(SUM(f.amount),0) AS fill_amount,
  COALESCE(SUM(f.fee),0) AS fill_fee,
  COALESCE(CASE
    WHEN SUM(f.qty) IS NULL OR SUM(f.qty)=0 THEN 0
    WHEN o.contract_value_type=2 THEN SUM(f.qty)/SUM(f.qty/f.price)
    ELSE SUM(f.price*f.qty)/SUM(f.qty)
  END,0) AS fill_avg_price
FROM t_trade_order o
JOIN t_trade_symbol s ON s.tenant_id=o.tenant_id AND s.id=o.symbol_id
LEFT JOIN t_trade_fill f ON f.tenant_id=o.tenant_id AND f.order_id=o.id
WHERE o.product_type=? AND o.id>? AND o.update_times<=?` + tenantClause + `
GROUP BY o.tenant_id,o.id,o.order_no,o.status,o.contract_value_type,s.price_scale,o.qty,o.filled_qty,
         o.filled_amount,o.canceled_qty,o.avg_price,o.fee
ORDER BY o.id
LIMIT ?`
	var rows []*contractOrderFillAudit
	if err := l.svcCtx.DB.QueryRowsCtx(l.ctx, &rows, query, args...); err != nil {
		return nil, err
	}
	return rows, nil
}

func orderFillAuditDifferences(row *contractOrderFillAudit) []string {
	if row == nil {
		return []string{"audit row is nil"}
	}
	var differences []string
	if !row.OrderFilledQty.Equal(row.FillQty) {
		differences = append(differences, "filled_qty")
	}
	if !row.OrderFilledAmount.Equal(row.FillAmount) {
		differences = append(differences, "filled_amount")
	}
	if !row.OrderFee.Equal(row.FillFee) {
		differences = append(differences, "fee")
	}
	if row.FillQty.IsPositive() && !pricesEqualAtScale(row.OrderAvgPrice, row.FillAvgPrice, row.PriceScale) {
		differences = append(differences, "avg_price")
	}
	if row.OrderFilledQty.Add(row.OrderCanceledQty).GreaterThan(row.OrderQty) {
		differences = append(differences, "filled_plus_canceled_exceeds_qty")
	}
	if row.Status == int64(trade.OrderStatus_ORDER_STATUS_FILLED) && !row.OrderFilledQty.Equal(row.OrderQty) {
		differences = append(differences, "filled_status_without_full_qty")
	}
	return differences
}

func pricesEqualAtScale(left, right decimal.Decimal, scale int64) bool {
	if scale < 0 {
		scale = 0
	}
	if scale > 18 {
		scale = 18
	}
	return left.Round(int32(scale)).Equal(right.Round(int32(scale)))
}

func (l *ReconcileContractAssetFlowsLogic) persistOrderFillAudit(row *contractOrderFillAudit, now int64) error {
	issueKey := fmt.Sprintf("ORDER_FILL:%d", row.OrderId)
	differences := orderFillAuditDifferences(row)
	if len(differences) == 0 {
		return l.svcCtx.ContractReconcileIssueModel.ResolveByKey(
			l.ctx, row.TenantId, issueKey, "Order projection matches immutable Fill aggregate", now,
		)
	}
	expected := fmt.Sprintf("qty=%s filled=%s amount=%s canceled=%s avg=%s fee=%s status=%d",
		row.OrderQty, row.OrderFilledQty, row.OrderFilledAmount, row.OrderCanceledQty,
		row.OrderAvgPrice, row.OrderFee, row.Status)
	actual := fmt.Sprintf("fills=%d qty=%s amount=%s avg=%s fee=%s",
		row.FillCount, row.FillQty, row.FillAmount, row.FillAvgPrice, row.FillFee)
	detail := "Order/Fill mismatch fields: " + strings.Join(differences, ",")
	if err := l.recordContractReconciliationFinding(&models.TContractReconciliationIssue{
		TenantId:      row.TenantId,
		IssueKey:      issueKey,
		CheckType:     orderFillCheck,
		BizType:       "order",
		BizNo:         row.OrderNo,
		ExpectedValue: expected,
		ActualValue:   actual,
		Detail:        detail,
		FirstSeenAt:   now,
		LastSeenAt:    now,
		CreateTimes:   now,
		UpdateTimes:   now,
	}); err != nil {
		return err
	}
	return nil
}

func (l *ReconcileContractAssetFlowsLogic) loadReconciliationCursor(tenantID int64, checkType string, now int64) (int64, error) {
	_, err := l.svcCtx.DB.ExecCtx(l.ctx, `
INSERT INTO t_contract_reconciliation_cursor
(tenant_id,check_type,last_scanned_id,last_cycle_completed_at,create_times,update_times)
VALUES(?,?,0,0,?,?)
ON DUPLICATE KEY UPDATE update_times=update_times`, tenantID, checkType, now, now)
	if err != nil {
		return 0, err
	}
	var cursor int64
	err = l.svcCtx.DB.QueryRowCtx(l.ctx, &cursor,
		"SELECT last_scanned_id FROM t_contract_reconciliation_cursor WHERE tenant_id=? AND check_type=? LIMIT 1",
		tenantID, checkType)
	return cursor, err
}

func (l *ReconcileContractAssetFlowsLogic) advanceReconciliationCursor(tenantID int64, checkType string, cursor, now int64) error {
	_, err := l.svcCtx.DB.ExecCtx(l.ctx,
		"UPDATE t_contract_reconciliation_cursor SET last_scanned_id=?,update_times=? WHERE tenant_id=? AND check_type=?",
		cursor, now, tenantID, checkType)
	return err
}

func (l *ReconcileContractAssetFlowsLogic) completeReconciliationCycle(tenantID int64, checkType string, now int64) error {
	_, err := l.svcCtx.DB.ExecCtx(l.ctx,
		"UPDATE t_contract_reconciliation_cursor SET last_scanned_id=0,last_cycle_completed_at=?,update_times=? WHERE tenant_id=? AND check_type=?",
		now, now, tenantID, checkType)
	return err
}
