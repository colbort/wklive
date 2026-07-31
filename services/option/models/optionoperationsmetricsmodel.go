package models

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type OptionOperationsMetric struct {
	TenantID int64
	Category string
	Count    int64
	Oldest   int64
}

type OptionOperationsAmountMetric struct {
	TenantID int64
	Category string
	Coin     string
	Amount   decimal.Decimal
}

type optionTenantCountOldest struct {
	TenantID int64 `db:"tenant_id"`
	Count    int64 `db:"count"`
	Oldest   int64 `db:"oldest"`
}

type optionTenantAmount struct {
	TenantID int64           `db:"tenant_id"`
	Coin     string          `db:"coin"`
	Amount   decimal.Decimal `db:"amount"`
}

func QueryOptionOperationsMetrics(
	ctx context.Context,
	conn sqlx.SqlConn,
	riskStaleBefore, comboStaleBefore, now int64,
) ([]*OptionOperationsMetric, []*OptionOperationsAmountMetric, error) {
	type countTarget struct {
		category  string
		table     string
		condition string
		timeField string
		args      []any
	}
	targets := []countTarget{
		{"asset_pending", "t_option_asset_instruction", "status IN (1,2)", "create_times", nil},
		{"asset_failed", "t_option_asset_instruction", "status=4", "create_times", nil},
		{"asset_manual_review", "t_option_asset_instruction", "status=5", "create_times", nil},
		{"reconciliation_open", "t_option_reconciliation_issue", "status=1", "create_times", nil},
		{"daily_conservation_issue", "t_option_reconciliation_issue", "check_type=3 AND status=1", "create_times", nil},
		{"settlement_price_pending", "t_option_settlement_price", "status=1", "create_times", nil},
		{"risk_account_stale", "t_option_risk_account", "(last_calc_time=0 OR last_calc_time<?)", "last_calc_time", []any{riskStaleBefore}},
		{"exercise_pending", "t_option_exercise", "status=1", "create_times", nil},
		{"settlement_pending", "t_option_settlement", "status IN (1,2)", "create_times", nil},
		{"settlement_failed", "t_option_settlement", "status=4", "create_times", nil},
		{"liquidation_pending", "t_option_liquidation", "status IN (1,2)", "create_times", nil},
		{"liquidation_exception", "t_option_liquidation", "status IN (4,5,6)", "create_times", nil},
		{"outbox_pending", "t_option_outbox", "status IN (1,2,4,5)", "create_times", nil},
		{"inbox_pending", "t_option_inbox", "status IN (1,3)", "create_times", nil},
		{"physical_delivery_exception", "t_option_physical_delivery_unit", "status IN (3,4,6)", "create_times", nil},
		{"physical_delivery_overdue", "t_option_physical_delivery_unit FORCE INDEX (idx_option_physical_delivery_monitor)", "(status=6 OR (status IN (3,4) AND cure_deadline<=?))", "cure_deadline", []any{now}},
		{"combo_stale", "t_option_combo_order", "status IN (1,5) AND update_times<?", "update_times", []any{comboStaleBefore}},
		{"combo_manual_review", "t_option_combo_order", "status=8", "update_times", nil},
		{"trading_control_unconfigured", "t_option_contract", `status=2 AND (
max_user_long_qty<=0 OR max_user_short_qty<=0 OR max_open_interest<=0
OR order_price_band_ratio<=0 OR circuit_breaker_ratio<=0)`, "update_times", nil},
		{"control_not_configured_recent", "t_option_trading_control_event", "event_type='ORDER_REJECTED' AND reason='CONTROL_NOT_CONFIGURED' AND create_times>=?", "create_times", []any{now - 300}},
		{"circuit_breaker_recent", "t_option_trading_control_event", "event_type='CIRCUIT_BREAKER' AND create_times>=?", "create_times", []any{now - 300}},
		{"trade_correction_exception", "t_option_trade_correction", "status IN (1,5) OR (status=3 AND update_times<?)", "update_times", []any{now - 60}},
	}
	counts := make([]*OptionOperationsMetric, 0, len(targets))
	for _, target := range targets {
		rows, err := queryOptionTenantCountOldest(
			ctx, conn, target.table, target.condition, target.timeField, target.args...,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("query option operations category %s: %w", target.category, err)
		}
		for _, row := range rows {
			counts = append(counts, &OptionOperationsMetric{
				TenantID: row.TenantID,
				Category: target.category,
				Count:    row.Count,
				Oldest:   row.Oldest,
			})
		}
	}
	invariantIssues, err := queryOptionComboInvariantIssuesByTenant(ctx, conn)
	if err != nil {
		return nil, nil, fmt.Errorf("query option combo invariant metrics: %w", err)
	}
	counts = append(counts, invariantIssues...)
	incompleteMatches, err := queryOptionComboIncompleteMatchesByTenant(ctx, conn)
	if err != nil {
		return nil, nil, fmt.Errorf("query option combo match metrics: %w", err)
	}
	counts = append(counts, incompleteMatches...)
	controlMetrics, err := queryOptionControlMetricsByTenant(ctx, conn, now)
	if err != nil {
		return nil, nil, fmt.Errorf("query option trading control metrics: %w", err)
	}
	counts = append(counts, controlMetrics...)
	marketMetrics, err := queryOptionMarketMetricsByTenant(ctx, conn, now)
	if err != nil {
		return nil, nil, fmt.Errorf("query option market freshness metrics: %w", err)
	}
	counts = append(counts, marketMetrics...)
	governanceMetrics, err := queryOptionGovernanceMetricsByTenant(ctx, conn, now)
	if err != nil {
		return nil, nil, fmt.Errorf("query option governance metrics: %w", err)
	}
	counts = append(counts, governanceMetrics...)
	timeSensitiveMetrics, err := queryOptionTimeSensitiveMetricsByTenant(ctx, conn, now)
	if err != nil {
		return nil, nil, fmt.Errorf("query option time-sensitive metrics: %w", err)
	}
	counts = append(counts, timeSensitiveMetrics...)
	dailyReconciliationMetrics, err := queryOptionDailyReconciliationMetricsByTenant(ctx, conn, now)
	if err != nil {
		return nil, nil, fmt.Errorf("query option daily reconciliation metrics: %w", err)
	}
	counts = append(counts, dailyReconciliationMetrics...)

	amounts, err := queryOptionOperationsAmounts(ctx, conn)
	if err != nil {
		return nil, nil, err
	}
	return counts, amounts, nil
}

