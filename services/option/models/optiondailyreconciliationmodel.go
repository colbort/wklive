package models

import (
	"context"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	OptionReconciliationScopeAccountMirror int64 = 1
	OptionReconciliationScopeFullFunds     int64 = 2

	OptionReconciliationRunSucceeded int64 = 1
	OptionReconciliationRunMismatch  int64 = 2
	OptionReconciliationRunFailed    int64 = 3

	OptionReconciliationDimensionUserWallet      int64 = 1
	OptionReconciliationDimensionPlatformAccount int64 = 2
	OptionReconciliationDimensionOptionSubledger int64 = 3

	OptionReconciliationDetailMatched    int64 = 1
	OptionReconciliationDetailMismatch   int64 = 2
	OptionReconciliationDetailIncomplete int64 = 3
)

type OptionAccountMirrorSummary struct {
	Coin            string          `db:"coin"`
	AccountCount    int64           `db:"account_count"`
	MismatchCount   int64           `db:"mismatch_count"`
	OptionTotal     decimal.Decimal `db:"option_total"`
	OptionAvailable decimal.Decimal `db:"option_available"`
	OptionFrozen    decimal.Decimal `db:"option_frozen"`
	AssetTotal      decimal.Decimal `db:"asset_total"`
	AssetAvailable  decimal.Decimal `db:"asset_available"`
	AssetFrozen     decimal.Decimal `db:"asset_frozen"`
	AssetLocked     decimal.Decimal `db:"asset_locked"`
}

type OptionUserWalletConservationSummary struct {
	Coin                  string          `db:"coin"`
	WalletCount           int64           `db:"wallet_count"`
	MismatchWalletCount   int64           `db:"mismatch_wallet_count"`
	IntegrityErrorCount   int64           `db:"integrity_error_count"`
	UnclassifiedFlowCount int64           `db:"unclassified_flow_count"`
	FlowCount             int64           `db:"flow_count"`
	MaxFlowID             int64           `db:"max_flow_id"`
	OpeningAmount         decimal.Decimal `db:"opening_amount"`
	ExternalNet           decimal.Decimal `db:"external_net"`
	OptionNet             decimal.Decimal `db:"option_net"`
	ManualNet             decimal.Decimal `db:"manual_net"`
	ExpectedClosing       decimal.Decimal `db:"expected_closing"`
	ActualClosing         decimal.Decimal `db:"actual_closing"`
	DifferenceAmount      decimal.Decimal `db:"difference_amount"`
}

type OptionPlatformAccountConservationSummary struct {
	AccountType          string          `db:"account_type"`
	Coin                 string          `db:"coin"`
	AccountCount         int64           `db:"account_count"`
	MismatchAccountCount int64           `db:"mismatch_account_count"`
	IntegrityErrorCount  int64           `db:"integrity_error_count"`
	FlowCount            int64           `db:"flow_count"`
	MaxPlatformFlowID    int64           `db:"max_platform_flow_id"`
	OpeningAmount        decimal.Decimal `db:"opening_amount"`
	OptionNet            decimal.Decimal `db:"option_net"`
	ManualNet            decimal.Decimal `db:"manual_net"`
	ExpectedClosing      decimal.Decimal `db:"expected_closing"`
	ActualClosing        decimal.Decimal `db:"actual_closing"`
	DifferenceAmount     decimal.Decimal `db:"difference_amount"`
}

type OptionReconciliationRun struct {
	TenantID      int64
	BusinessDate  string
	Scope         int64
	AttemptNo     int64
	Status        int64
	SnapshotTime  int64
	SnapshotRef   string
	CoinCount     int64
	AccountCount  int64
	MismatchCount int64
	Detail        string
	CompletedAt   int64
}

type OptionReconciliationRunDetail struct {
	RunID, TenantID, Scope, DimensionType int64
	BusinessDate, DimensionKey            string
	OpeningAmount, ExternalNet            decimal.Decimal
	OptionNet, ManualNet                  decimal.Decimal
	ExpectedClosing, ActualClosing        decimal.Decimal
	DifferenceAmount                      decimal.Decimal
	FlowCount, MismatchCount, Status      int64
	EvidenceRef, Detail                   string
	CreateTimes                           int64
}

