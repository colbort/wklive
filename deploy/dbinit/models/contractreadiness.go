package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"strings"
)

type ContractReadinessInput struct {
	SourceAuthorities       []string
	SourceMarkets           []string
	IndexSourceWeights      []string
	DeliverySourceWeights   []string
	CategoryCode            string
	Market                  string
	PriceSymbol             string
	PerpetualSymbol         string
	DeliverySymbol          string
	PerpetualPriceAuthority string
	PerpetualPriceMarket    string
	TenantID                int64
	SettlementCoin          string
	IndexAlgorithm          int
	IndexFormulaVersion     string
	IndexMaxDeviationBps    int64
	MarkFormulaVersion      string
	MarkMaxBasisBps         int64
	MarkCurrentWeight       string
	MarkPreviousWeight      string
	FundingFormulaVersion   string
	PriceFormulaIntervalMs  int64
	DeliveryAlgorithm       int
	DeliveryFormulaVersion  string
	DeliveryMaxLookbackMs   int64
	DeliveryMaxDeviationBps int64
}

type ContractReadinessResult struct {
	Detailed                    bool
	ActiveSourceAuthorityCount  int64
	DistinctSourceProviderCount int64
	PublicRestSourceCount       int64
	PriceEngineAuthorityCount   int64
	IndexFormulaCount           int64
	MarkFormulaCount            int64
	FundingFormulaCount         int64
	DeliveryFormulaCount        int64
	FreshSourceCount            int64
	FreshOutputKindCount        int64
	InsuranceFundCount          int64
	FeeRevenueCount             int64
	PerpetualContractCount      int64
	DeliveryContractCount       int64
	InsuranceConfigCount        int64
	PendingOutboxCount          int64
	ProcessingOutboxCount       int64
	FailedOutboxCount           int64
	ManualOutboxCount           int64
	OpenOutboxCount             int64
	UnhealthyOutboxCount        int64
	OpenReconciliationCount     int64
	OpenSettlementCount         int64
}

type ContractReadinessModel interface {
	Inspect(ctx context.Context, input ContractReadinessInput, detailed bool) (ContractReadinessResult, error)
}

type defaultContractReadinessModel struct {
	db *sql.DB
}

func NewContractReadinessModel(db *sql.DB) ContractReadinessModel {
	return &defaultContractReadinessModel{db: db}
}

func (m *defaultContractReadinessModel) Inspect(
	ctx context.Context,
	input ContractReadinessInput,
	detailed bool,
) (ContractReadinessResult, error) {
	result := ContractReadinessResult{Detailed: detailed}
	if detailed {
		if err := validateContractReadinessInput(input); err != nil {
			return result, err
		}
		if err := m.inspectAuthorities(ctx, input, &result); err != nil {
			return result, fmt.Errorf("inspect authorities: %w", err)
		}
		if err := m.inspectFormulas(ctx, input, &result); err != nil {
			return result, fmt.Errorf("inspect formulas: %w", err)
		}
		if err := m.inspectLivePrices(ctx, input, &result); err != nil {
			return result, fmt.Errorf("inspect live prices: %w", err)
		}
		if err := m.inspectPlatformAccounts(ctx, input, &result); err != nil {
			return result, fmt.Errorf("inspect platform accounts: %w", err)
		}
		if err := m.inspectContracts(ctx, input, &result); err != nil {
			return result, fmt.Errorf("inspect contracts: %w", err)
		}
		if err := m.inspectInsuranceConfig(ctx, input, &result); err != nil {
			return result, fmt.Errorf("inspect insurance config: %w", err)
		}
	}
	if err := m.inspectOpenFacts(ctx, &result); err != nil {
		return result, fmt.Errorf("inspect open facts: %w", err)
	}
	return result, nil
}