func queryOptionDailyReconciliationMetricsByTenant(
	ctx context.Context,
	conn sqlx.SqlConn,
	now int64,
) ([]*OptionOperationsMetric, error) {
	type queryTarget struct {
		category string
		query    string
		args     []any
	}
	latestRunCTE := `WITH latest_run AS (
  SELECT tenant_id,status,completed_at,
         ROW_NUMBER() OVER (PARTITION BY tenant_id ORDER BY completed_at DESC,id DESC) row_no
  FROM t_option_reconciliation_run
  WHERE scope=1
)
SELECT tenant_id,1 count,completed_at oldest
FROM latest_run
WHERE row_no=1 AND status=?
ORDER BY tenant_id`
	targets := []queryTarget{
		{
			category: "daily_conservation_heartbeat",
			query: `WITH active_tenants AS (
  SELECT tenant_id FROM t_option_contract WHERE is_deleted=2
  UNION
  SELECT tenant_id FROM t_option_account
  UNION
  SELECT tenant_id FROM t_user_asset WHERE wallet_type=5
), last_success AS (
  SELECT tenant_id,MAX(completed_at) completed_at
  FROM t_option_reconciliation_run
  WHERE scope=2 AND status=1
  GROUP BY tenant_id
)
SELECT active_tenants.tenant_id,0 count,COALESCE(last_success.completed_at,0) oldest
FROM active_tenants
LEFT JOIN last_success ON last_success.tenant_id=active_tenants.tenant_id
WHERE active_tenants.tenant_id>0
ORDER BY active_tenants.tenant_id`,
		},
		{
			category: "daily_conservation_heartbeat_missing",
			query: `WITH active_tenants AS (
  SELECT tenant_id FROM t_option_contract WHERE is_deleted=2
  UNION
  SELECT tenant_id FROM t_option_account
  UNION
  SELECT tenant_id FROM t_user_asset WHERE wallet_type=5
), last_success AS (
  SELECT tenant_id,MAX(completed_at) completed_at
  FROM t_option_reconciliation_run
  WHERE scope=2 AND status=1
  GROUP BY tenant_id
)
SELECT active_tenants.tenant_id,1 count,COALESCE(last_success.completed_at,0) oldest
FROM active_tenants
LEFT JOIN last_success ON last_success.tenant_id=active_tenants.tenant_id
WHERE active_tenants.tenant_id>0
  AND (last_success.completed_at IS NULL OR last_success.completed_at<?)
ORDER BY active_tenants.tenant_id`,
			args: []any{now - 36*60*60},
		},
		{
			category: "daily_mirror_reconciliation_heartbeat",
			query: `WITH active_tenants AS (
  SELECT tenant_id FROM t_option_contract WHERE is_deleted=2
  UNION
  SELECT tenant_id FROM t_option_account
  UNION
  SELECT tenant_id FROM t_user_asset WHERE wallet_type=5
), last_success AS (
  SELECT tenant_id,MAX(completed_at) completed_at
  FROM t_option_reconciliation_run
  WHERE scope=1 AND status=1
  GROUP BY tenant_id
)
SELECT active_tenants.tenant_id,0 count,COALESCE(last_success.completed_at,0) oldest
FROM active_tenants
LEFT JOIN last_success ON last_success.tenant_id=active_tenants.tenant_id
WHERE active_tenants.tenant_id>0
ORDER BY active_tenants.tenant_id`,
		},
		{
			category: "daily_mirror_reconciliation_mismatch",
			query:    latestRunCTE,
			args:     []any{OptionReconciliationRunMismatch},
		},
		{
			category: "daily_mirror_reconciliation_failed",
			query:    latestRunCTE,
			args:     []any{OptionReconciliationRunFailed},
		},
		{
			category: "daily_mirror_reconciliation_missing",
			query: `WITH active_tenants AS (
  SELECT tenant_id FROM t_option_contract WHERE is_deleted=2
  UNION
  SELECT tenant_id FROM t_option_account
  UNION
  SELECT tenant_id FROM t_user_asset WHERE wallet_type=5
), last_success AS (
  SELECT tenant_id,MAX(completed_at) completed_at
  FROM t_option_reconciliation_run
  WHERE scope=1 AND status=1
  GROUP BY tenant_id
)
SELECT active_tenants.tenant_id,1 count,COALESCE(last_success.completed_at,0) oldest
FROM active_tenants
LEFT JOIN last_success ON last_success.tenant_id=active_tenants.tenant_id
WHERE active_tenants.tenant_id>0
  AND (last_success.completed_at IS NULL OR last_success.completed_at<?)
ORDER BY active_tenants.tenant_id`,
			args: []any{now - 36*60*60},
		},
	}
	result := make([]*OptionOperationsMetric, 0, len(targets))
	for _, target := range targets {
		var rows []*optionTenantCountOldest
		if err := conn.QueryRowsCtx(ctx, &rows, target.query, target.args...); err != nil {
			return nil, fmt.Errorf("%s: %w", target.category, err)
		}
		result = append(result, toOptionOperationsMetrics(rows, target.category)...)
	}
	return result, nil
}

