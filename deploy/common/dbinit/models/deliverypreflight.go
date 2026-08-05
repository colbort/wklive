package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type DeliveryPreflightInput struct {
	TenantID               int64
	Symbol                 string
	SettlementAsset        string
	CategoryCode           string
	Market                 string
	PriceSymbol            string
	FormulaVersion         string
	FormulaAlgorithm       int
	FormulaMaxLookbackMs   int64
	FormulaMaxDeviationBps int64
	FormulaMinInputCount   int64
	FormulaIntervalMs      int64
}

type DeliveryPreflightResult struct {
	ProductCount             int64
	ConfiguredProductCount   int64
	SafeDisabledProductCount int64
	SymbolID                 int64
	ProductStatus            int64
	DeliveryTimeMs           int64
	OpenCutoffTimeMs         int64
	MatchingStopTimeMs       int64
	ServerTimeMs             int64

	IsolatedLeverageConfigCount  int64
	IsolatedLeverageDefaultCount int64
	CrossLeverageConfigCount     int64
	EnabledRiskTierCount         int64
	ValidRiskTierCount           int64
	BaseRiskTierCount            int64
	RiskCoverageEndCount         int64

	FormulaCount           int64
	ConformingFormulaCount int64
	FreshSnapshotCount     int64
	LatestSnapshotTimeMs   int64
	PriceServerTimeMs      int64

	OrderCount                 int64
	FillCount                  int64
	PositionCount              int64
	PositionHistoryCount       int64
	ReservationCount           int64
	SettlementInstructionCount int64
	DeliveryBatchCount         int64
	DeliverySettlementCount    int64
}

func (r DeliveryPreflightResult) HistoricalFactCount() int64 {
	return r.OrderCount +
		r.FillCount +
		r.PositionCount +
		r.PositionHistoryCount +
		r.ReservationCount +
		r.SettlementInstructionCount +
		r.DeliveryBatchCount +
		r.DeliverySettlementCount
}

// TechnicalReady reports only whether the disabled delivery product is in a
// safe, technically complete pre-enable state. It is not production approval.
func (r DeliveryPreflightResult) TechnicalReady() bool {
	return r.ProductCount == 1 &&
		r.ConfiguredProductCount == 1 &&
		r.SafeDisabledProductCount == 1 &&
		r.SymbolID > 0 &&
		r.ProductStatus == 2 &&
		r.OpenCutoffTimeMs > r.ServerTimeMs &&
		r.OpenCutoffTimeMs < r.MatchingStopTimeMs &&
		r.MatchingStopTimeMs <= r.DeliveryTimeMs &&
		r.IsolatedLeverageConfigCount == 1 &&
		r.IsolatedLeverageDefaultCount == 1 &&
		r.CrossLeverageConfigCount == 0 &&
		r.EnabledRiskTierCount > 0 &&
		r.ValidRiskTierCount == r.EnabledRiskTierCount &&
		r.BaseRiskTierCount == 1 &&
		r.RiskCoverageEndCount == 1 &&
		r.ConformingFormulaCount == 1 &&
		r.FreshSnapshotCount > 0 &&
		r.LatestSnapshotTimeMs > 0 &&
		r.LatestSnapshotTimeMs <= r.PriceServerTimeMs &&
		r.HistoricalFactCount() == 0
}

type DeliveryPreflightModel interface {
	Inspect(ctx context.Context, input DeliveryPreflightInput) (DeliveryPreflightResult, error)
}

type defaultDeliveryPreflightModel struct {
	db *sql.DB
}

func NewDeliveryPreflightModel(db *sql.DB) DeliveryPreflightModel {
	return &defaultDeliveryPreflightModel{db: db}
}