func validateContractReadinessInput(input ContractReadinessInput) error {
	if len(input.SourceAuthorities) < 3 {
		return errors.New("at least three source authorities are required")
	}
	if len(input.SourceMarkets) != len(input.SourceAuthorities) {
		return errors.New("source markets must match source authorities")
	}
	if len(input.IndexSourceWeights) != len(input.SourceAuthorities) {
		return errors.New("index source weights must match source authorities")
	}
	if len(input.DeliverySourceWeights) != len(input.SourceAuthorities) {
		return errors.New("delivery source weights must match source authorities")
	}
	weights := append([]string(nil), input.IndexSourceWeights...)
	weights = append(weights, input.DeliverySourceWeights...)
	weights = append(weights, input.MarkCurrentWeight, input.MarkPreviousWeight)
	for _, weight := range weights {
		value, ok := new(big.Rat).SetString(weight)
		if !ok || value.Sign() <= 0 {
			return errors.New("price formula weights must be positive decimals")
		}
	}
	if input.CategoryCode == "" || input.Market == "" || input.PriceSymbol == "" ||
		input.PerpetualSymbol == "" || input.DeliverySymbol == "" ||
		input.PerpetualPriceAuthority == "" || input.PerpetualPriceMarket == "" ||
		input.TenantID <= 0 || input.SettlementCoin == "" ||
		input.IndexFormulaVersion == "" || input.IndexMaxDeviationBps <= 0 ||
		input.MarkFormulaVersion == "" || input.MarkMaxBasisBps <= 0 ||
		input.FundingFormulaVersion == "" || input.PriceFormulaIntervalMs <= 0 ||
		input.DeliveryFormulaVersion == "" || input.DeliveryMaxLookbackMs <= 0 ||
		input.DeliveryMaxDeviationBps <= 0 {
		return errors.New("production price, formula, contract, and tenant dimensions are required")
	}
	if input.DeliveryMaxLookbackMs%1000 != 0 {
		return errors.New("delivery max lookback must use whole seconds")
	}
	if input.IndexAlgorithm != 1 && input.IndexAlgorithm != 2 ||
		input.DeliveryAlgorithm != 1 && input.DeliveryAlgorithm != 2 {
		return errors.New("index and delivery algorithms must be weighted mean or median")
	}
	return nil
}

func (m *defaultContractReadinessModel) inspectAuthorities(
	ctx context.Context,
	input ContractReadinessInput,
	result *ContractReadinessResult,
) error {
	query := fmt.Sprintf(`
SELECT
  (SELECT COUNT(*)
   FROM t_market_authority_registry
   WHERE status=1
     AND authority IN (%s)
     AND JSON_CONTAINS(allowed_kinds, JSON_QUOTE('FINAL_QUOTE'))),
  (SELECT COUNT(DISTINCT provider_code)
   FROM t_market_authority_registry
   WHERE status=1
     AND authority IN (%s)
     AND JSON_CONTAINS(allowed_kinds, JSON_QUOTE('FINAL_QUOTE'))),
  (SELECT COUNT(*)
   FROM t_market_authority_registry
   WHERE status=1
     AND authority IN (%s)
     AND producer_type='PUBLIC_REST'
     AND JSON_CONTAINS(allowed_kinds, JSON_QUOTE('FINAL_QUOTE'))),
  (SELECT COUNT(*)
   FROM t_market_authority_registry
   WHERE status=1
     AND authority='price-engine'
     AND JSON_CONTAINS(allowed_kinds, JSON_QUOTE('INDEX'))
     AND JSON_CONTAINS(allowed_kinds, JSON_QUOTE('MARK'))
     AND JSON_CONTAINS(allowed_kinds, JSON_QUOTE('FUNDING'))
     AND JSON_CONTAINS(allowed_kinds, JSON_QUOTE('DELIVERY')))`,
		placeholders(len(input.SourceAuthorities)),
		placeholders(len(input.SourceAuthorities)),
		placeholders(len(input.SourceAuthorities)),
	)
	args := appendStringArgs(nil, input.SourceAuthorities)
	args = appendStringArgs(args, input.SourceAuthorities)
	args = appendStringArgs(args, input.SourceAuthorities)
	return m.db.QueryRowContext(ctx, query, args...).Scan(
		&result.ActiveSourceAuthorityCount,
		&result.DistinctSourceProviderCount,
		&result.PublicRestSourceCount,
		&result.PriceEngineAuthorityCount,
	)
}