func ListOptionReconciliationTenantIDs(
	ctx context.Context, conn sqlx.SqlConn, tenantID int64,
) ([]int64, error) {
	if tenantID > 0 {
		return []int64{tenantID}, nil
	}
	var tenantIDs []int64
	err := conn.QueryRowsCtx(ctx, &tenantIDs, `SELECT tenant_id
FROM (
  SELECT tenant_id FROM t_option_contract WHERE is_deleted=2
  UNION
  SELECT tenant_id FROM t_option_account
  UNION
  SELECT tenant_id FROM t_user_asset WHERE wallet_type=5
) tenants
WHERE tenant_id>0
ORDER BY tenant_id`)
	return tenantIDs, err
}

func NextOptionReconciliationAttempt(
	ctx context.Context, conn sqlx.SqlConn, tenantID int64, businessDate string, scope int64,
) (int64, error) {
	var attempt int64
	err := conn.QueryRowCtx(ctx, &attempt, `SELECT COALESCE(MAX(attempt_no),0)+1
FROM t_option_reconciliation_run
WHERE tenant_id=? AND business_date=? AND scope=?`, tenantID, businessDate, scope)
	return attempt, err
}

func HasSuccessfulOptionReconciliationRun(
	ctx context.Context, conn sqlx.SqlConn, tenantID int64, businessDate string, scope int64,
) (bool, error) {
	var count int64
	err := conn.QueryRowCtx(ctx, &count, `SELECT COUNT(1)
FROM t_option_reconciliation_run
WHERE tenant_id=? AND business_date=? AND scope=? AND status=1`,
		tenantID, businessDate, scope)
	return count > 0, err
}

func QueryOptionAccountMirrorSummaries(
	ctx context.Context, conn sqlx.SqlConn, tenantID int64,
) ([]*OptionAccountMirrorSummary, error) {
	var rows []*OptionAccountMirrorSummary
	err := conn.QueryRowsCtx(ctx, &rows, `WITH wallet_keys AS (
  SELECT tenant_id,user_id,margin_coin coin
  FROM t_option_account
  WHERE tenant_id=?
  UNION
  SELECT tenant_id,user_id,coin
  FROM t_user_asset
  WHERE tenant_id=? AND wallet_type=5
), option_balances AS (
  SELECT tenant_id,user_id,margin_coin coin,
         COUNT(1) option_rows,
         SUM(CASE WHEN account_id<>0 THEN 1 ELSE 0 END) legacy_rows,
         SUM(balance) total_amount,
         SUM(available_balance) available_amount,
         SUM(frozen_balance) frozen_amount
  FROM t_option_account
  WHERE tenant_id=?
  GROUP BY tenant_id,user_id,margin_coin
), asset_balances AS (
  SELECT tenant_id,user_id,coin,
         COUNT(1) asset_rows,
         SUM(total_amount) total_amount,
         SUM(available_amount) available_amount,
         SUM(frozen_amount) frozen_amount,
         SUM(locked_amount) locked_amount
  FROM t_user_asset
  WHERE tenant_id=? AND wallet_type=5
  GROUP BY tenant_id,user_id,coin
), compared AS (
  SELECT wallet_key.coin,
         COALESCE(option_balances.total_amount,0) option_total,
         COALESCE(option_balances.available_amount,0) option_available,
         COALESCE(option_balances.frozen_amount,0) option_frozen,
         COALESCE(asset_balances.total_amount,0) asset_total,
         COALESCE(asset_balances.available_amount,0) asset_available,
         COALESCE(asset_balances.frozen_amount,0) asset_frozen,
         COALESCE(asset_balances.locked_amount,0) asset_locked,
         CASE WHEN
           COALESCE(option_balances.option_rows,0)<>1
           OR COALESCE(option_balances.legacy_rows,0)<>0
           OR COALESCE(asset_balances.asset_rows,0)<>1
           OR COALESCE(option_balances.total_amount,0)<>COALESCE(asset_balances.total_amount,0)
           OR COALESCE(option_balances.available_amount,0)<>COALESCE(asset_balances.available_amount,0)
           OR COALESCE(option_balances.frozen_amount,0)<>COALESCE(asset_balances.frozen_amount,0)
           OR COALESCE(asset_balances.locked_amount,0)<>0
           OR COALESCE(asset_balances.total_amount,0)<>
              COALESCE(asset_balances.available_amount,0)+
              COALESCE(asset_balances.frozen_amount,0)+
              COALESCE(asset_balances.locked_amount,0)
         THEN 1 ELSE 0 END mismatch
  FROM wallet_keys wallet_key
  LEFT JOIN option_balances
    ON option_balances.tenant_id=wallet_key.tenant_id
   AND option_balances.user_id=wallet_key.user_id
   AND option_balances.coin=wallet_key.coin
  LEFT JOIN asset_balances
    ON asset_balances.tenant_id=wallet_key.tenant_id
   AND asset_balances.user_id=wallet_key.user_id
   AND asset_balances.coin=wallet_key.coin
)
SELECT coin,COUNT(1) account_count,SUM(mismatch) mismatch_count,
       SUM(option_total) option_total,
       SUM(option_available) option_available,
       SUM(option_frozen) option_frozen,
       SUM(asset_total) asset_total,
       SUM(asset_available) asset_available,
       SUM(asset_frozen) asset_frozen,
       SUM(asset_locked) asset_locked
FROM compared
GROUP BY coin
ORDER BY coin`, tenantID, tenantID, tenantID, tenantID)
	return rows, err
}