func (m *defaultDeliveryPreflightModel) Inspect(
	ctx context.Context,
	input DeliveryPreflightInput,
) (DeliveryPreflightResult, error) {
	var result DeliveryPreflightResult
	if err := validateDeliveryPreflightInput(input); err != nil {
		return result, err
	}
	if err := m.inspectDeliveryProduct(ctx, input, &result); err != nil {
		return result, fmt.Errorf("inspect delivery product: %w", err)
	}
	if result.SymbolID == 0 {
		return result, nil
	}
	if err := m.inspectDeliveryRiskProfile(ctx, input, result.SymbolID, &result); err != nil {
		return result, fmt.Errorf("inspect delivery risk profile: %w", err)
	}
	if err := m.inspectDeliveryPrice(ctx, input, &result); err != nil {
		return result, fmt.Errorf("inspect delivery price: %w", err)
	}
	if err := m.inspectDeliveryFacts(ctx, input.TenantID, result.SymbolID, &result); err != nil {
		return result, fmt.Errorf("inspect delivery facts: %w", err)
	}
	return result, nil
}

func validateDeliveryPreflightInput(input DeliveryPreflightInput) error {
	if input.TenantID <= 0 ||
		strings.TrimSpace(input.Symbol) == "" ||
		strings.TrimSpace(input.SettlementAsset) == "" ||
		strings.TrimSpace(input.CategoryCode) == "" ||
		strings.TrimSpace(input.Market) == "" ||
		strings.TrimSpace(input.PriceSymbol) == "" ||
		strings.TrimSpace(input.FormulaVersion) == "" {
		return errors.New("delivery product, tenant, price, and formula dimensions are required")
	}
	if input.FormulaAlgorithm != 1 && input.FormulaAlgorithm != 2 {
		return errors.New("delivery formula algorithm must be weighted mean or median")
	}
	if input.FormulaMaxLookbackMs <= 0 || input.FormulaMaxLookbackMs%1000 != 0 ||
		input.FormulaMaxDeviationBps <= 0 ||
		input.FormulaMinInputCount < 3 ||
		input.FormulaIntervalMs <= 0 {
		return errors.New("delivery formula window, deviation, source count, and interval are invalid")
	}
	return nil
}

func (m *defaultDeliveryPreflightModel) inspectDeliveryProduct(
	ctx context.Context,
	input DeliveryPreflightInput,
	result *DeliveryPreflightResult,
) error {
	const query = `
SELECT
  COUNT(s.id),
  COALESCE(SUM(c.id IS NOT NULL
    AND s.product_type=2
    AND s.contract_type=2
    AND s.contract_value_type IN (1,2)
    AND s.settle_asset=?
    AND s.margin_asset=?
    AND s.price_tick>0
    AND s.qty_step>0
    AND s.min_price>=0
    AND (s.max_price=0 OR s.max_price>=s.min_price)
    AND s.min_qty>=0
    AND (s.max_qty=0 OR s.max_qty>=s.min_qty)
    AND s.min_notional>=0
    AND (s.max_notional=0 OR s.max_notional>=s.min_notional)
    AND s.listing_time>0
    AND s.listing_time<=CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3))*1000 AS UNSIGNED)
    AND s.trading_start_time>=s.listing_time
    AND s.trading_start_time<=CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3))*1000 AS UNSIGNED)
    AND s.trading_end_time=c.matching_stop_time
    AND c.contract_size>0
    AND c.multiplier>0
    AND c.maintenance_margin_rate>0
    AND c.initial_margin_rate>=c.maintenance_margin_rate
    AND c.maker_fee_rate>=0
    AND c.taker_fee_rate>=0
    AND c.delivery_fee_rate>=0
    AND c.liquidation_fee_rate>=0
    AND c.settlement_price_source=?
    AND c.delivery_time>CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3))*1000 AS UNSIGNED)
    AND c.open_cutoff_time>CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3))*1000 AS UNSIGNED)
    AND c.open_cutoff_time<c.matching_stop_time
    AND c.matching_stop_time<=c.delivery_time
    AND c.settlement_window_seconds=?
    AND c.settlement_price_algorithm=?
    AND c.support_cross=0
    AND c.support_isolated=1),0),
  COALESCE(SUM(s.status=2
    AND c.open_long_enabled=2
    AND c.open_short_enabled=2
    AND c.close_long_enabled=1
    AND c.close_short_enabled=1),0),
  COALESCE(MAX(s.id),0),
  COALESCE(MAX(s.status),0),
  COALESCE(MAX(c.delivery_time),0),
  COALESCE(MAX(c.open_cutoff_time),0),
  COALESCE(MAX(c.matching_stop_time),0),
  CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3))*1000 AS UNSIGNED)
FROM t_trade_symbol AS s
LEFT JOIN t_trade_symbol_contract AS c
  ON c.tenant_id=s.tenant_id AND c.symbol_id=s.id
WHERE s.tenant_id=?
  AND s.symbol=?`
	settlementSource := input.CategoryCode + ":" + input.Market + ":" + input.PriceSymbol
	return m.db.QueryRowContext(
		ctx,
		query,
		input.SettlementAsset,
		input.SettlementAsset,
		settlementSource,
		input.FormulaMaxLookbackMs/1000,
		input.FormulaVersion,
		input.TenantID,
		input.Symbol,
	).Scan(
		&result.ProductCount,
		&result.ConfiguredProductCount,
		&result.SafeDisabledProductCount,
		&result.SymbolID,
		&result.ProductStatus,
		&result.DeliveryTimeMs,
		&result.OpenCutoffTimeMs,
		&result.MatchingStopTimeMs,
		&result.ServerTimeMs,
	)
}

