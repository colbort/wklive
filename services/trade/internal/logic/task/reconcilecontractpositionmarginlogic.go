package tasklogic

import (
	"fmt"

	"wklive/common/utils"
	"wklive/services/trade/models"

	"github.com/shopspring/decimal"
)

const (
	positionMarginCheck     = "POSITION_MARGIN_ASSET_FLOW"
	positionMarginScanLimit = 500
)

type contractPositionMarginAudit struct {
	Id               int64           `db:"id"`
	TenantId         int64           `db:"tenant_id"`
	UserId           int64           `db:"user_id"`
	SymbolId         int64           `db:"symbol_id"`
	MarginAsset      string          `db:"margin_asset"`
	PositionMargin   decimal.Decimal `db:"position_margin"`
	IsolatedMargin   decimal.Decimal `db:"isolated_margin"`
	MarginConsumed   decimal.Decimal `db:"margin_consumed"`
	MarginReleased   decimal.Decimal `db:"margin_released"`
	UnfinishedCount  int64           `db:"unfinished_count"`
	LiquidationCount int64           `db:"liquidation_count"`
}

func (l *ReconcileContractAssetFlowsLogic) reconcilePositionMarginCustody(tenantID int64) error {
	now := utils.NowMillis()
	cursor, err := l.loadReconciliationCursor(tenantID, positionMarginCheck, now)
	if err != nil {
		return err
	}
	rows, err := l.findContractPositionMarginAudits(
		tenantID, cursor, now-orderFillStableDelay, positionMarginScanLimit,
	)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return l.completeReconciliationCycle(tenantID, positionMarginCheck, now)
	}
	for _, row := range rows {
		if err = l.persistPositionMarginAudit(row, now); err != nil {
			return err
		}
	}
	return l.advanceReconciliationCursor(tenantID, positionMarginCheck, rows[len(rows)-1].Id, now)
}

func (l *ReconcileContractAssetFlowsLogic) findContractPositionMarginAudits(tenantID, cursor, cutoff int64, limit int) ([]*contractPositionMarginAudit, error) {
	var rows []*contractPositionMarginAudit
	if err := l.svcCtx.ContractReconcileCursorModel.FindContractPositionMarginAudits(
		l.ctx, &rows, tenantID, cursor, cutoff, limit,
	); err != nil {
		return nil, err
	}
	return rows, nil
}

func positionMarginAuditMatches(row *contractPositionMarginAudit) (bool, string) {
	if row == nil {
		return false, "position margin audit row is nil"
	}
	if row.UnfinishedCount > 0 {
		return true, "deferred while margin settlement instructions are unfinished"
	}
	if row.LiquidationCount > 0 {
		return true, "deferred to liquidation/insurance/ADL reconciliation"
	}
	expected := row.MarginConsumed.Sub(row.MarginReleased)
	actual := row.PositionMargin.Add(row.IsolatedMargin)
	if expected.IsNegative() {
		return false, "verified margin releases exceed verified margin consumption"
	}
	if !actual.Equal(expected) {
		return false, "position custody balance differs from verified Asset margin flows"
	}
	return true, ""
}

func (l *ReconcileContractAssetFlowsLogic) persistPositionMarginAudit(row *contractPositionMarginAudit, now int64) error {
	issueKey := fmt.Sprintf("POSITION_MARGIN:%d", row.Id)
	matched, detail := positionMarginAuditMatches(row)
	if matched {
		if row.UnfinishedCount > 0 || row.LiquidationCount > 0 {
			return nil
		}
		return l.svcCtx.ContractReconcileIssueModel.ResolveByKey(
			l.ctx, row.TenantId, issueKey, "Position margin equals verified Asset custody flows", now,
		)
	}
	expected := fmt.Sprintf("verified_consumed=%s verified_released=%s custody=%s",
		row.MarginConsumed, row.MarginReleased, row.MarginConsumed.Sub(row.MarginReleased))
	actual := fmt.Sprintf("position_margin=%s isolated_margin=%s total=%s",
		row.PositionMargin, row.IsolatedMargin, row.PositionMargin.Add(row.IsolatedMargin))
	if err := l.recordContractReconciliationFinding(&models.TContractReconciliationIssue{
		TenantId:      row.TenantId,
		IssueKey:      issueKey,
		CheckType:     positionMarginCheck,
		BizType:       "position",
		BizNo:         fmt.Sprintf("%d", row.Id),
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