func QueryOptionUserWalletConservationSummaries(
	ctx context.Context, conn sqlx.SqlConn, tenantID, startMillis, endMillis, snapshotMillis int64,
) ([]*OptionUserWalletConservationSummary, error) {
	var rows []*OptionUserWalletConservationSummary
	err := conn.QueryRowsCtx(ctx, &rows, `WITH cutoff AS (
  SELECT COALESCE(MAX(id),0) max_flow_id
  FROM t_asset_flow
  WHERE tenant_id=? AND wallet_type=5 AND create_times<?
), prior_ranked AS (
  SELECT f.*,ROW_NUMBER() OVER (
    PARTITION BY f.user_id,f.coin ORDER BY f.create_times DESC,f.id DESC
  ) rn
  FROM t_asset_flow f JOIN cutoff c ON f.id<=c.max_flow_id
  WHERE f.tenant_id=? AND f.wallet_type=5 AND f.create_times<?
), relevant_flows AS (
  SELECT p.*,1 is_prior
  FROM prior_ranked p WHERE p.rn=1
  UNION ALL
  SELECT f.*,NULL rn,0 is_prior
  FROM t_asset_flow f JOIN cutoff c ON f.id<=c.max_flow_id
  WHERE f.tenant_id=? AND f.wallet_type=5
    AND f.create_times>=? AND f.create_times<?
), ordered_flows AS (
  SELECT f.*,
    LAG(after_total_amount) OVER (PARTITION BY user_id,coin ORDER BY create_times,id) prev_total,
    LAG(after_available_amount) OVER (PARTITION BY user_id,coin ORDER BY create_times,id) prev_available,
    LAG(after_frozen_amount) OVER (PARTITION BY user_id,coin ORDER BY create_times,id) prev_frozen,
    LAG(after_locked_amount) OVER (PARTITION BY user_id,coin ORDER BY create_times,id) prev_locked
  FROM relevant_flows f
), checked_flows AS (
  SELECT f.*,(after_total_amount-before_total_amount) total_delta,
    CASE WHEN is_prior=1 THEN 0 WHEN
      before_total_amount<>before_available_amount+before_frozen_amount+before_locked_amount
      OR after_total_amount<>after_available_amount+after_frozen_amount+after_locked_amount
      OR change_amount<=0 OR flow_no='' OR biz_no='' OR biz_type='' OR scene_type=''
      OR (prev_total IS NULL AND before_total_amount<>0)
      OR (prev_total IS NOT NULL AND (
        before_total_amount<>prev_total OR before_available_amount<>prev_available
        OR before_frozen_amount<>prev_frozen OR before_locked_amount<>prev_locked))
      OR (op_type IN (1,9) AND after_total_amount-before_total_amount<>change_amount)
      OR (op_type IN (2,7,8,10) AND after_total_amount-before_total_amount<>-change_amount)
      OR (op_type IN (3,4,5,6) AND after_total_amount<>before_total_amount)
      OR op_type NOT IN (1,2,3,4,5,6,7,8,9,10)
    THEN 1 ELSE 0 END integrity_error,
    CASE WHEN is_prior=0 AND create_times>=? AND create_times<?
      AND NOT (scene_type IN ('manual_add','manual_sub') OR biz_type IN ('option','payment','transfer'))
    THEN 1 ELSE 0 END unclassified
  FROM ordered_flows f
), flow_aggregates AS (
  SELECT user_id,coin,
    COUNT(CASE WHEN is_prior=0 AND create_times>=? AND create_times<? THEN 1 END) flow_count,
    SUM(CASE WHEN is_prior=0 THEN integrity_error ELSE 0 END) integrity_errors,
    SUM(unclassified) unclassified_flows,
    SUM(CASE WHEN is_prior=0 AND create_times>=? AND create_times<?
      AND biz_type IN ('payment','transfer') AND scene_type NOT IN ('manual_add','manual_sub')
      THEN total_delta ELSE 0 END) external_net,
    SUM(CASE WHEN is_prior=0 AND create_times>=? AND create_times<?
      AND biz_type='option' AND scene_type NOT IN ('manual_add','manual_sub')
      THEN total_delta ELSE 0 END) option_net,
    SUM(CASE WHEN is_prior=0 AND create_times>=? AND create_times<?
      AND scene_type IN ('manual_add','manual_sub') THEN total_delta ELSE 0 END) manual_net,
    SUM(CASE WHEN is_prior=0 AND create_times>=? THEN total_delta ELSE 0 END) post_net
  FROM checked_flows
  GROUP BY user_id,coin
), first_day AS (
  SELECT user_id,coin,before_total_amount,ROW_NUMBER() OVER (
    PARTITION BY user_id,coin ORDER BY create_times,id
  ) rn
  FROM checked_flows WHERE is_prior=0 AND create_times>=? AND create_times<?
), first_post AS (
  SELECT user_id,coin,before_total_amount,ROW_NUMBER() OVER (
    PARTITION BY user_id,coin ORDER BY create_times,id
  ) rn
  FROM checked_flows WHERE is_prior=0 AND create_times>=?
), last_flow AS (
  SELECT user_id,coin,after_total_amount,after_available_amount,after_frozen_amount,after_locked_amount,
    ROW_NUMBER() OVER (PARTITION BY user_id,coin ORDER BY create_times DESC,id DESC) rn
  FROM checked_flows
), wallet_keys AS (
  SELECT user_id,coin FROM t_user_asset WHERE tenant_id=? AND wallet_type=5
  UNION
  SELECT user_id,coin FROM checked_flows
), wallet_values AS (
  SELECT k.user_id,k.coin,
    COALESCE(fd.before_total_amount,fp.before_total_amount,a.total_amount) opening_amount,
    COALESCE(g.external_net,0) external_net,COALESCE(g.option_net,0) option_net,
    COALESCE(g.manual_net,0) manual_net,COALESCE(g.flow_count,0) flow_count,
    COALESCE(g.integrity_errors,0)+CASE WHEN a.id IS NULL THEN 1 ELSE 0 END+
      CASE WHEN a.id IS NOT NULL AND a.total_amount<>a.available_amount+a.frozen_amount+a.locked_amount THEN 1 ELSE 0 END+
      CASE WHEN lf.user_id IS NOT NULL AND a.id IS NOT NULL AND (
        a.total_amount<>lf.after_total_amount OR a.available_amount<>lf.after_available_amount
        OR a.frozen_amount<>lf.after_frozen_amount OR a.locked_amount<>lf.after_locked_amount) THEN 1 ELSE 0 END
      integrity_errors,
    COALESCE(g.unclassified_flows,0) unclassified_flows,
    COALESCE(a.total_amount,0)-COALESCE(g.post_net,0) actual_closing
  FROM wallet_keys k
  LEFT JOIN t_user_asset a ON a.tenant_id=? AND a.wallet_type=5 AND a.user_id=k.user_id AND a.coin=k.coin
  LEFT JOIN flow_aggregates g ON g.user_id=k.user_id AND g.coin=k.coin
  LEFT JOIN first_day fd ON fd.user_id=k.user_id AND fd.coin=k.coin AND fd.rn=1
  LEFT JOIN first_post fp ON fp.user_id=k.user_id AND fp.coin=k.coin AND fp.rn=1
  LEFT JOIN last_flow lf ON lf.user_id=k.user_id AND lf.coin=k.coin AND lf.rn=1
), calculated AS (
  SELECT *,opening_amount+external_net+option_net+manual_net expected_closing
  FROM wallet_values
)
SELECT c.coin,COUNT(1) wallet_count,
  SUM(CASE WHEN c.integrity_errors>0 OR c.unclassified_flows>0
    OR c.actual_closing<>c.expected_closing THEN 1 ELSE 0 END) mismatch_wallet_count,
  SUM(c.integrity_errors) integrity_error_count,
  SUM(c.unclassified_flows) unclassified_flow_count,SUM(c.flow_count) flow_count,
  MAX(cutoff.max_flow_id) max_flow_id,SUM(c.opening_amount) opening_amount,
  SUM(c.external_net) external_net,SUM(c.option_net) option_net,SUM(c.manual_net) manual_net,
  SUM(c.expected_closing) expected_closing,SUM(c.actual_closing) actual_closing,
  SUM(c.actual_closing-c.expected_closing) difference_amount
FROM calculated c CROSS JOIN cutoff
GROUP BY c.coin ORDER BY c.coin`,
		tenantID, snapshotMillis,
		tenantID, startMillis,
		tenantID, startMillis, snapshotMillis,
		startMillis, endMillis,
		startMillis, endMillis,
		startMillis, endMillis,
		startMillis, endMillis,
		startMillis, endMillis,
		endMillis,
		startMillis, endMillis,
		endMillis,
		tenantID,
		tenantID,
	)
	return rows, err
}