func (m *defaultDeliveryPreflightModel) inspectDeliveryRiskProfile(
	ctx context.Context,
	input DeliveryPreflightInput,
	symbolID int64,
	result *DeliveryPreflightResult,
) error {
	const query = `
SELECT
  (SELECT COUNT(*)
   FROM t_trade_symbol_leverage_config AS l
   WHERE l.tenant_id=?
     AND l.symbol_id=?
     AND l.margin_mode=2
     AND l.enabled=1
     AND JSON_LENGTH(l.leverage_values)>0),
  (SELECT COUNT(*)
   FROM t_trade_symbol_leverage_default AS d
   JOIN t_trade_symbol_leverage_config AS l
     ON l.tenant_id=d.tenant_id
    AND l.symbol_id=d.symbol_id
    AND l.margin_mode=d.margin_mode
    AND l.enabled=1
   WHERE d.tenant_id=?
     AND d.symbol_id=?
     AND d.margin_mode=2
     AND JSON_CONTAINS(l.leverage_values, CAST(d.leverage AS JSON))),
  (SELECT COUNT(*)
   FROM t_trade_symbol_leverage_config AS l
   WHERE l.tenant_id=?
     AND l.symbol_id=?
     AND l.margin_mode=1
     AND l.enabled=1),
  (SELECT COUNT(*)
   FROM t_contract_risk_limit_tier AS r
   WHERE r.tenant_id=?
     AND r.symbol_id=?
     AND r.enabled=1),
  (SELECT COUNT(*)
   FROM t_contract_risk_limit_tier AS r
   WHERE r.tenant_id=?
     AND r.symbol_id=?
     AND r.enabled=1
     AND r.tier_no>0
     AND r.notional_floor>=0
     AND (r.notional_cap=0 OR r.notional_cap>r.notional_floor)
     AND r.max_leverage>0
     AND r.initial_margin_rate>=r.maintenance_margin_rate
     AND r.maintenance_margin_rate>0
     AND r.maintenance_amount>=0
     AND ((r.tier_no=1 AND r.notional_floor=0)
       OR EXISTS (
         SELECT 1
         FROM t_contract_risk_limit_tier AS p
         WHERE p.tenant_id=r.tenant_id
           AND p.symbol_id=r.symbol_id
           AND p.enabled=1
           AND p.tier_no=r.tier_no-1
           AND p.notional_cap=r.notional_floor
       ))
     AND EXISTS (
       SELECT 1
       FROM t_trade_symbol_leverage_config AS l
       WHERE l.tenant_id=r.tenant_id
         AND l.symbol_id=r.symbol_id
         AND l.margin_mode=2
         AND l.enabled=1
         AND JSON_CONTAINS(l.leverage_values, CAST(r.max_leverage AS JSON))
     )),
  (SELECT COUNT(*)
   FROM t_contract_risk_limit_tier AS r
   WHERE r.tenant_id=?
     AND r.symbol_id=?
     AND r.enabled=1
     AND r.tier_no=1
     AND r.notional_floor=0),
  (SELECT COUNT(*)
   FROM t_contract_risk_limit_tier AS r
   JOIN t_trade_symbol AS s
     ON s.tenant_id=r.tenant_id AND s.id=r.symbol_id
   WHERE r.tenant_id=?
     AND r.symbol_id=?
     AND r.enabled=1
     AND NOT EXISTS (
       SELECT 1
       FROM t_contract_risk_limit_tier AS n
       WHERE n.tenant_id=r.tenant_id
         AND n.symbol_id=r.symbol_id
         AND n.enabled=1
         AND n.tier_no=r.tier_no+1
     )
     AND (r.notional_cap=0 OR r.notional_cap=s.max_notional))`
	args := make([]any, 0, 14)
	for range 7 {
		args = append(args, input.TenantID, symbolID)
	}
	return m.db.QueryRowContext(ctx, query, args...).Scan(
		&result.IsolatedLeverageConfigCount,
		&result.IsolatedLeverageDefaultCount,
		&result.CrossLeverageConfigCount,
		&result.EnabledRiskTierCount,
		&result.ValidRiskTierCount,
		&result.BaseRiskTierCount,
		&result.RiskCoverageEndCount,
	)
}

