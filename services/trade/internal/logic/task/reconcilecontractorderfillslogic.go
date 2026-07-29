package tasklogic

import (
	"fmt"
	"strings"

	"wklive/common/utils"
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
	var rows []*contractOrderFillAudit
	if err := l.svcCtx.ContractReconcileCursorModel.FindContractOrderFillAudits(
		l.ctx, &rows, tenantID, cursor, cutoff, limit,
	); err != nil {
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
	return l.svcCtx.ContractReconcileCursorModel.LoadReconciliationCursor(
		l.ctx, tenantID, checkType, now,
	)
}

func (l *ReconcileContractAssetFlowsLogic) advanceReconciliationCursor(tenantID int64, checkType string, cursor, now int64) error {
	return l.svcCtx.ContractReconcileCursorModel.AdvanceReconciliationCursor(
		l.ctx, tenantID, checkType, cursor, now,
	)
}

func (l *ReconcileContractAssetFlowsLogic) completeReconciliationCycle(tenantID int64, checkType string, now int64) error {
	return l.svcCtx.ContractReconcileCursorModel.CompleteReconciliationCycle(
		l.ctx, tenantID, checkType, now,
	)
}
