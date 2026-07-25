package adminlogic

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"wklive/proto/liquidity"
	"wklive/proto/trade"
	"wklive/services/liquidity/internal/svc"
	"wklive/services/liquidity/models"
)

func parseNumber(name, value string) (float64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}
	number, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("%s is invalid: %w", name, err)
	}
	if number < 0 {
		return 0, fmt.Errorf("%s cannot be negative", name)
	}
	return number, nil
}

func parseSignedNumber(name, value string) (float64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}
	number, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("%s is invalid: %w", name, err)
	}
	return number, nil
}

func validateReferencePriceSources(value string) error {
	sources := strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == '|' || r == ';'
	})
	if len(sources) == 0 {
		return fmt.Errorf("reference_price_source is required")
	}
	for _, source := range sources {
		parts := strings.Split(strings.TrimSpace(source), ":")
		if len(parts) != 3 ||
			strings.TrimSpace(parts[0]) == "" ||
			strings.TrimSpace(parts[1]) == "" ||
			strings.TrimSpace(parts[2]) == "" {
			return fmt.Errorf("reference_price_source must use category:market:symbol format")
		}
	}
	return nil
}

func buildSymbolConfig(ctx context.Context, svcCtx *svc.ServiceContext, in *liquidity.SaveSymbolConfigReq, current *models.TLiquiditySymbolConfig) (*models.TLiquiditySymbolConfig, error) {
	if in.SymbolId <= 0 {
		return nil, fmt.Errorf("symbol_id is required")
	}
	detail, err := svcCtx.TradeClient.GetSymbolDetail(ctx, &trade.GetSymbolDetailReq{
		// TenantId: 0, // TODO 这里之后看怎么处理
		SymbolId: in.SymbolId,
	})
	if err != nil {
		return nil, fmt.Errorf("get trade symbol: %w", err)
	}
	if detail.GetBase().GetCode() != 200 || detail.GetData().GetSymbol() == nil {
		return nil, fmt.Errorf("trade symbol unavailable: %s", detail.GetBase().GetMsg())
	}
	symbol := detail.GetData().GetSymbol()
	values := make([]float64, 13)
	inputs := []struct{ name, value string }{
		{"reprice_threshold_bps", in.RepriceThresholdBps}, {"base_spread_bps", in.BaseSpreadBps},
		{"max_spread_bps", in.MaxSpreadBps}, {"max_price_deviation_bps", in.MaxPriceDeviationBps},
		{"min_quote_qty", in.MinQuoteQty}, {"max_quote_qty", in.MaxQuoteQty},
		{"max_quote_notional", in.MaxQuoteNotional}, {"target_base_inventory", in.TargetBaseInventory},
		{"min_base_inventory", in.MinBaseInventory}, {"max_base_inventory", in.MaxBaseInventory},
		{"max_net_exposure", in.MaxNetExposure}, {"max_daily_notional", in.MaxDailyNotional},
		{"inventory_skew_bps", in.InventorySkewBps},
	}
	for i, item := range inputs {
		values[i], err = parseNumber(item.name, item.value)
		if err != nil {
			return nil, err
		}
	}
	hedgeThreshold, err := parseNumber("hedge_threshold", in.HedgeThreshold)
	if err != nil {
		return nil, err
	}
	hedgeRatio, err := parseNumber("hedge_ratio", in.HedgeRatio)
	if err != nil || hedgeRatio > 1 {
		return nil, fmt.Errorf("hedge_ratio must be between 0 and 1")
	}
	priceTick, err := parseNumber("price_tick", symbol.PriceTick)
	if err != nil || priceTick <= 0 {
		return nil, fmt.Errorf("trade symbol price_tick is invalid")
	}
	qtyStep, err := parseNumber("qty_step", symbol.QtyStep)
	if err != nil || qtyStep <= 0 {
		return nil, fmt.Errorf("trade symbol qty_step is invalid")
	}
	if in.LiquidityMode == liquidity.LiquidityMode_LIQUIDITY_MODE_UNKNOWN {
		return nil, fmt.Errorf("liquidity_mode is required")
	}
	if err := validateReferencePriceSources(in.ReferencePriceSource); err != nil {
		return nil, err
	}
	if values[4] <= 0 || (values[5] > 0 && values[5] < values[4]) {
		return nil, fmt.Errorf("invalid quote quantity range")
	}
	if values[2] > 0 && values[1] > values[2] {
		return nil, fmt.Errorf("base_spread_bps cannot exceed max_spread_bps")
	}
	now := time.Now().UnixMilli()
	row := current
	if row == nil {
		row = &models.TLiquiditySymbolConfig{
			SymbolId: in.SymbolId,
			Status:   int64(liquidity.SymbolLiquidityStatus_SYMBOL_LIQUIDITY_STATUS_DISABLED),
			Version:  1, CreateTimes: now,
		}
	}
	row.Symbol, row.ProductType, row.ContractType = symbol.Symbol, int64(symbol.ProductType), int64(symbol.ContractType)
	row.LiquidityMode, row.InternalProviderId, row.ExternalProviderId = int64(in.LiquidityMode), in.InternalProviderId, in.ExternalProviderId
	row.ExternalSymbol, row.ReferencePriceSource, row.ReferencePriceKind = strings.TrimSpace(in.ExternalSymbol), strings.TrimSpace(in.ReferencePriceSource), strings.TrimSpace(in.ReferencePriceKind)
	row.QuoteValidityMs, row.RefreshIntervalMs, row.QuoteTtlMs = int64(in.QuoteValidityMs), int64(in.RefreshIntervalMs), int64(in.QuoteTtlMs)
	row.RepriceThresholdBps, row.BaseSpreadBps, row.MaxSpreadBps, row.MaxPriceDeviationBps = values[0], values[1], values[2], values[3]
	row.PriceTick, row.QtyStep, row.MinQuoteQty, row.MaxQuoteQty, row.MaxQuoteNotional = priceTick, qtyStep, values[4], values[5], values[6]
	row.TargetBaseInventory, row.MinBaseInventory, row.MaxBaseInventory = values[7], values[8], values[9]
	row.MaxNetExposure, row.MaxDailyNotional, row.InventorySkewBps = values[10], values[11], values[12]
	row.HedgeThreshold, row.HedgeRatio = hedgeThreshold, hedgeRatio
	row.SelfTradePrevention, row.UpdateTimes = int64(in.SelfTradePrevention), now
	return row, nil
}