func (m *defaultContractReadinessModel) inspectFormulas(
	ctx context.Context,
	input ContractReadinessInput,
	result *ContractReadinessResult,
) error {
	indexSourcePredicate := sourceIdentityWeightPredicate(len(input.SourceAuthorities))
	deliverySourcePredicate := sourceIdentityWeightPredicate(len(input.SourceAuthorities))
	query := fmt.Sprintf(`
SELECT
  COALESCE(SUM(f.snapshot_kind='INDEX'
    AND f.algorithm=?
    AND f.formula_version=?
    AND f.max_lookback_ms=?
    AND f.max_deviation_bps=?
    AND f.min_input_count=?
    AND f.interval_ms=?
    AND JSON_LENGTH(f.components)=?
    AND (SELECT COUNT(DISTINCT j.authority)
         FROM JSON_TABLE(f.components, '$[*]' COLUMNS(
           kind VARCHAR(32) PATH '$.kind',
           authority VARCHAR(32) PATH '$.authority',
           category_code VARCHAR(64) PATH '$.category_code',
           market VARCHAR(32) PATH '$.market',
           symbol VARCHAR(64) PATH '$.symbol',
           weight DECIMAL(36,18) PATH '$.weight'
         )) AS j
         WHERE j.kind='FINAL_QUOTE'
           AND (%s))=?),0),
  COALESCE(SUM(f.snapshot_kind='MARK'
    AND f.algorithm=4
    AND f.formula_version=?
    AND f.max_lookback_ms=?
    AND f.max_deviation_bps=?
    AND f.min_input_count=3
    AND f.interval_ms=?
    AND JSON_LENGTH(f.components)=3
    AND (SELECT COUNT(DISTINCT j.kind)
         FROM JSON_TABLE(f.components, '$[*]' COLUMNS(
           kind VARCHAR(32) PATH '$.kind',
           authority VARCHAR(32) PATH '$.authority',
           category_code VARCHAR(64) PATH '$.category_code',
           market VARCHAR(32) PATH '$.market',
           symbol VARCHAR(64) PATH '$.symbol',
           weight DECIMAL(36,18) PATH '$.weight'
         )) AS j
         WHERE (j.kind='INDEX'
             AND j.authority='price-engine'
             AND j.category_code=f.category_code
             AND j.market=f.market
             AND j.symbol=f.symbol
             AND j.weight=CAST(? AS DECIMAL(36,18)))
            OR (j.kind='FINAL_QUOTE'
             AND j.authority=?
             AND j.category_code=f.category_code
             AND j.market=?
             AND j.symbol=f.symbol
             AND j.weight=CAST(? AS DECIMAL(36,18)))
            OR (j.kind='MARK'
             AND j.authority='price-engine'
             AND j.category_code=f.category_code
             AND j.market=f.market
             AND j.symbol=f.symbol
             AND j.weight=CAST(? AS DECIMAL(36,18))))=3),0),
  COALESCE(SUM(f.snapshot_kind='FUNDING'
    AND f.algorithm=3
    AND f.formula_version=?
    AND f.max_lookback_ms=?
    AND f.min_input_count=2
    AND f.interval_ms=?
    AND JSON_LENGTH(f.components)=2
    AND (SELECT COUNT(DISTINCT j.kind)
         FROM JSON_TABLE(f.components, '$[*]' COLUMNS(
           kind VARCHAR(32) PATH '$.kind',
           authority VARCHAR(32) PATH '$.authority',
           category_code VARCHAR(64) PATH '$.category_code',
           market VARCHAR(32) PATH '$.market',
           symbol VARCHAR(64) PATH '$.symbol'
         )) AS j
         WHERE j.kind IN ('MARK','INDEX')
           AND j.authority='price-engine'
           AND j.category_code=f.category_code
           AND j.market=f.market
           AND j.symbol=f.symbol)=2),0),
  COALESCE(SUM(f.snapshot_kind='DELIVERY'
    AND f.algorithm=?
    AND f.formula_version=?
    AND f.max_lookback_ms=?
    AND f.max_deviation_bps=?
    AND f.min_input_count=?
    AND f.interval_ms=?
    AND JSON_LENGTH(f.components)=?
    AND (SELECT COUNT(DISTINCT j.authority)
         FROM JSON_TABLE(f.components, '$[*]' COLUMNS(
           kind VARCHAR(32) PATH '$.kind',
           authority VARCHAR(32) PATH '$.authority',
           category_code VARCHAR(64) PATH '$.category_code',
           market VARCHAR(32) PATH '$.market',
           symbol VARCHAR(64) PATH '$.symbol',
           weight DECIMAL(36,18) PATH '$.weight'
         )) AS j
         WHERE j.kind='FINAL_QUOTE'
           AND (%s))=?),0)
FROM t_market_price_formula AS f
WHERE f.status=1
  AND f.category_code=?
  AND f.market=?
  AND f.symbol=?`, indexSourcePredicate, deliverySourcePredicate)
	args := make([]any, 0, len(input.SourceAuthorities)*6+32)
	args = append(args,
		input.IndexAlgorithm,
		input.IndexFormulaVersion,
		input.DeliveryMaxLookbackMs,
		input.IndexMaxDeviationBps,
		len(input.SourceAuthorities),
		input.PriceFormulaIntervalMs,
		len(input.SourceAuthorities),
	)
	args = appendSourceIdentityWeightArgs(
		args,
		input.SourceAuthorities,
		input.SourceMarkets,
		input.IndexSourceWeights,
	)
	args = append(args,
		len(input.SourceAuthorities),
		input.MarkFormulaVersion,
		input.DeliveryMaxLookbackMs,
		input.MarkMaxBasisBps,
		input.PriceFormulaIntervalMs,
		input.MarkCurrentWeight,
		input.PerpetualPriceAuthority,
		input.PerpetualPriceMarket,
		input.MarkCurrentWeight,
		input.MarkPreviousWeight,
		input.FundingFormulaVersion,
		input.DeliveryMaxLookbackMs,
		input.PriceFormulaIntervalMs,
	)
	args = append(args,
		input.DeliveryAlgorithm,
		input.DeliveryFormulaVersion,
		input.DeliveryMaxLookbackMs,
		input.DeliveryMaxDeviationBps,
		len(input.SourceAuthorities),
		input.PriceFormulaIntervalMs,
		len(input.SourceAuthorities),
	)
	args = appendSourceIdentityWeightArgs(
		args,
		input.SourceAuthorities,
		input.SourceMarkets,
		input.DeliverySourceWeights,
	)
	args = append(args, len(input.SourceAuthorities))
	args = append(args, input.CategoryCode, input.Market, input.PriceSymbol)
	return m.db.QueryRowContext(ctx, query, args...).Scan(
		&result.IndexFormulaCount,
		&result.MarkFormulaCount,
		&result.FundingFormulaCount,
		&result.DeliveryFormulaCount,
	)
}