func queryOptionTenantCountOldest(
	ctx context.Context,
	conn sqlx.SqlConn,
	table, condition, timeField string,
	args ...any,
) ([]*optionTenantCountOldest, error) {
	query := fmt.Sprintf(
		`SELECT tenant_id,COUNT(1) count,COALESCE(MIN(%s),0) oldest
FROM %s WHERE %s GROUP BY tenant_id ORDER BY tenant_id`,
		timeField, table, condition,
	)
	var rows []*optionTenantCountOldest
	if err := conn.QueryRowsCtx(ctx, &rows, query, args...); err != nil {
		return nil, err
	}
	return rows, nil
}

func queryOptionControlMetricsByTenant(
	ctx context.Context,
	conn sqlx.SqlConn,
	now int64,
) ([]*OptionOperationsMetric, error) {
	type queryTarget struct {
		category string
		query    string
		args     []any
	}
	targets := []queryTarget{
		{
			category: "price_band_contract_breach",
			query: `
SELECT tenant_id,COUNT(1) count,COALESCE(MIN(first_time),0) oldest
FROM (
  SELECT tenant_id,contract_id,MIN(create_times) first_time
  FROM t_option_trading_control_event
  WHERE event_type='ORDER_REJECTED' AND reason='ORDER_PRICE_BAND'
    AND create_times>=?
  GROUP BY tenant_id,contract_id
  HAVING COUNT(1)>20
) breaches
GROUP BY tenant_id
ORDER BY tenant_id`,
			args: []any{now - 60},
		},
		{
			category: "stp_user_breach",
			query: `
SELECT tenant_id,COUNT(1) count,COALESCE(MIN(first_time),0) oldest
FROM (
  SELECT tenant_id,user_id,MIN(create_times) first_time
  FROM t_option_trading_control_event
  WHERE event_type='STP_PREVENTED' AND reason='SELF_TRADE_PREVENTED'
    AND create_times>=?
  GROUP BY tenant_id,user_id
  HAVING COUNT(1)>=5
) breaches
GROUP BY tenant_id
ORDER BY tenant_id`,
			args: []any{now - 300},
		},
		{
			category: "stp_tenant_breach",
			query: `
SELECT tenant_id,1 count,COALESCE(MIN(create_times),0) oldest
FROM t_option_trading_control_event
WHERE event_type='STP_PREVENTED' AND reason='SELF_TRADE_PREVENTED'
  AND create_times>=?
GROUP BY tenant_id
HAVING COUNT(1)>50
ORDER BY tenant_id`,
			args: []any{now - 300},
		},
		{
			category: "position_limit_group_breach",
			query: `
SELECT tenant_id,COUNT(1) count,COALESCE(MIN(first_time),0) oldest
FROM (
  SELECT tenant_id,user_id,contract_id,MIN(create_times) first_time
  FROM t_option_trading_control_event
  WHERE event_type='ORDER_REJECTED'
    AND reason IN ('USER_LONG_LIMIT','USER_SHORT_LIMIT','OPEN_INTEREST_LIMIT')
    AND create_times>=?
  GROUP BY tenant_id,user_id,contract_id
  HAVING COUNT(1)>20
) breaches
GROUP BY tenant_id
ORDER BY tenant_id`,
			args: []any{now - 300},
		},
		{
			category: "paused_contract_active_orders",
			query: `
SELECT c.tenant_id,COUNT(DISTINCT c.id) count,
       COALESCE(MIN(halt.started_at),0) oldest
FROM t_option_contract c
JOIN t_option_trading_halt halt
  ON halt.tenant_id=c.tenant_id AND halt.contract_id=c.id AND halt.status=1
JOIN t_option_order o
  ON o.tenant_id=c.tenant_id AND o.contract_id=c.id
WHERE c.status=3 AND halt.started_at<?
  AND o.status IN (1,2,7,8,9)
GROUP BY c.tenant_id
ORDER BY c.tenant_id`,
			args: []any{now - 30},
		},
		{
			category: "kill_switch_active_orders",
			query: `
SELECT control.tenant_id,COUNT(DISTINCT control.user_id) count,
       COALESCE(MIN(control.activated_at),0) oldest
FROM t_option_user_trading_control control
JOIN t_option_order o
  ON o.tenant_id=control.tenant_id AND o.user_id=control.user_id
WHERE control.kill_switch=1 AND control.activated_at>0
  AND control.activated_at<?
  AND o.status IN (1,2,7,8,9)
GROUP BY control.tenant_id
ORDER BY control.tenant_id`,
			args: []any{now - 30},
		},
		{
			category: "kill_switch_release_failure",
			query: `
SELECT control.tenant_id,COUNT(DISTINCT instruction.id) count,
       COALESCE(MIN(instruction.update_times),0) oldest
FROM t_option_user_trading_control control
JOIN t_option_asset_instruction instruction
  FORCE INDEX (idx_option_asset_instruction_control_monitor)
  ON instruction.tenant_id=control.tenant_id
 AND instruction.user_id=control.user_id
WHERE control.kill_switch=1 AND control.activated_at>0
  AND instruction.action=3 AND instruction.status IN (4,5)
  AND instruction.create_times>=control.activated_at
  AND instruction.instruction_no LIKE '%-CONTROL-RELEASE'
GROUP BY control.tenant_id
ORDER BY control.tenant_id`,
		},
		{
			category: "mmp_exception",
			query: `
SELECT config.tenant_id,COUNT(DISTINCT config.id) count,
       COALESCE(MIN(CASE
         WHEN config.status=2 AND config.triggered_at>0 THEN config.triggered_at
         ELSE config.update_times
       END),0) oldest
FROM t_option_mmp_config config
LEFT JOIN t_option_order o
  ON o.tenant_id=config.tenant_id
 AND o.user_id=config.user_id
 AND o.contract_id=config.contract_id
 AND o.mmp=1
 AND o.mmp_group=config.group_code
 AND o.status IN (1,2,7,8,9)
WHERE (
  config.status=2
  AND (
    config.last_error_msg<>''
    OR (config.triggered_at>0 AND config.triggered_at<? AND o.id IS NOT NULL)
  )
) OR (
  config.enabled=1 AND config.status=3 AND config.update_times<?
)
GROUP BY config.tenant_id
ORDER BY config.tenant_id`,
			args: []any{now - 10, now - 60},
		},
	}
	result := make([]*OptionOperationsMetric, 0, len(targets))
	for _, target := range targets {
		var rows []*optionTenantCountOldest
		if err := conn.QueryRowsCtx(ctx, &rows, target.query, target.args...); err != nil {
			return nil, fmt.Errorf("%s: %w", target.category, err)
		}
		result = append(result, toOptionOperationsMetrics(rows, target.category)...)
	}
	return result, nil
}

