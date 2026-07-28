package tasklogic

import (
	"fmt"

	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/proto/trade"
	"wklive/services/trade/models"

	"github.com/shopspring/decimal"
)

const (
	liquidationAuditCheck     = "LIQUIDATION_INSURANCE_ADL"
	liquidationAuditScanLimit = 500
)

type contractLiquidationAudit struct {
	Id                    int64           `db:"id"`
	TenantId              int64           `db:"tenant_id"`
	LiquidationNo         string          `db:"liquidation_no"`
	PositionId            int64           `db:"position_id"`
	Status                int64           `db:"status"`
	TriggerQty            decimal.Decimal `db:"trigger_qty"`
	LiquidatedQty         decimal.Decimal `db:"liquidated_qty"`
	InsuranceFundAmount   decimal.Decimal `db:"insurance_fund_amount"`
	AdlQty                decimal.Decimal `db:"adl_qty"`
	CompletedAt           int64           `db:"completed_at"`
	PositionQty           decimal.Decimal `db:"position_qty"`
	PositionMargin        decimal.Decimal `db:"position_margin"`
	IsolatedMargin        decimal.Decimal `db:"isolated_margin"`
	PositionStatus        int64           `db:"position_status"`
	MarginAsset           string          `db:"margin_asset"`
	LiquidationHistory    int64           `db:"liquidation_history"`
	CompletionEvent       int64           `db:"completion_event"`
	AdlExecutionCount     int64           `db:"adl_execution_count"`
	AdlCompletedCount     int64           `db:"adl_completed_count"`
	AdlFailedCount        int64           `db:"adl_failed_count"`
	AdlExecutionQty       decimal.Decimal `db:"adl_execution_qty"`
	AdlReliefAmount       decimal.Decimal `db:"adl_relief_amount"`
	AdlUnreconciledAssets int64           `db:"adl_unreconciled_assets"`
}

func (l *ReconcileContractAssetFlowsLogic) reconcileLiquidations(tenantID int64) error {
	now := utils.NowMillis()
	cursor, err := l.loadReconciliationCursor(tenantID, liquidationAuditCheck, now)
	if err != nil {
		return err
	}
	rows, err := l.findContractLiquidationAudits(tenantID, cursor, now-orderFillStableDelay, liquidationAuditScanLimit)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return l.completeReconciliationCycle(tenantID, liquidationAuditCheck, now)
	}
	for _, row := range rows {
		if err = l.persistLiquidationAudit(row, now); err != nil {
			return err
		}
	}
	return l.advanceReconciliationCursor(tenantID, liquidationAuditCheck, rows[len(rows)-1].Id, now)
}