func (m *defaultContractReadinessModel) inspectLivePrices(
	ctx context.Context,
	input ContractReadinessInput,
	result *ContractReadinessResult,
) error {
	query := fmt.Sprintf(`
SELECT
  (SELECT COUNT(DISTINCT j.authority)
   FROM t_market_price_formula AS f
   JOIN JSON_TABLE(f.components, '$[*]' COLUMNS(
     kind VARCHAR(32) PATH '$.kind',
     authority VARCHAR(32) PATH '$.authority',
     category_code VARCHAR(64) PATH '$.category_code',
     market VARCHAR(32) PATH '$.market',
     symbol VARCHAR(64) PATH '$.symbol'
   )) AS j
   JOIN t_market_authoritative_snapshot AS s
     ON s.authority=j.authority
    AND s.snapshot_kind=j.kind
    AND s.category_code=j.category_code
    AND s.market=j.market
    AND s.symbol=j.symbol
    AND s.source_timestamp>=CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3))*1000 AS UNSIGNED)-f.max_lookback_ms
   WHERE f.status=1
     AND f.snapshot_kind='INDEX'
     AND f.category_code=?
     AND f.market=?
     AND f.symbol=?
     AND j.kind='FINAL_QUOTE'
     AND j.authority IN (%s)),
  (SELECT COUNT(DISTINCT s.snapshot_kind)
   FROM t_market_authoritative_snapshot AS s
   WHERE s.authority='price-engine'
     AND s.snapshot_kind IN ('INDEX','MARK','FUNDING','DELIVERY')
     AND s.category_code=?
     AND s.market=?
     AND s.symbol=?
     AND s.source_timestamp>=CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3))*1000 AS UNSIGNED)-?)`,
		placeholders(len(input.SourceAuthorities)),
	)
	args := []any{input.CategoryCode, input.Market, input.PriceSymbol}
	args = appendStringArgs(args, input.SourceAuthorities)
	args = append(args,
		input.CategoryCode,
		input.Market,
		input.PriceSymbol,
		input.DeliveryMaxLookbackMs,
	)
	return m.db.QueryRowContext(ctx, query, args...).Scan(
		&result.FreshSourceCount,
		&result.FreshOutputKindCount,
	)
}

func (m *defaultContractReadinessModel) inspectPlatformAccounts(
	ctx context.Context,
	input ContractReadinessInput,
	result *ContractReadinessResult,
) error {
	const query = `
SELECT
  COALESCE(SUM(account_type='INSURANCE_FUND' AND status=1 AND available_amount>0),0),
  COALESCE(SUM(account_type='FEE_REVENUE' AND status=1),0)
FROM t_asset_platform_account
WHERE tenant_id=?
  AND coin=?`
	return m.db.QueryRowContext(ctx, query, input.TenantID, input.SettlementCoin).Scan(
		&result.InsuranceFundCount,
		&result.FeeRevenueCount,
	)
}