func queryOptionTimeSensitiveMetricsByTenant(
	ctx context.Context,
	conn sqlx.SqlConn,
	now int64,
) ([]*OptionOperationsMetric, error) {
	var rows []*optionTenantCountOldest
	err := conn.QueryRowsCtx(ctx, &rows, `
SELECT contract.tenant_id,COUNT(DISTINCT exercise.id) count,
       COALESCE(MIN(exercise.create_times),0) oldest
FROM t_option_contract contract
  FORCE INDEX (idx_option_contract_lifecycle_monitor)
JOIN t_option_exercise exercise
  FORCE INDEX (idx_option_exercise_monitor)
  ON exercise.tenant_id=contract.tenant_id
 AND exercise.contract_id=contract.id
WHERE contract.is_deleted=2 AND contract.status IN (2,3)
  AND contract.expire_time>? AND contract.expire_time<=?
  AND exercise.status=1
GROUP BY contract.tenant_id
ORDER BY contract.tenant_id`, now, now+1800)
	if err != nil {
		return nil, fmt.Errorf("exercise_near_expiry: %w", err)
	}
	return toOptionOperationsMetrics(rows, "exercise_near_expiry"), nil
}

func queryOptionMarketMetricsByTenant(
	ctx context.Context,
	conn sqlx.SqlConn,
	now int64,
) ([]*OptionOperationsMetric, error) {
	type queryTarget struct {
		category string
		query    string
		args     []any
	}
	freshnessQuery := func(field string) string {
		return fmt.Sprintf(`
SELECT contract.tenant_id,COUNT(1) count,
       COALESCE(MIN(CASE
         WHEN market.%[1]s BETWEEN 1 AND ? THEN market.%[1]s
         ELSE 0
       END),0) oldest
FROM t_option_contract contract
LEFT JOIN t_option_market market
  ON market.tenant_id=contract.tenant_id AND market.contract_id=contract.id
WHERE contract.status=2
  AND (
    market.id IS NULL OR market.%[1]s<=0
    OR market.%[1]s<? OR market.%[1]s>?
  )
GROUP BY contract.tenant_id
ORDER BY contract.tenant_id`, field)
	}
	targets := []queryTarget{
		{
			category: "underlying_market_stale",
			query:    freshnessQuery("underlying_snapshot_time"),
			args:     []any{now, now - 30, now},
		},
		{
			category: "mark_market_stale",
			query:    freshnessQuery("mark_snapshot_time"),
			args:     []any{now, now - 30, now},
		},
		{
			category: "greeks_market_invalid",
			query: `
SELECT contract.tenant_id,COUNT(1) count,0 oldest
FROM t_option_contract contract
LEFT JOIN t_option_market market
  ON market.tenant_id=contract.tenant_id AND market.contract_id=contract.id
WHERE contract.status=2
  AND (
    market.id IS NULL OR market.greeks_snapshot_time<=0
    OR market.greeks_snapshot_time>?
  )
GROUP BY contract.tenant_id
ORDER BY contract.tenant_id`,
			args: []any{now},
		},
	}
	result := make([]*OptionOperationsMetric, 0, len(targets))
	for _, target := range targets {
		var rows []*optionTenantCountOldest
		if err := conn.QueryRowsCtx(ctx, &rows, target.query, target.args...); err != nil {
			return nil, fmt.Errorf("%s: %w", target.category, err)
		}
		result = append(result, toOptionOperationsMetrics(rows, target.category)...)
	}
	return result, nil
}