func QueryOptionPlatformAccountConservationSummaries(
	ctx context.Context, conn sqlx.SqlConn, tenantID, startMillis, endMillis, snapshotMillis int64,
) ([]*OptionPlatformAccountConservationSummary, error) {
	var rows []*OptionPlatformAccountConservationSummary
	err := conn.QueryRowsCtx(ctx, &rows, `WITH cutoff AS (
  SELECT COALESCE(MAX(id),0) max_platform_flow_id
  FROM t_asset_platform_flow
  WHERE tenant_id=? AND account_type IN ('FEE_REVENUE','INSURANCE_FUND','OPTION_BACKSTOP')
    AND create_times<?
), prior_ranked AS (
  SELECT f.*,ROW_NUMBER() OVER (
    PARTITION BY f.platform_account_id ORDER BY f.create_times DESC,f.id DESC
  ) rn
  FROM t_asset_platform_flow f JOIN cutoff c ON f.id<=c.max_platform_flow_id
  WHERE f.tenant_id=? AND f.account_type IN ('FEE_REVENUE','INSURANCE_FUND','OPTION_BACKSTOP')
    AND f.create_times<?
), relevant_flows AS (
  SELECT p.*,1 is_prior FROM prior_ranked p WHERE p.rn=1
  UNION ALL
  SELECT f.*,NULL rn,0 is_prior
  FROM t_asset_platform_flow f JOIN cutoff c ON f.id<=c.max_platform_flow_id
  WHERE f.tenant_id=? AND f.account_type IN ('FEE_REVENUE','INSURANCE_FUND','OPTION_BACKSTOP')
    AND f.create_times>=? AND f.create_times<?
), ordered_flows AS (
  SELECT f.*,LAG(after_available) OVER (
    PARTITION BY platform_account_id ORDER BY create_times,id
  ) prev_available
  FROM relevant_flows f
), checked_flows AS (
  SELECT f.*,(after_available-before_available) amount_delta,
    CASE WHEN is_prior=1 THEN 0 WHEN
      amount<=0 OR biz_type='' OR scene_type='' OR biz_no='' OR account_type='' OR coin=''
      OR (prev_available IS NULL AND before_available<>0)
      OR (prev_available IS NOT NULL AND before_available<>prev_available)
      OR (op_type=1 AND after_available-before_available<>amount)
      OR (op_type=2 AND after_available-before_available<>-amount)
      OR op_type NOT IN (1,2)
    THEN 1 ELSE 0 END integrity_error
  FROM ordered_flows f
), flow_aggregates AS (
  SELECT platform_account_id,
    COUNT(CASE WHEN is_prior=0 AND create_times>=? AND create_times<? THEN 1 END) flow_count,
    SUM(CASE WHEN is_prior=0 THEN integrity_error ELSE 0 END) integrity_errors,
    SUM(CASE WHEN is_prior=0 AND create_times>=? AND create_times<?
      AND scene_type='platform_manual_adjust' THEN amount_delta ELSE 0 END) manual_net,
    SUM(CASE WHEN is_prior=0 AND create_times>=? AND create_times<?
      AND scene_type<>'platform_manual_adjust' THEN amount_delta ELSE 0 END) option_net,
    SUM(CASE WHEN is_prior=0 AND create_times>=? THEN amount_delta ELSE 0 END) post_net
  FROM checked_flows GROUP BY platform_account_id
), first_day AS (
  SELECT platform_account_id,before_available,ROW_NUMBER() OVER (
    PARTITION BY platform_account_id ORDER BY create_times,id
  ) rn FROM checked_flows WHERE is_prior=0 AND create_times>=? AND create_times<?
), first_post AS (
  SELECT platform_account_id,before_available,ROW_NUMBER() OVER (
    PARTITION BY platform_account_id ORDER BY create_times,id
  ) rn FROM checked_flows WHERE is_prior=0 AND create_times>=?
), last_flow AS (
  SELECT platform_account_id,account_type,coin,after_available,ROW_NUMBER() OVER (
    PARTITION BY platform_account_id ORDER BY create_times DESC,id DESC
  ) rn FROM checked_flows
), account_keys AS (
  SELECT id platform_account_id FROM t_asset_platform_account
  WHERE tenant_id=? AND account_type IN ('FEE_REVENUE','INSURANCE_FUND','OPTION_BACKSTOP')
  UNION
  SELECT platform_account_id FROM checked_flows
), account_values AS (
  SELECT k.platform_account_id,COALESCE(a.account_type,lf.account_type) account_type,
    COALESCE(a.coin,lf.coin) coin,
    COALESCE(fd.before_available,fp.before_available,a.available_amount) opening_amount,
    COALESCE(g.option_net,0) option_net,COALESCE(g.manual_net,0) manual_net,
    COALESCE(g.flow_count,0) flow_count,
    COALESCE(g.integrity_errors,0)+CASE WHEN a.id IS NULL THEN 1 ELSE 0 END+
      CASE WHEN a.id IS NOT NULL AND a.frozen_amount<>0 THEN 1 ELSE 0 END+
      CASE WHEN lf.platform_account_id IS NOT NULL AND a.id IS NOT NULL AND (
        a.account_type<>lf.account_type OR a.coin<>lf.coin OR a.available_amount<>lf.after_available)
      THEN 1 ELSE 0 END integrity_errors,
    COALESCE(a.available_amount,0)-COALESCE(g.post_net,0) actual_closing
  FROM account_keys k
  LEFT JOIN t_asset_platform_account a ON a.tenant_id=? AND a.id=k.platform_account_id
  LEFT JOIN flow_aggregates g ON g.platform_account_id=k.platform_account_id
  LEFT JOIN first_day fd ON fd.platform_account_id=k.platform_account_id AND fd.rn=1
  LEFT JOIN first_post fp ON fp.platform_account_id=k.platform_account_id AND fp.rn=1
  LEFT JOIN last_flow lf ON lf.platform_account_id=k.platform_account_id AND lf.rn=1
), calculated AS (
  SELECT *,opening_amount+option_net+manual_net expected_closing FROM account_values
)
SELECT c.account_type,c.coin,COUNT(1) account_count,
  SUM(CASE WHEN c.integrity_errors>0 OR c.actual_closing<>c.expected_closing THEN 1 ELSE 0 END)
    mismatch_account_count,
  SUM(c.integrity_errors) integrity_error_count,SUM(c.flow_count) flow_count,
  MAX(cutoff.max_platform_flow_id) max_platform_flow_id,SUM(c.opening_amount) opening_amount,
  SUM(c.option_net) option_net,SUM(c.manual_net) manual_net,
  SUM(c.expected_closing) expected_closing,SUM(c.actual_closing) actual_closing,
  SUM(c.actual_closing-c.expected_closing) difference_amount
FROM calculated c CROSS JOIN cutoff
GROUP BY c.account_type,c.coin ORDER BY c.account_type,c.coin`,
		tenantID, snapshotMillis,
		tenantID, startMillis,
		tenantID, startMillis, snapshotMillis,
		startMillis, endMillis,
		startMillis, endMillis,
		startMillis, endMillis,
		endMillis,
		startMillis, endMillis,
		endMillis,
		tenantID,
		tenantID,
	)
	return rows, err
}