func (m *defaultContractReadinessModel) inspectContracts(
	ctx context.Context,
	input ContractReadinessInput,
	result *ContractReadinessResult,
) error {
	const query = `
SELECT
  COALESCE(SUM(s.symbol=?
    AND s.contract_type=1
    AND s.status=1
    AND c.funding_interval_minutes>0
    AND c.funding_rate_source<>''
    AND c.index_symbol<>''
    AND c.mark_price_source<>''
    AND c.delivery_time=0),0),
  COALESCE(SUM(s.symbol=?
    AND s.contract_type=2
    AND s.status=1
    AND c.delivery_time>CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3))*1000 AS UNSIGNED)
    AND c.settlement_window_seconds*1000=?
    AND c.settlement_price_source<>''
    AND c.settlement_price_algorithm=?),0)
FROM t_trade_symbol AS s
JOIN t_trade_symbol_contract AS c
  ON c.tenant_id=s.tenant_id AND c.symbol_id=s.id
WHERE s.tenant_id=?
  AND s.product_type=2
  AND s.settle_asset=?`
	return m.db.QueryRowContext(
		ctx,
		query,
		input.PerpetualSymbol,
		input.DeliverySymbol,
		input.DeliveryMaxLookbackMs,
		input.DeliveryFormulaVersion,
		input.TenantID,
		input.SettlementCoin,
	).Scan(
		&result.PerpetualContractCount,
		&result.DeliveryContractCount,
	)
}

func (m *defaultContractReadinessModel) inspectInsuranceConfig(
	ctx context.Context,
	input ContractReadinessInput,
	result *ContractReadinessResult,
) error {
	const query = `
SELECT COUNT(*)
FROM t_contract_insurance_fund_account
WHERE tenant_id=?
  AND symbol_id=0
  AND settle_asset=?
  AND status=1`
	return m.db.QueryRowContext(ctx, query, input.TenantID, input.SettlementCoin).
		Scan(&result.InsuranceConfigCount)
}

func (m *defaultContractReadinessModel) inspectOpenFacts(
	ctx context.Context,
	result *ContractReadinessResult,
) error {
	const query = `
SELECT
  COALESCE(SUM(status=1),0),
  COALESCE(SUM(status=2),0),
  COALESCE(SUM(status=4),0),
  COALESCE(SUM(status=5),0),
  COALESCE(MIN(CASE WHEN status IN (1,2,4,5) THEN create_times END),0),
  CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3))*1000 AS UNSIGNED),
  (SELECT COUNT(*) FROM t_contract_reconciliation_issue WHERE status=1),
  (SELECT COUNT(*) FROM t_trade_settlement_instruction WHERE status IN (1,2,4,5))
FROM t_market_snapshot_outbox`
	var oldestOpenAt, serverNow int64
	err := m.db.QueryRowContext(ctx, query).Scan(
		&result.PendingOutboxCount,
		&result.ProcessingOutboxCount,
		&result.FailedOutboxCount,
		&result.ManualOutboxCount,
		&oldestOpenAt,
		&serverNow,
		&result.OpenReconciliationCount,
		&result.OpenSettlementCount,
	)
	if err != nil {
		return err
	}
	result.OpenOutboxCount = result.PendingOutboxCount + result.ProcessingOutboxCount +
		result.FailedOutboxCount + result.ManualOutboxCount
	if result.FailedOutboxCount > 0 || result.ManualOutboxCount > 0 ||
		result.PendingOutboxCount+result.ProcessingOutboxCount > 0 && serverNow-oldestOpenAt > 60_000 {
		result.UnhealthyOutboxCount = result.OpenOutboxCount
	}
	return nil
}

func placeholders(count int) string {
	return strings.TrimSuffix(strings.Repeat("?,", count), ",")
}

func appendStringArgs(args []any, values []string) []any {
	for _, value := range values {
		args = append(args, value)
	}
	return args
}

func sourceIdentityWeightPredicate(count int) string {
	return strings.TrimSuffix(
		strings.Repeat(
			"(j.authority=? AND j.market=? AND j.category_code=f.category_code "+
				"AND j.symbol=f.symbol AND j.weight=CAST(? AS DECIMAL(36,18))) OR ",
			count,
		),
		" OR ",
	)
}

func appendSourceIdentityWeightArgs(args []any, authorities, markets, weights []string) []any {
	for index, authority := range authorities {
		args = append(args, authority, markets[index], weights[index])
	}
	return args
}