func queryOptionGovernanceMetricsByTenant(
	ctx context.Context,
	conn sqlx.SqlConn,
	now int64,
) ([]*OptionOperationsMetric, error) {
	type queryTarget struct {
		category string
		query    string
		args     []any
	}
	targets := []queryTarget{
		{
			category: "settlement_price_overdue",
			query: `
SELECT contract.tenant_id,COUNT(1) count,
       COALESCE(MIN(contract.expire_time),0) oldest
FROM t_option_contract contract
WHERE contract.expire_time>0 AND contract.expire_time<?
  AND contract.status IN (2,3,4)
  AND NOT EXISTS (
    SELECT 1 FROM t_option_settlement_price price
    WHERE price.tenant_id=contract.tenant_id
      AND price.contract_id=contract.id AND price.status=2
  )
GROUP BY contract.tenant_id
ORDER BY contract.tenant_id`,
			args: []any{now - 60},
		},
		{
			category: "settlement_price_invalid",
			query: `
SELECT price.tenant_id,COUNT(1) count,
       COALESCE(MIN(price.confirmed_at),0) oldest
FROM t_option_settlement_price price
JOIN t_option_contract contract
  ON contract.tenant_id=price.tenant_id AND contract.id=price.contract_id
WHERE price.status=2
  AND (
    price.sample_count<contract.settlement_min_samples
    OR price.price_source<>contract.settlement_price_source
    OR price.calculation_method<>contract.settlement_price_method
    OR price.window_end<>contract.expire_time
    OR price.window_start<>contract.expire_time-contract.settlement_window_seconds
    OR price.source_snapshot_ids=''
  )
GROUP BY price.tenant_id
ORDER BY price.tenant_id`,
		},
		{
			category: "portfolio_risk_config_missing",
			query: `
SELECT tenant_id,COUNT(1) count,0 oldest
FROM (
  SELECT position.tenant_id,position.user_id,contract.settle_coin
  FROM t_option_position position
  JOIN t_option_contract contract
    ON contract.tenant_id=position.tenant_id AND contract.id=position.contract_id
  WHERE position.status=1 AND position.position_qty>0
    AND contract.seller_margin_mode=3
    AND NOT EXISTS (
      SELECT 1 FROM t_option_portfolio_risk_config config
      WHERE config.tenant_id=position.tenant_id
        AND config.settle_coin=contract.settle_coin
        AND config.status IN (2,4)
        AND config.effective_from<=?
        AND (config.effective_until=0 OR config.effective_until>?)
    )
  GROUP BY position.tenant_id,position.user_id,contract.settle_coin
) missing
GROUP BY tenant_id
ORDER BY tenant_id`,
			args: []any{now, now},
		},
		{
			category: "portfolio_risk_version_mismatch",
			query: `
SELECT account.tenant_id,COUNT(1) count,
       COALESCE(MIN(account.last_calc_time),0) oldest
FROM t_option_risk_account account
JOIN t_option_portfolio_risk_config config
  ON config.id=(
    SELECT active.id
    FROM t_option_portfolio_risk_config active
    WHERE active.tenant_id=account.tenant_id
      AND active.settle_coin=account.settle_coin
      AND active.status IN (2,4)
      AND active.effective_from<=?
      AND (active.effective_until=0 OR active.effective_until>?)
    ORDER BY active.effective_from DESC,active.version DESC
    LIMIT 1
  )
WHERE account.portfolio_risk_method=1
  AND account.last_calc_time<?
  AND (
    account.portfolio_risk_config_id<>config.id
    OR account.portfolio_risk_config_version<>config.version
  )
GROUP BY account.tenant_id
ORDER BY account.tenant_id`,
			args: []any{now, now, now - 60},
		},
		{
			category: "corporate_action_due",
			query: `
SELECT action.tenant_id,COUNT(1) count,
       COALESCE(MIN(action.effective_time),0) oldest
FROM t_option_corporate_action action
WHERE action.effective_time<=? AND action.status IN (2,4)
GROUP BY action.tenant_id
ORDER BY action.tenant_id`,
			args: []any{now},
		},
		{
			category: "corporate_action_exception",
			query: `
SELECT tenant_id,COUNT(1) count,COALESCE(MIN(effective_time),0) oldest
FROM (
  SELECT action.tenant_id,action.id,action.effective_time
  FROM t_option_corporate_action action
  WHERE action.status IN (6,7)
  UNION
  SELECT action.tenant_id,action.id,action.effective_time
  FROM t_option_corporate_action action
  JOIN t_option_corporate_action_contract mapping
    ON mapping.tenant_id=action.tenant_id AND mapping.action_id=action.id
  LEFT JOIN t_option_contract source_contract
    ON source_contract.tenant_id=mapping.tenant_id
   AND source_contract.id=mapping.source_contract_id
  LEFT JOIN t_option_contract successor_contract
    ON successor_contract.tenant_id=mapping.tenant_id
   AND successor_contract.id=mapping.successor_contract_id
  WHERE action.status IN (2,4,6,7)
    AND (
      mapping.status IN (5,6)
      OR mapping.position_failed>0
      OR source_contract.id IS NULL OR source_contract.status<>3
      OR (mapping.successor_contract_id>0
          AND (successor_contract.id IS NULL OR successor_contract.status<>1))
    )
) issues
GROUP BY tenant_id
ORDER BY tenant_id`,
		},
		{
			category: "contract_series_review_stale",
			query: `
SELECT series.tenant_id,COUNT(1) count,
       COALESCE(MIN(series.create_times),0) oldest
FROM t_option_contract_series series
WHERE series.status=1 AND series.create_times<?
GROUP BY series.tenant_id
ORDER BY series.tenant_id`,
			args: []any{now - 86400},
		},
		{
			category: "contract_series_invariant_issue",
			query: `
SELECT tenant_id,COUNT(1) count,COALESCE(MIN(issue_time),0) oldest
FROM (
  SELECT series.tenant_id,series.id,series.create_times issue_time
  FROM t_option_contract_series series
  WHERE series.status=2
    AND series.generated_contract_count<>series.expected_contract_count
  UNION
  SELECT series.tenant_id,series.id,series.create_times issue_time
  FROM t_option_contract_series series
  LEFT JOIN t_option_contract_series_detail detail
    ON detail.tenant_id=series.tenant_id AND detail.series_id=series.id
  LEFT JOIN t_option_contract contract
    ON contract.tenant_id=detail.tenant_id AND contract.id=detail.contract_id
  WHERE series.status=2
  GROUP BY series.tenant_id,series.id,series.create_times,
           series.expected_contract_count,series.launch_status
  HAVING COUNT(detail.id)<>series.expected_contract_count
    OR SUM(CASE
      WHEN contract.id IS NULL
        OR (series.launch_status<>2 AND contract.status<>1)
      THEN 1 ELSE 0 END)<>0
) issues
GROUP BY tenant_id
ORDER BY tenant_id`,
		},
		{
			category: "public_chain_pair_issue",
			query: `
SELECT tenant_id,COUNT(1) count,0 oldest
FROM (
  SELECT tenant_id,underlying_symbol,expire_time,strike_price
  FROM t_option_contract
  WHERE status=2 AND is_deleted=2
  GROUP BY tenant_id,underlying_symbol,expire_time,strike_price
  HAVING SUM(option_type=1)<>1 OR SUM(option_type=2)<>1
) issues
GROUP BY tenant_id
ORDER BY tenant_id`,
		},
		{
			category: "open_interest_imbalance",
			query: `
SELECT tenant_id,COUNT(1) count,0 oldest
FROM (
  SELECT tenant_id,contract_id
  FROM t_option_position
  WHERE status=1 AND position_qty>0
  GROUP BY tenant_id,contract_id
  HAVING SUM(CASE WHEN side=1 THEN position_qty ELSE 0 END)
      <>SUM(CASE WHEN side=2 THEN position_qty ELSE 0 END)
) issues
GROUP BY tenant_id
ORDER BY tenant_id`,
		},
	}
	result := make([]*OptionOperationsMetric, 0, len(targets))
	for _, target := range targets {
		var rows []*optionTenantCountOldest
		if err := conn.QueryRowsCtx(ctx, &rows, target.query, target.args...); err != nil {
			return nil, fmt.Errorf("%s: %w", target.category, err)
		}
		result = append(result, toOptionOperationsMetrics(rows, target.category)...)
	}
	return result, nil
}