func (m *defaultDeliveryPreflightModel) inspectDeliveryPrice(
	ctx context.Context,
	input DeliveryPreflightInput,
	result *DeliveryPreflightResult,
) error {
	const query = `
SELECT
  COUNT(*),
  COALESCE(SUM(f.status=1
    AND f.algorithm=?
    AND f.formula_version=?
    AND f.max_lookback_ms=?
    AND f.max_deviation_bps=?
    AND f.min_input_count=?
    AND f.interval_ms=?
    AND JSON_LENGTH(f.components)=?
    AND (SELECT COUNT(*)
         FROM JSON_TABLE(f.components, '$[*]' COLUMNS(
           kind VARCHAR(32) PATH '$.kind',
           authority VARCHAR(32) PATH '$.authority',
           category_code VARCHAR(64) PATH '$.category_code',
           market VARCHAR(32) PATH '$.market',
           symbol VARCHAR(64) PATH '$.symbol',
           weight DECIMAL(36,18) PATH '$.weight'
         )) AS j
         WHERE j.kind='FINAL_QUOTE'
           AND j.authority<>''
           AND j.category_code=f.category_code
           AND j.market<>''
           AND j.symbol=f.symbol
           AND j.weight>0)=?
    AND (SELECT COUNT(DISTINCT j.authority)
         FROM JSON_TABLE(f.components, '$[*]' COLUMNS(
           authority VARCHAR(32) PATH '$.authority'
         )) AS j)=?),0),
  (SELECT COUNT(*)
   FROM t_itick_authoritative_snapshot AS s
   WHERE s.authority='price-engine'
     AND s.snapshot_kind='DELIVERY'
     AND s.category_code=?
     AND s.market=?
     AND s.symbol=?
     AND s.formula_version=?
     AND s.price>0
     AND s.source_timestamp>=CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3))*1000 AS UNSIGNED)-?),
  COALESCE((SELECT MAX(s.source_timestamp)
   FROM t_itick_authoritative_snapshot AS s
   WHERE s.authority='price-engine'
     AND s.snapshot_kind='DELIVERY'
     AND s.category_code=?
     AND s.market=?
     AND s.symbol=?
     AND s.formula_version=?),0),
  CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3))*1000 AS UNSIGNED)
FROM t_itick_price_formula AS f
WHERE f.authority='price-engine'
  AND f.snapshot_kind='DELIVERY'
  AND f.category_code=?
  AND f.market=?
  AND f.symbol=?`
	return m.db.QueryRowContext(
		ctx,
		query,
		input.FormulaAlgorithm,
		input.FormulaVersion,
		input.FormulaMaxLookbackMs,
		input.FormulaMaxDeviationBps,
		input.FormulaMinInputCount,
		input.FormulaIntervalMs,
		input.FormulaMinInputCount,
		input.FormulaMinInputCount,
		input.FormulaMinInputCount,
		input.CategoryCode,
		input.Market,
		input.PriceSymbol,
		input.FormulaVersion,
		input.FormulaMaxLookbackMs,
		input.CategoryCode,
		input.Market,
		input.PriceSymbol,
		input.FormulaVersion,
		input.CategoryCode,
		input.Market,
		input.PriceSymbol,
	).Scan(
		&result.FormulaCount,
		&result.ConformingFormulaCount,
		&result.FreshSnapshotCount,
		&result.LatestSnapshotTimeMs,
		&result.PriceServerTimeMs,
	)
}

