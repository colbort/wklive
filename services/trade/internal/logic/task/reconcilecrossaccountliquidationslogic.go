package tasklogic

import (
	"errors"
	"fmt"

	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/services/trade/models"

	"github.com/shopspring/decimal"
)

const (
	crossAccountLiquidationAuditCheck = "CROSS_ACCOUNT_LIQUIDATION"
	crossAccountLiquidationScanLimit  = 500
)

type crossAccountLiquidationAudit struct {
	Id                  int64           `db:"id"`
	TenantId            int64           `db:"tenant_id"`
	LiquidationNo       string          `db:"liquidation_no"`
	MarginAsset         string          `db:"margin_asset"`
	Status              int64           `db:"status"`
	PositionCount       int64           `db:"position_count"`
	StartedAt           int64           `db:"started_at"`
	CompletedAt         int64           `db:"completed_at"`
	GrossSettlement     decimal.Decimal `db:"gross_settlement"`
	PositionMargin      decimal.Decimal `db:"position_margin"`
	LiquidationFee      decimal.Decimal `db:"liquidation_fee"`
	UserCredit          decimal.Decimal `db:"user_credit"`
	UserDebit           decimal.Decimal `db:"user_debit"`
	DeficitAmount       decimal.Decimal `db:"deficit_amount"`
	InsuranceFundAmount decimal.Decimal `db:"insurance_fund_amount"`
	AdlReliefAmount     decimal.Decimal `db:"adl_relief_amount"`
	AdlQty              decimal.Decimal `db:"adl_qty"`
	ItemCount           int64           `db:"item_count"`
	ClosedItemCount     int64           `db:"closed_item_count"`
	ClosedPositionCount int64           `db:"closed_position_count"`
	HistoryCount        int64           `db:"history_count"`
	CompletionEvent     int64           `db:"completion_event"`
	ItemPositionMargin  decimal.Decimal `db:"item_position_margin"`
	ItemRealizedPnl     decimal.Decimal `db:"item_realized_pnl"`
	ItemFee             decimal.Decimal `db:"item_fee"`
	ItemDeficit         decimal.Decimal `db:"item_deficit"`
	ItemAdlRelief       decimal.Decimal `db:"item_adl_relief"`
	ItemAdlQty          decimal.Decimal `db:"item_adl_qty"`
	AdlExecutionCount   int64           `db:"adl_execution_count"`
	AdlCompletedCount   int64           `db:"adl_completed_count"`
	AdlFailedCount      int64           `db:"adl_failed_count"`
	AdlExecutionQty     decimal.Decimal `db:"adl_execution_qty"`
	AdlExecutionRelief  decimal.Decimal `db:"adl_execution_relief"`
	AdlUnreconciled     int64           `db:"adl_unreconciled"`
	NetInstructionCount int64           `db:"net_instruction_count"`
	NetInstructionDone  int64           `db:"net_instruction_done"`
	FeeInstructionCount int64           `db:"fee_instruction_count"`
	FeeInstructionDone  int64           `db:"fee_instruction_done"`
}

func (l *ReconcileContractAssetFlowsLogic) reconcileCrossAccountLiquidations(tenantID int64) error {
	now := utils.NowMillis()
	cursor, err := l.loadReconciliationCursor(tenantID, crossAccountLiquidationAuditCheck, now)
	if err != nil {
		return err
	}
	rows, err := l.findCrossAccountLiquidationAudits(
		tenantID, cursor, now-orderFillStableDelay, crossAccountLiquidationScanLimit,
	)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return l.completeReconciliationCycle(tenantID, crossAccountLiquidationAuditCheck, now)
	}
	for _, row := range rows {
		if err = l.persistCrossAccountLiquidationAudit(row, now); err != nil {
			return err
		}
	}
	return l.advanceReconciliationCursor(
		tenantID, crossAccountLiquidationAuditCheck, rows[len(rows)-1].Id, now,
	)
}

func (l *ReconcileContractAssetFlowsLogic) findCrossAccountLiquidationAudits(
	tenantID, cursor, cutoff int64, limit int,
) ([]*crossAccountLiquidationAudit, error) {
	var rows []*crossAccountLiquidationAudit
	if err := l.svcCtx.ContractReconcileCursorModel.FindCrossAccountLiquidationAudits(
		l.ctx, &rows, tenantID, cursor, cutoff, limit,
	); err != nil {
		return nil, err
	}
	return rows, nil
}