func (l *ReconcileContractAssetFlowsLogic) findContractLiquidationAudits(tenantID, cursor, cutoff int64, limit int) ([]*contractLiquidationAudit, error) {
	tenantClause := ""
	args := []any{
		int64(trade.PositionActionType_POSITION_ACTION_TYPE_LIQUIDATION),
		int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_SUCCESS),
		cursor,
		cutoff,
	}
	if tenantID > 0 {
		tenantClause = " AND q.tenant_id=?"
		args = append(args, tenantID)
	}
	args = append(args, limit)
	query := `
SELECT
  q.id,q.tenant_id,q.liquidation_no,q.position_id,q.status,q.trigger_qty,q.liquidated_qty,
  q.insurance_fund_amount,q.adl_qty,q.completed_at,
  COALESCE(p.qty,0) AS position_qty,COALESCE(p.position_margin,0) AS position_margin,
  COALESCE(p.isolated_margin,0) AS isolated_margin,COALESCE(p.status,0) AS position_status,
  COALESCE(p.margin_asset,'') AS margin_asset,
  COALESCE((SELECT COUNT(1) FROM t_contract_position_history h
            WHERE h.tenant_id=q.tenant_id AND h.position_id=q.position_id
              AND h.action_key=q.liquidation_no AND h.action_type=?),0) AS liquidation_history,
  COALESCE((SELECT COUNT(1) FROM t_biz_trade_event e
            WHERE e.tenant_id=q.tenant_id AND e.event_no=CONCAT(q.liquidation_no,'-COMPLETED')),0) AS completion_event,
  COALESCE((SELECT COUNT(1) FROM t_contract_adl_execution a
            WHERE a.tenant_id=q.tenant_id AND a.liquidation_id=q.id),0) AS adl_execution_count,
  COALESCE((SELECT SUM(a.status=3) FROM t_contract_adl_execution a
            WHERE a.tenant_id=q.tenant_id AND a.liquidation_id=q.id),0) AS adl_completed_count,
  COALESCE((SELECT SUM(a.status IN (4,5)) FROM t_contract_adl_execution a
            WHERE a.tenant_id=q.tenant_id AND a.liquidation_id=q.id),0) AS adl_failed_count,
  COALESCE((SELECT SUM(CASE WHEN a.status=3 THEN a.qty ELSE 0 END) FROM t_contract_adl_execution a
            WHERE a.tenant_id=q.tenant_id AND a.liquidation_id=q.id),0) AS adl_execution_qty,
  COALESCE((SELECT SUM(CASE WHEN a.status=3 THEN a.relief_amount ELSE 0 END) FROM t_contract_adl_execution a
            WHERE a.tenant_id=q.tenant_id AND a.liquidation_id=q.id),0) AS adl_relief_amount,
  COALESCE((SELECT COUNT(1) FROM t_contract_adl_execution a
            LEFT JOIN t_trade_settlement_instruction i
              ON i.tenant_id=a.tenant_id AND i.instruction_no=a.execution_no
            WHERE a.tenant_id=q.tenant_id AND a.liquidation_id=q.id AND a.asset_credit>0
              AND (i.id IS NULL OR i.status<>? OR i.reconciled_at=0)),0) AS adl_unreconciled_assets
FROM t_contract_liquidation q
LEFT JOIN t_contract_position p ON p.tenant_id=q.tenant_id AND p.id=q.position_id
WHERE q.id>? AND q.update_times<=?` + tenantClause + `
ORDER BY q.id
LIMIT ?`
	var rows []*contractLiquidationAudit
	if err := l.svcCtx.DB.QueryRowsCtx(l.ctx, &rows, query, args...); err != nil {
		return nil, err
	}
	return rows, nil
}

func liquidationAuditMatches(row *contractLiquidationAudit) (bool, string) {
	if row == nil {
		return false, "liquidation audit row is nil"
	}
	completed := row.Status == int64(trade.LiquidationStatus_LIQUIDATION_STATUS_COMPLETED)
	if !completed {
		if row.Status == int64(trade.LiquidationStatus_LIQUIDATION_STATUS_MANUAL_REVIEW) {
			return false, "liquidation requires manual review"
		}
		return true, "deferred while liquidation saga is unfinished"
	}
	if row.CompletedAt <= 0 || !row.LiquidatedQty.Equal(row.TriggerQty) {
		return false, "completed liquidation has invalid completion time or liquidated quantity"
	}
	if !row.PositionQty.IsZero() || !row.PositionMargin.IsZero() || !row.IsolatedMargin.IsZero() ||
		row.PositionStatus != int64(trade.PositionStatus_POSITION_STATUS_CLOSED) {
		return false, "completed liquidation did not close and clear the bankrupt position"
	}
	if row.LiquidationHistory != 1 || row.CompletionEvent != 1 {
		return false, "completed liquidation requires exactly one position history and completion event"
	}
	if row.AdlFailedCount > 0 || row.AdlCompletedCount != row.AdlExecutionCount {
		return false, "ADL executions are failed or unfinished"
	}
	if !row.AdlExecutionQty.Equal(row.AdlQty) {
		return false, "liquidation ADL quantity differs from completed executions"
	}
	if row.AdlUnreconciledAssets > 0 {
		return false, "ADL asset credits are missing, failed or unreconciled"
	}
	if row.InsuranceFundAmount.IsNegative() || row.AdlReliefAmount.IsNegative() {
		return false, "insurance or ADL relief amount is invalid"
	}
	return true, ""
}