func queryOptionComboInvariantIssuesByTenant(
	ctx context.Context,
	conn sqlx.SqlConn,
) ([]*OptionOperationsMetric, error) {
	query := `
SELECT tenant_id,COUNT(1) count,0 oldest
FROM (
  SELECT p.tenant_id,CONCAT('P:',p.id) issue_key
  FROM t_option_combo_order p
  LEFT JOIN t_option_combo_order_leg l
    ON l.tenant_id=p.tenant_id AND l.combo_order_id=p.id
  LEFT JOIN t_option_order o
    ON o.tenant_id=l.tenant_id AND o.id=l.child_order_id
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
  SELECT l.tenant_id,CONCAT('L:',l.id) issue_key
  FROM t_option_combo_order_leg l
  LEFT JOIN t_option_combo_order p
    ON p.tenant_id=l.tenant_id AND p.id=l.combo_order_id
  WHERE p.id IS NULL
  UNION ALL
  SELECT o.tenant_id,CONCAT('O:',o.id) issue_key
  FROM t_option_order o
  LEFT JOIN t_option_combo_order_leg l
    ON l.tenant_id=o.tenant_id
   AND l.combo_order_id=o.combo_order_id
   AND l.leg_no=o.combo_leg_no
   AND l.child_order_id=o.id
  WHERE o.combo_order_id>0 AND l.id IS NULL
) issues
GROUP BY tenant_id
ORDER BY tenant_id`
	var rows []*optionTenantCountOldest
	if err := conn.QueryRowsCtx(ctx, &rows, query); err != nil {
		return nil, err
	}
	return toOptionOperationsMetrics(rows, "combo_invariant_issue"), nil
}