func crossAccountLiquidationAuditMatches(row *crossAccountLiquidationAudit) (bool, string) {
	if row == nil {
		return false, "cross account liquidation audit row is nil"
	}
	if row.Status == models.ContractAccountLiquidationStatusManualReview {
		return false, "cross account liquidation requires manual review"
	}
	if row.Status != models.ContractAccountLiquidationStatusCompleted {
		return true, "deferred while cross account liquidation saga is unfinished"
	}
	if row.CompletedAt <= 0 {
		return false, "completed cross account liquidation has no completion time"
	}
	if row.StartedAt == 0 {
		if row.ItemCount != 0 || row.HistoryCount != 0 || row.CompletionEvent != 0 ||
			row.NetInstructionCount != 0 || row.FeeInstructionCount != 0 ||
			!row.GrossSettlement.IsZero() || !row.LiquidationFee.IsZero() ||
			!row.UserCredit.IsZero() || !row.UserDebit.IsZero() ||
			!row.DeficitAmount.IsZero() || !row.InsuranceFundAmount.IsZero() ||
			!row.AdlReliefAmount.IsZero() || !row.AdlQty.IsZero() {
			return false, "no-op account liquidation contains takeover side effects"
		}
		return true, ""
	}
	if row.PositionCount <= 0 ||
		row.ItemCount != row.PositionCount ||
		row.ClosedItemCount != row.PositionCount ||
		row.ClosedPositionCount != row.PositionCount ||
		row.HistoryCount != row.PositionCount {
		return false, "account liquidation items, positions or histories do not match the parent position count"
	}
	if row.CompletionEvent != 1 {
		return false, "completed account liquidation requires exactly one completion event"
	}
	if !row.ItemPositionMargin.Equal(row.PositionMargin) ||
		!row.ItemPositionMargin.Add(row.ItemRealizedPnl).Equal(row.GrossSettlement) ||
		!row.ItemFee.Equal(row.LiquidationFee) {
		return false, "account liquidation parent amounts differ from item totals"
	}
	if row.UserCredit.IsPositive() && row.UserDebit.IsPositive() {
		return false, "account liquidation cannot credit and debit the user simultaneously"
	}
	if !row.GrossSettlement.Sub(row.LiquidationFee).Add(row.DeficitAmount).Equal(row.UserCredit.Sub(row.UserDebit)) {
		return false, "account liquidation net settlement equation does not balance"
	}
	if row.DeficitAmount.IsNegative() || row.InsuranceFundAmount.IsNegative() ||
		row.AdlReliefAmount.IsNegative() || row.AdlQty.IsNegative() {
		return false, "cross account deficit, insurance or ADL checkpoint is invalid"
	}
	if !crossAmountsEqual(row.DeficitAmount, row.InsuranceFundAmount.Add(row.AdlReliefAmount)) {
		return false, "cross account deficit is not fully covered by insurance fund and ADL"
	}
	if !crossAmountsEqual(row.ItemDeficit, row.AdlReliefAmount) ||
		!crossAmountsEqual(row.ItemAdlRelief, row.AdlReliefAmount) ||
		!row.ItemAdlQty.Equal(row.AdlQty) {
		return false, "cross account item ADL totals differ from parent checkpoints"
	}
	if row.AdlFailedCount > 0 || row.AdlCompletedCount != row.AdlExecutionCount ||
		!row.AdlExecutionQty.Equal(row.AdlQty) ||
		!crossAmountsEqual(row.AdlExecutionRelief, row.AdlReliefAmount) ||
		row.AdlUnreconciled > 0 {
		return false, "cross account ADL executions are incomplete or unreconciled"
	}
	if row.UserCredit.IsPositive() || row.UserDebit.IsPositive() {
		if row.NetInstructionCount != 1 || row.NetInstructionDone != 1 {
			return false, "account liquidation net Asset instruction is missing, duplicated or unreconciled"
		}
	} else if row.NetInstructionCount != 0 {
		return false, "zero net account liquidation must not have a net Asset instruction"
	}
	if row.LiquidationFee.IsPositive() {
		if row.FeeInstructionCount != 1 || row.FeeInstructionDone != 1 {
			return false, "account liquidation fee instruction is missing, duplicated or unreconciled"
		}
	} else if row.FeeInstructionCount != 0 {
		return false, "zero-fee account liquidation must not have a fee instruction"
	}
	return true, ""
}

