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
	fillPositionCheck     = "FILL_POSITION_HISTORY"
	fillPositionScanLimit = 500
)

type contractFillPositionAudit struct {
	Id               int64           `db:"id"`
	TenantId         int64           `db:"tenant_id"`
	FillNo           string          `db:"fill_no"`
	OrderId          int64           `db:"order_id"`
	UserId           int64           `db:"user_id"`
	SymbolId         int64           `db:"symbol_id"`
	PositionSide     int64           `db:"position_side"`
	FillQty          decimal.Decimal `db:"fill_qty"`
	FillPrice        decimal.Decimal `db:"fill_price"`
	FillFee          decimal.Decimal `db:"fill_fee"`
	FeeAsset         string          `db:"fee_asset"`
	FillCreateTimes  int64           `db:"fill_create_times"`
	HistoryCount     int64           `db:"history_count"`
	ProjectedQty     decimal.Decimal `db:"projected_qty"`
	ProjectedFee     decimal.Decimal `db:"projected_fee"`
	IdentityMismatch int64           `db:"identity_mismatch"`
	PriceMismatch    int64           `db:"price_mismatch"`
	TimeMismatch     int64           `db:"time_mismatch"`
	VersionMismatch  int64           `db:"version_mismatch"`
	FeeAssetMismatch int64           `db:"fee_asset_mismatch"`
}

func (l *ReconcileContractAssetFlowsLogic) reconcileFillPositionHistories(tenantID int64) error {
	now := utils.NowMillis()
	cursor, err := l.loadReconciliationCursor(tenantID, fillPositionCheck, now)
	if err != nil {
		return err
	}
	rows, err := l.findContractFillPositionAudits(
		tenantID, cursor, now-orderFillStableDelay, fillPositionScanLimit,
	)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return l.completeReconciliationCycle(tenantID, fillPositionCheck, now)
	}
	for _, row := range rows {
		if err = l.persistFillPositionAudit(row, now); err != nil {
			return err
		}
	}
	return l.advanceReconciliationCursor(tenantID, fillPositionCheck, rows[len(rows)-1].Id, now)
}

func (l *ReconcileContractAssetFlowsLogic) findContractFillPositionAudits(tenantID, cursor, cutoff int64, limit int) ([]*contractFillPositionAudit, error) {
	var rows []*contractFillPositionAudit
	if err := l.svcCtx.ContractReconcileCursorModel.FindContractFillPositionAudits(
		l.ctx, &rows, tenantID, cursor, cutoff, limit,
	); err != nil {
		return nil, err
	}
	return rows, nil
}

func fillPositionAuditDifferences(row *contractFillPositionAudit) []string {
	if row == nil {
		return []string{"audit row is nil"}
	}
	var differences []string
	if row.HistoryCount == 0 {
		differences = append(differences, "missing_position_history")
	}
	if row.PositionSide != int64(trade.PositionSide_POSITION_SIDE_NET) && row.HistoryCount != 1 {
		differences = append(differences, "non_net_fill_history_count")
	}
	if row.PositionSide == int64(trade.PositionSide_POSITION_SIDE_NET) &&
		(row.HistoryCount < 1 || row.HistoryCount > 2) {
		differences = append(differences, "net_fill_history_count")
	}
	if !row.ProjectedQty.Equal(row.FillQty) {
		differences = append(differences, "projected_qty")
	}
	if !row.ProjectedFee.Equal(row.FillFee) {
		differences = append(differences, "fee_delta")
	}
	if row.IdentityMismatch > 0 {
		differences = append(differences, "identity")
	}
	if row.PriceMismatch > 0 {
		differences = append(differences, "mark_price")
	}
	if row.TimeMismatch > 0 {
		differences = append(differences, "business_time")
	}
	if row.VersionMismatch > 0 {
		differences = append(differences, "position_version")
	}
	if row.FeeAssetMismatch > 0 {
		differences = append(differences, "fee_asset")
	}
	return differences
}

func (l *ReconcileContractAssetFlowsLogic) persistFillPositionAudit(row *contractFillPositionAudit, now int64) error {
	issueKey := fmt.Sprintf("FILL_POSITION:%d", row.Id)
	differences := fillPositionAuditDifferences(row)
	if len(differences) == 0 {
		return l.svcCtx.ContractReconcileIssueModel.ResolveByKey(
			l.ctx, row.TenantId, issueKey, "Fill matches Position History projection facts", now,
		)
	}
	expected := fmt.Sprintf("fill=%s order=%d user=%d symbol=%d qty=%s price=%s fee=%s %s",
		row.FillNo, row.OrderId, row.UserId, row.SymbolId, row.FillQty, row.FillPrice,
		row.FillFee, row.FeeAsset)
	actual := fmt.Sprintf("histories=%d projected_qty=%s projected_fee=%s identity=%d price=%d time=%d version=%d fee_asset=%d",
		row.HistoryCount, row.ProjectedQty, row.ProjectedFee, row.IdentityMismatch,
		row.PriceMismatch, row.TimeMismatch, row.VersionMismatch, row.FeeAssetMismatch)
	detail := "Fill/Position History mismatch fields: " + strings.Join(differences, ",")
	if err := l.recordContractReconciliationFinding(&models.TContractReconciliationIssue{
		TenantId:      row.TenantId,
		IssueKey:      issueKey,
		CheckType:     fillPositionCheck,
		BizType:       "fill",
		BizNo:         row.FillNo,
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