func queryOptionComboIncompleteMatchesByTenant(
	ctx context.Context,
	conn sqlx.SqlConn,
) ([]*OptionOperationsMetric, error) {
	query := `
SELECT tenant_id,COUNT(1) count,0 oldest
FROM (
  SELECT tenant_id,combo_match_no
  FROM t_option_trade
  WHERE combo_match_no<>''
  GROUP BY tenant_id,combo_match_no
  HAVING COUNT(*)<>COUNT(DISTINCT combo_leg_no)
    OR MIN(combo_leg_no)<>1
    OR MAX(combo_leg_no)<>COUNT(*)
    OR COUNT(*) NOT BETWEEN 2 AND 4
) issues
GROUP BY tenant_id
ORDER BY tenant_id`
	var rows []*optionTenantCountOldest
	if err := conn.QueryRowsCtx(ctx, &rows, query); err != nil {
		return nil, err
	}
	return toOptionOperationsMetrics(rows, "combo_incomplete_match"), nil
}

func toOptionOperationsMetrics(
	rows []*optionTenantCountOldest,
	category string,
) []*OptionOperationsMetric {
	result := make([]*OptionOperationsMetric, 0, len(rows))
	for _, row := range rows {
		result = append(result, &OptionOperationsMetric{
			TenantID: row.TenantID,
			Category: category,
			Count:    row.Count,
			Oldest:   row.Oldest,
		})
	}
	return result
}