func (l *ReconcileContractAssetFlowsLogic) persistCrossAccountLiquidationAudit(
	row *crossAccountLiquidationAudit, now int64,
) error {
	issueKey := fmt.Sprintf("CROSS_ACCOUNT_LIQUIDATION:%d", row.Id)
	matched, detail := crossAccountLiquidationAuditMatches(row)
	insuranceActual := "not_used"
	if matched && row.Status == models.ContractAccountLiquidationStatusCompleted &&
		row.DeficitAmount.IsPositive() {
		var err error
		matched, insuranceActual, err = l.crossAccountInsuranceCoverMatches(row)
		if err != nil {
			detail = err.Error()
		} else if !matched {
			detail = "Asset insurance cover differs from cross account checkpoints"
		}
	}
	if matched {
		if row.Status != models.ContractAccountLiquidationStatusCompleted {
			return nil
		}
		return l.svcCtx.ContractReconcileIssueModel.ResolveByKey(
			l.ctx, row.TenantId, issueKey,
			"Cross account liquidation parent, items, positions, histories, events and Asset instructions are consistent",
			now,
		)
	}
	expected := fmt.Sprintf(
		"positions=%d items_closed=%d histories=%d event=1 net=%s fee=%s",
		row.PositionCount, row.PositionCount, row.PositionCount,
		row.UserCredit.Sub(row.UserDebit), row.LiquidationFee,
	)
	actual := fmt.Sprintf(
		"status=%d started=%t completed=%t items=%d/%d positions=%d histories=%d event=%d "+
			"net_instruction=%d/%d fee_instruction=%d/%d deficit=%s insurance=%s adl=%s/%s insurance_asset=%s",
		row.Status, row.StartedAt > 0, row.CompletedAt > 0,
		row.ClosedItemCount, row.ItemCount, row.ClosedPositionCount,
		row.HistoryCount, row.CompletionEvent,
		row.NetInstructionDone, row.NetInstructionCount,
		row.FeeInstructionDone, row.FeeInstructionCount,
		row.DeficitAmount, row.InsuranceFundAmount,
		row.AdlReliefAmount, row.AdlQty, insuranceActual,
	)
	return l.recordContractReconciliationFinding(&models.TContractReconciliationIssue{
		TenantId: row.TenantId, IssueKey: issueKey,
		CheckType: crossAccountLiquidationAuditCheck,
		BizType:   crossAccountLiquidationBizType, BizNo: row.LiquidationNo,
		ExpectedValue: expected, ActualValue: actual, Detail: detail,
		FirstSeenAt: now, LastSeenAt: now, CreateTimes: now, UpdateTimes: now,
	})
}

func (l *ReconcileContractAssetFlowsLogic) crossAccountInsuranceCoverMatches(
	row *crossAccountLiquidationAudit,
) (bool, string, error) {
	resp, err := l.svcCtx.AssetAdminClient.GetInsuranceCover(l.ctx, &asset.GetInsuranceCoverReq{
		TenantId: row.TenantId, LiquidationNo: row.LiquidationNo + "-INSURANCE",
	})
	if err != nil {
		return false, "query_failed", fmt.Errorf("query cross account Asset insurance cover: %w", err)
	}
	if resp == nil || resp.GetBase() == nil || resp.GetBase().GetCode() != 200 {
		return false, "query_rejected", errors.New("cross account Asset insurance cover query rejected")
	}
	requested, requestErr := decimal.NewFromString(resp.GetRequestedAmount())
	covered, coveredErr := decimal.NewFromString(resp.GetCoveredAmount())
	remaining, remainingErr := decimal.NewFromString(resp.GetRemainingAmount())
	actual := fmt.Sprintf(
		"liquidation_id=%d coin=%s requested=%s covered=%s remaining=%s status=%d",
		resp.GetLiquidationId(), resp.GetCoin(), resp.GetRequestedAmount(),
		resp.GetCoveredAmount(), resp.GetRemainingAmount(), resp.GetStatus(),
	)
	if requestErr != nil || coveredErr != nil || remainingErr != nil {
		return false, actual, nil
	}
	return resp.GetLiquidationId() == row.Id &&
		resp.GetLiquidationNo() == row.LiquidationNo+"-INSURANCE" &&
		resp.GetCoin() == row.MarginAsset &&
		resp.GetStatus() == 1 &&
		requested.Equal(row.DeficitAmount) &&
		covered.Equal(row.InsuranceFundAmount) &&
		crossAmountsEqual(remaining, row.AdlReliefAmount) &&
		requested.Equal(covered.Add(remaining)), actual, nil
}