func InsertOptionReconciliationRun(
	ctx context.Context, conn sqlx.SqlConn, run *OptionReconciliationRun,
) error {
	_, err := InsertOptionReconciliationRunWithID(ctx, conn, run)
	return err
}

func InsertOptionReconciliationRunWithID(
	ctx context.Context, conn sqlx.SqlConn, run *OptionReconciliationRun,
) (int64, error) {
	result, err := conn.ExecCtx(ctx, `INSERT INTO t_option_reconciliation_run
(tenant_id,business_date,scope,attempt_no,status,snapshot_time,snapshot_ref,
 coin_count,account_count,mismatch_count,detail,completed_at,create_times,update_times)
VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		run.TenantID, run.BusinessDate, run.Scope, run.AttemptNo, run.Status,
		run.SnapshotTime, run.SnapshotRef, run.CoinCount, run.AccountCount,
		run.MismatchCount, run.Detail, run.CompletedAt, run.CompletedAt, run.CompletedAt)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func InsertOptionReconciliationRunDetail(
	ctx context.Context, conn sqlx.SqlConn, detail *OptionReconciliationRunDetail,
) error {
	_, err := conn.ExecCtx(ctx, `INSERT INTO t_option_reconciliation_run_detail
(run_id,tenant_id,business_date,scope,dimension_type,dimension_key,
 opening_amount,external_net,option_net,manual_net,expected_closing,actual_closing,
 difference_amount,flow_count,mismatch_count,status,evidence_ref,detail,create_times)
VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		detail.RunID, detail.TenantID, detail.BusinessDate, detail.Scope,
		detail.DimensionType, detail.DimensionKey, detail.OpeningAmount,
		detail.ExternalNet, detail.OptionNet, detail.ManualNet, detail.ExpectedClosing,
		detail.ActualClosing, detail.DifferenceAmount, detail.FlowCount,
		detail.MismatchCount, detail.Status, detail.EvidenceRef, detail.Detail,
		detail.CreateTimes)
	return err
}