func (l *ReconcileContractAssetFlowsLogic) persistLiquidationAudit(row *contractLiquidationAudit, now int64) error {
	issueKey := fmt.Sprintf("LIQUIDATION:%d", row.Id)
	matched, detail := liquidationAuditMatches(row)
	insuranceActual := "not_used"
	if matched && row.Status == int64(trade.LiquidationStatus_LIQUIDATION_STATUS_COMPLETED) && row.InsuranceFundAmount.IsPositive() {
		var err error
		matched, insuranceActual, err = l.insuranceCoverMatches(row)
		if err != nil {
			detail = err.Error()
		} else if !matched {
			detail = "Asset insurance cover differs from liquidation checkpoint and ADL relief"
		}
	}
	if matched {
		if row.Status != int64(trade.LiquidationStatus_LIQUIDATION_STATUS_COMPLETED) {
			return nil
		}
		return l.svcCtx.ContractReconcileIssueModel.ResolveByKey(
			l.ctx, row.TenantId, issueKey, "Liquidation, insurance checkpoint and ADL records are internally consistent", now,
		)
	}
	expected := fmt.Sprintf("trigger_qty=%s closed_position=true history=1 event=1 adl_qty=%s adl_assets_reconciled=true",
		row.TriggerQty, row.AdlQty)
	actual := fmt.Sprintf("status=%d liquidated_qty=%s position_qty=%s margins=%s/%s history=%d event=%d adl=%d/%d qty=%s relief=%s bad_assets=%d insurance=%s insurance_asset=%s",
		row.Status, row.LiquidatedQty, row.PositionQty, row.PositionMargin, row.IsolatedMargin,
		row.LiquidationHistory, row.CompletionEvent, row.AdlCompletedCount, row.AdlExecutionCount,
		row.AdlExecutionQty, row.AdlReliefAmount, row.AdlUnreconciledAssets, row.InsuranceFundAmount, insuranceActual)
	if err := l.recordContractReconciliationFinding(&models.TContractReconciliationIssue{
		TenantId: row.TenantId, IssueKey: issueKey, CheckType: liquidationAuditCheck,
		BizType: "liquidation", BizNo: row.LiquidationNo,
		ExpectedValue: expected, ActualValue: actual, Detail: detail,
		FirstSeenAt: now, LastSeenAt: now, CreateTimes: now, UpdateTimes: now,
	}); err != nil {
		return err
	}
	return nil
}

func (l *ReconcileContractAssetFlowsLogic) insuranceCoverMatches(row *contractLiquidationAudit) (bool, string, error) {
	resp, err := l.svcCtx.AssetAdminClient.GetInsuranceCover(l.ctx, &asset.GetInsuranceCoverReq{
		TenantId: row.TenantId, LiquidationNo: row.LiquidationNo + "-INSURANCE",
	})
	if err != nil {
		return false, "query_failed", fmt.Errorf("query Asset insurance cover: %w", err)
	}
	if resp == nil || resp.GetBase() == nil || resp.GetBase().GetCode() != 200 {
		return false, "query_rejected", fmt.Errorf("Asset insurance cover query rejected")
	}
	requested, requestErr := decimal.NewFromString(resp.GetRequestedAmount())
	covered, coveredErr := decimal.NewFromString(resp.GetCoveredAmount())
	remaining, remainingErr := decimal.NewFromString(resp.GetRemainingAmount())
	actual := fmt.Sprintf("liquidation_id=%d coin=%s requested=%s covered=%s remaining=%s status=%d",
		resp.GetLiquidationId(), resp.GetCoin(), resp.GetRequestedAmount(), resp.GetCoveredAmount(),
		resp.GetRemainingAmount(), resp.GetStatus())
	if requestErr != nil || coveredErr != nil || remainingErr != nil {
		return false, actual, nil
	}
	expectedRequested := row.InsuranceFundAmount.Add(row.AdlReliefAmount)
	return resp.GetLiquidationId() == row.Id &&
		resp.GetLiquidationNo() == row.LiquidationNo+"-INSURANCE" &&
		resp.GetCoin() == row.MarginAsset &&
		resp.GetStatus() == 1 &&
		covered.Equal(row.InsuranceFundAmount) &&
		remaining.Equal(row.AdlReliefAmount) &&
		requested.Equal(expectedRequested) &&
		requested.Equal(covered.Add(remaining)), actual, nil
}
