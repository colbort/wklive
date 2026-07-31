package models

import (
	"context"
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type OptionCoinAmount struct {
	Coin   string          `db:"coin"`
	Amount decimal.Decimal `db:"amount"`
}

type OptionOperationsOverview struct {
	AssetPendingCount           int64
	AssetFailedCount            int64
	AssetManualReviewCount      int64
	OldestAssetInstructionTime  int64
	OpenReconciliationCount     int64
	OldestReconciliationTime    int64
	PendingSettlementPriceCount int64
	StaleRiskAccountCount       int64
	OldestRiskCalcTime          int64
	PendingExerciseCount        int64
	OldestExerciseTime          int64
	PendingSettlementCount      int64
	FailedSettlementCount       int64
	OldestSettlementTime        int64
	PendingLiquidationCount     int64
	ExceptionLiquidationCount   int64
	OldestLiquidationTime       int64
	PendingOutboxCount          int64
	OldestOutboxTime            int64
	PendingInboxCount           int64
	OldestInboxTime             int64
	PhysicalExceptionCount      int64
	ComboStaleCount             int64
	ComboManualReviewCount      int64
	OldestComboExceptionTime    int64
	ComboInvariantIssueCount    int64
	ComboIncompleteMatchCount   int64
	InsuranceLedger             []*OptionCoinAmount
	BackstopLiability           []*OptionCoinAmount
	UnresolvedDeficit           []*OptionCoinAmount
}

type optionCountOldest struct {
	Count  int64 `db:"count"`
	Oldest int64 `db:"oldest"`
}

func QueryOptionOperationsOverview(
	ctx context.Context,
	conn sqlx.SqlConn,
	tenantID, riskStaleBefore, comboStaleBefore int64,
) (*OptionOperationsOverview, error) {
	result := &OptionOperationsOverview{}
	type countTarget struct {
		table     string
		condition string
		args      []any
		count     *int64
		oldest    *int64
		timeField string
	}
	targets := []countTarget{
		{"t_option_asset_instruction", "status IN (1,2)", nil, &result.AssetPendingCount, nil, "create_times"},
		{"t_option_asset_instruction", "status = ?", []any{4}, &result.AssetFailedCount, nil, "create_times"},
		{"t_option_asset_instruction", "status = ?", []any{5}, &result.AssetManualReviewCount, nil, "create_times"},
		{"t_option_asset_instruction", "status IN (1,2,4,5)", nil, nil, &result.OldestAssetInstructionTime, "create_times"},
		{"t_option_reconciliation_issue", "status = ?", []any{1}, &result.OpenReconciliationCount, &result.OldestReconciliationTime, "create_times"},
		{"t_option_settlement_price", "status = ?", []any{1}, &result.PendingSettlementPriceCount, nil, "create_times"},
		{"t_option_risk_account", "(last_calc_time = 0 OR last_calc_time < ?)", []any{riskStaleBefore}, &result.StaleRiskAccountCount, &result.OldestRiskCalcTime, "last_calc_time"},
		{"t_option_exercise", "status = ?", []any{1}, &result.PendingExerciseCount, &result.OldestExerciseTime, "create_times"},
		{"t_option_settlement", "status IN (1,2)", nil, &result.PendingSettlementCount, &result.OldestSettlementTime, "create_times"},
		{"t_option_settlement", "status = ?", []any{4}, &result.FailedSettlementCount, nil, "create_times"},
		{"t_option_liquidation", "status IN (1,2)", nil, &result.PendingLiquidationCount, &result.OldestLiquidationTime, "create_times"},
		{"t_option_liquidation", "status IN (4,5,6)", nil, &result.ExceptionLiquidationCount, nil, "create_times"},
		{"t_option_outbox", "status IN (1,2,4,5)", nil, &result.PendingOutboxCount, &result.OldestOutboxTime, "create_times"},
		{"t_option_inbox", "status IN (1,3)", nil, &result.PendingInboxCount, &result.OldestInboxTime, "create_times"},
		{"t_option_physical_delivery_unit", "status IN (3,4,6)", nil, &result.PhysicalExceptionCount, nil, "create_times"},
	}
	for _, target := range targets {
		stat, err := queryOptionCountOldest(
			ctx, conn, target.table, target.condition, target.timeField, tenantID, target.args...,
		)
		if err != nil {
			return nil, err
		}
		if target.count != nil {
			*target.count = stat.Count
		}
		if target.oldest != nil {
			*target.oldest = stat.Oldest
		}
	}
	comboExceptions, err := queryOptionComboExceptions(ctx, conn, tenantID, comboStaleBefore)
	if err != nil {
		return nil, err
	}
	result.ComboStaleCount = comboExceptions.StaleCount
	result.ComboManualReviewCount = comboExceptions.ManualReviewCount
	result.OldestComboExceptionTime = comboExceptions.Oldest
	result.ComboInvariantIssueCount, err = queryOptionComboInvariantIssueCount(ctx, conn, tenantID)
	if err != nil {
		return nil, err
	}
	result.ComboIncompleteMatchCount, err = queryOptionComboIncompleteMatchCount(ctx, conn, tenantID)
	if err != nil {
		return nil, err
	}
	result.InsuranceLedger, err = queryOptionCoinAmounts(ctx, conn, tenantID, `
SELECT coin,COALESCE(SUM(amount),0) amount
FROM t_option_insurance_fund_flow WHERE %s GROUP BY coin ORDER BY coin`)
	if err != nil {
		return nil, err
	}
	result.BackstopLiability, err = queryOptionCoinAmounts(ctx, conn, tenantID, `
SELECT contract.settle_coin coin,COALESCE(SUM(liq.backstop_amount),0) amount
FROM t_option_liquidation liq
JOIN t_option_contract contract
  ON contract.tenant_id=liq.tenant_id AND contract.id=liq.contract_id
WHERE %s AND liq.backstop_amount > 0
GROUP BY contract.settle_coin ORDER BY contract.settle_coin`)
	if err != nil {
		return nil, err
	}
	result.UnresolvedDeficit, err = queryOptionCoinAmounts(ctx, conn, tenantID, `
SELECT contract.settle_coin coin,COALESCE(SUM(liq.remaining_deficit),0) amount
FROM t_option_liquidation liq
JOIN t_option_contract contract
  ON contract.tenant_id=liq.tenant_id AND contract.id=liq.contract_id
WHERE %s AND liq.remaining_deficit > 0
GROUP BY contract.settle_coin ORDER BY contract.settle_coin`)
	if err != nil {
		return nil, err
	}
	return result, nil
}

type optionComboExceptionStats struct {
	StaleCount        int64 `db:"stale_count"`
	ManualReviewCount int64 `db:"manual_review_count"`
	Oldest            int64 `db:"oldest"`
}

func queryOptionComboExceptions(
	ctx context.Context,
	conn sqlx.SqlConn,
	tenantID, staleBefore int64,
) (*optionComboExceptionStats, error) {
	where, args := "1=1", []any{staleBefore, staleBefore}
	if tenantID > 0 {
		where = "tenant_id = ?"
		args = append(args, tenantID)
	}
	query := fmt.Sprintf(`
SELECT
  COALESCE(SUM(CASE WHEN status IN (1,5) AND update_times < ? THEN 1 ELSE 0 END),0) stale_count,
  COALESCE(SUM(CASE WHEN status = 8 THEN 1 ELSE 0 END),0) manual_review_count,
  COALESCE(MIN(CASE
    WHEN (status IN (1,5) AND update_times < ?) OR status = 8 THEN update_times
  END),0) oldest
FROM t_option_combo_order
WHERE %s`, where)
	var result optionComboExceptionStats
	if err := conn.QueryRowCtx(ctx, &result, query, args...); err != nil {
		return nil, err
	}
	return &result, nil
}

func queryOptionComboInvariantIssueCount(
	ctx context.Context,
	conn sqlx.SqlConn,
	tenantID int64,
) (int64, error) {
	parentScope, legScope, orderScope := "1=1", "1=1", "1=1"
	args := make([]any, 0, 3)
	if tenantID > 0 {
		parentScope, legScope, orderScope = "p.tenant_id = ?", "l.tenant_id = ?", "o.tenant_id = ?"
		args = append(args, tenantID, tenantID, tenantID)
	}
	query := fmt.Sprintf(`
SELECT COUNT(1) count
FROM (
  SELECT CONCAT('P:',p.tenant_id,':',p.id) issue_key
  FROM t_option_combo_order p
  LEFT JOIN t_option_combo_order_leg l
    ON l.tenant_id=p.tenant_id AND l.combo_order_id=p.id
  LEFT JOIN t_option_order o
    ON o.tenant_id=l.tenant_id AND o.id=l.child_order_id
  WHERE %s
  GROUP BY p.tenant_id,p.id,p.user_id,p.account_id,p.filled_qty,p.unfilled_qty
  HAVING COUNT(l.id) NOT BETWEEN 2 AND 4
    OR SUM(CASE WHEN l.id IS NOT NULL AND (
      l.filled_qty<>p.filled_qty*l.ratio
      OR l.unfilled_qty<>p.unfilled_qty*l.ratio
      OR o.id IS NULL
      OR o.combo_order_id<>p.id
      OR o.combo_leg_no<>l.leg_no
      OR o.contract_id<>l.contract_id
      OR o.user_id<>p.user_id
      OR o.account_id<>p.account_id
      OR o.side<>l.side
      OR o.position_effect<>l.position_effect
      OR o.qty<>l.qty
      OR o.filled_qty<>l.filled_qty
      OR o.unfilled_qty<>l.unfilled_qty
    ) THEN 1 ELSE 0 END)<>0
  UNION ALL
  SELECT CONCAT('L:',l.tenant_id,':',l.id) issue_key
  FROM t_option_combo_order_leg l
  LEFT JOIN t_option_combo_order p
    ON p.tenant_id=l.tenant_id AND p.id=l.combo_order_id
  WHERE p.id IS NULL AND %s
  UNION ALL
  SELECT CONCAT('O:',o.tenant_id,':',o.id) issue_key
  FROM t_option_order o
  LEFT JOIN t_option_combo_order_leg l
    ON l.tenant_id=o.tenant_id
   AND l.combo_order_id=o.combo_order_id
   AND l.leg_no=o.combo_leg_no
   AND l.child_order_id=o.id
  WHERE o.combo_order_id>0 AND l.id IS NULL AND %s
) issues`, parentScope, legScope, orderScope)
	var result struct {
		Count int64 `db:"count"`
	}
	if err := conn.QueryRowCtx(ctx, &result, query, args...); err != nil {
		return 0, err
	}
	return result.Count, nil
}

func queryOptionComboIncompleteMatchCount(
	ctx context.Context,
	conn sqlx.SqlConn,
	tenantID int64,
) (int64, error) {
	where, args := "combo_match_no<>''", []any{}
	if tenantID > 0 {
		where += " AND tenant_id = ?"
		args = append(args, tenantID)
	}
	query := fmt.Sprintf(`
SELECT COUNT(1) count
FROM (
  SELECT tenant_id,combo_match_no
  FROM t_option_trade
  WHERE %s
  GROUP BY tenant_id,combo_match_no
  HAVING COUNT(*)<>COUNT(DISTINCT combo_leg_no)
    OR MIN(combo_leg_no)<>1
    OR MAX(combo_leg_no)<>COUNT(*)
    OR COUNT(*) NOT BETWEEN 2 AND 4
) issues`, where)
	var result struct {
		Count int64 `db:"count"`
	}
	if err := conn.QueryRowCtx(ctx, &result, query, args...); err != nil {
		return 0, err
	}
	return result.Count, nil
}

func queryOptionCountOldest(
	ctx context.Context,
	conn sqlx.SqlConn,
	table, condition, timeField string,
	tenantID int64,
	args ...any,
) (*optionCountOldest, error) {
	where := condition
	if where == "" {
		where = "1=1"
	}
	if tenantID > 0 {
		where += " AND tenant_id = ?"
		args = append(args, tenantID)
	}
	query := fmt.Sprintf(
		"SELECT COUNT(1) count,COALESCE(MIN(%s),0) oldest FROM %s WHERE %s",
		timeField, table, where,
	)
	var result optionCountOldest
	if err := conn.QueryRowCtx(ctx, &result, query, args...); err != nil {
		return nil, err
	}
	return &result, nil
}

func queryOptionCoinAmounts(
	ctx context.Context,
	conn sqlx.SqlConn,
	tenantID int64,
	queryTemplate string,
) ([]*OptionCoinAmount, error) {
	where, args := "1=1", []any{}
	if tenantID > 0 {
		where = "liq.tenant_id = ?"
		if !strings.Contains(queryTemplate, "liq.") {
			where = "tenant_id = ?"
		}
		args = append(args, tenantID)
	}
	query := fmt.Sprintf(queryTemplate, where)
	var result []*OptionCoinAmount
	if err := conn.QueryRowsCtx(ctx, &result, query, args...); err != nil {
		return nil, err
	}
	return result, nil
}