func changeSymbolStatus(ctx context.Context, svcCtx *svc.ServiceContext, in *liquidity.SymbolActionReq, target liquidity.SymbolLiquidityStatus) error {
	if in.ConfigId <= 0 {
		return fmt.Errorf("config_id is required")
	}
	row, err := svcCtx.SymbolConfigModel.FindOne(ctx, in.ConfigId)
	if err != nil {
		return err
	}
	if row.Version != in.Version {
		return fmt.Errorf("symbol config version conflict")
	}
	if target == liquidity.SymbolLiquidityStatus_SYMBOL_LIQUIDITY_STATUS_RUNNING {
		enabledLevels, err := svcCtx.StrategyLevelModel.CountEnabled(ctx, in.ConfigId)
		if err != nil {
			return err
		}
		if enabledLevels == 0 {
			return fmt.Errorf("at least one enabled strategy level is required")
		}
	}
	row.Status, row.PauseReason = int64(target), strings.TrimSpace(in.Reason)
	if target == liquidity.SymbolLiquidityStatus_SYMBOL_LIQUIDITY_STATUS_RUNNING {
		row.PauseReason = ""
	}
	row.Version++
	row.UpdateTimes = time.Now().UnixMilli()
	return svcCtx.SymbolConfigModel.Update(ctx, row)
}