func queryOptionOperationsAmounts(
	ctx context.Context,
	conn sqlx.SqlConn,
) ([]*OptionOperationsAmountMetric, error) {
	type amountTarget struct {
		category string
		query    string
	}
	targets := []amountTarget{
		{
			category: "insurance_ledger",
			query: `SELECT tenant_id,coin,COALESCE(SUM(amount),0) amount
FROM t_option_insurance_fund_flow
GROUP BY tenant_id,coin ORDER BY tenant_id,coin`,
		},
		{
			category: "backstop_liability",
			query: `SELECT liq.tenant_id,contract.settle_coin coin,
  COALESCE(SUM(liq.backstop_amount),0) amount
FROM t_option_liquidation liq
JOIN t_option_contract contract
  ON contract.tenant_id=liq.tenant_id AND contract.id=liq.contract_id
WHERE liq.backstop_amount>0
GROUP BY liq.tenant_id,contract.settle_coin
ORDER BY liq.tenant_id,contract.settle_coin`,
		},
		{
			category: "unresolved_deficit",
			query: `SELECT liq.tenant_id,contract.settle_coin coin,
  COALESCE(SUM(liq.remaining_deficit),0) amount
FROM t_option_liquidation liq
JOIN t_option_contract contract
  ON contract.tenant_id=liq.tenant_id AND contract.id=liq.contract_id
WHERE liq.remaining_deficit>0
GROUP BY liq.tenant_id,contract.settle_coin
ORDER BY liq.tenant_id,contract.settle_coin`,
		},
	}
	result := make([]*OptionOperationsAmountMetric, 0)
	for _, target := range targets {
		var rows []*optionTenantAmount
		if err := conn.QueryRowsCtx(ctx, &rows, target.query); err != nil {
			return nil, fmt.Errorf("query option operations amount %s: %w", target.category, err)
		}
		for _, row := range rows {
			result = append(result, &OptionOperationsAmountMetric{
				TenantID: row.TenantID,
				Category: target.category,
				Coin:     row.Coin,
				Amount:   row.Amount,
			})
		}
	}
	return result, nil
}