func (m *defaultDeliveryPreflightModel) inspectDeliveryFacts(
	ctx context.Context,
	tenantID int64,
	symbolID int64,
	result *DeliveryPreflightResult,
) error {
	const query = `
SELECT
  (SELECT COUNT(*) FROM t_trade_order AS o
   WHERE o.tenant_id=? AND o.symbol_id=?),
  (SELECT COUNT(*) FROM t_trade_fill AS f
   WHERE f.tenant_id=? AND f.symbol_id=?),
  (SELECT COUNT(*) FROM t_contract_position AS p
   WHERE p.tenant_id=? AND p.symbol_id=?),
  (SELECT COUNT(*) FROM t_contract_position_history AS h
   WHERE h.tenant_id=? AND h.symbol_id=?),
  (SELECT COUNT(*)
   FROM t_trade_asset_reservation AS r
   JOIN t_trade_order AS o
     ON o.tenant_id=r.tenant_id AND o.id=r.order_id
   WHERE r.tenant_id=? AND o.symbol_id=?),
  (SELECT COUNT(DISTINCT i.id)
   FROM t_trade_settlement_instruction AS i
   LEFT JOIN t_trade_order AS o
     ON o.tenant_id=i.tenant_id AND o.id=i.order_id
   LEFT JOIN t_trade_fill AS f
     ON f.tenant_id=i.tenant_id AND f.id=i.fill_id
   LEFT JOIN t_contract_position AS p
     ON p.tenant_id=i.tenant_id AND p.id=i.position_id
   LEFT JOIN t_contract_delivery_batch AS b
     ON b.tenant_id=i.tenant_id
    AND i.biz_type='delivery'
    AND b.batch_no=i.batch_no
   WHERE i.tenant_id=?
     AND (o.symbol_id=? OR f.symbol_id=? OR p.symbol_id=? OR b.symbol_id=?)),
  (SELECT COUNT(*) FROM t_contract_delivery_batch AS b
   WHERE b.tenant_id=? AND b.symbol_id=?),
  (SELECT COUNT(*) FROM t_contract_delivery_settlement AS s
   WHERE s.tenant_id=? AND s.symbol_id=?)`
	return m.db.QueryRowContext(
		ctx,
		query,
		tenantID, symbolID,
		tenantID, symbolID,
		tenantID, symbolID,
		tenantID, symbolID,
		tenantID, symbolID,
		tenantID, symbolID, symbolID, symbolID, symbolID,
		tenantID, symbolID,
		tenantID, symbolID,
	).Scan(
		&result.OrderCount,
		&result.FillCount,
		&result.PositionCount,
		&result.PositionHistoryCount,
		&result.ReservationCount,
		&result.SettlementInstructionCount,
		&result.DeliveryBatchCount,
		&result.DeliverySettlementCount,
	)
}
