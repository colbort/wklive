package tasklogic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"wklive/common/conv"
	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/trade"
	"wklive/services/trade/models"

	"github.com/shopspring/decimal"
)

func mustParseFloat(v string) decimal.Decimal {
	value, _ := conv.ParseDecimalField(v)
	return value
}

func enableToProto(value int64) common.Enable {
	return common.Enable(value)
}

func enableToModel(value common.Enable, defaultValue int64) int64 {
	if value == common.Enable_ENABLE_UNKNOWN {
		return defaultValue
	}
	return int64(value)
}

func yesNoToModel(value common.YesNo, defaultValue int64) int64 {
	if value == common.YesNo_YES_NO_UNKNOWN {
		return defaultValue
	}
	return int64(value)
}

type orderAssetExt struct {
	FreezeNo          string `json:"freezeNo,omitempty"`
	OriginalOrderType int64  `json:"originalOrderType,omitempty"`
	TriggeredAt       int64  `json:"triggeredAt,omitempty"`
	TriggerPrice      string `json:"triggerPrice,omitempty"`
	TriggerSource     string `json:"triggerSource,omitempty"`
}

func symbolToProto(item *models.TTradeSymbol) *trade.TradeSymbol {
	if item == nil {
		return nil
	}
	return &trade.TradeSymbol{
		Id:                item.Id,
		TenantId:          item.TenantId,
		Symbol:            item.Symbol,
		DisplaySymbol:     item.DisplaySymbol,
		ProductType:       common.ProductType(item.ProductType),
		BaseAsset:         item.BaseAsset,
		QuoteAsset:        item.QuoteAsset,
		SettleAsset:       item.SettleAsset,
		ContractType:      common.ContractType(item.ContractType),
		ContractValueType: trade.ContractValueType(item.ContractValueType),
		MarginAsset:       item.MarginAsset,
		Status:            trade.SymbolStatus(item.Status),
		PriceScale:        item.PriceScale,
		QtyScale:          item.QtyScale,
		MinPrice:          conv.FloatString(item.MinPrice),
		MaxPrice:          conv.FloatString(item.MaxPrice),
		PriceTick:         conv.FloatString(item.PriceTick),
		MinQty:            conv.FloatString(item.MinQty),
		MaxQty:            conv.FloatString(item.MaxQty),
		QtyStep:           conv.FloatString(item.QtyStep),
		MinNotional:       conv.FloatString(item.MinNotional),
		MaxNotional:       conv.FloatString(item.MaxNotional),
		ListingTime:       item.ListingTime,
		TradingStartTime:  item.TradingStartTime,
		TradingEndTime:    item.TradingEndTime,
		Sort:              item.Sort,
		Remark:            item.Remark,
		CreateTimes:       item.CreateTimes,
		UpdateTimes:       item.UpdateTimes,
	}
}

func spotSymbolToProto(item *models.TTradeSymbolSpot) *trade.TradeSymbolSpot {
	if item == nil {
		return nil
	}
	return &trade.TradeSymbolSpot{
		Id:           item.Id,
		TenantId:     item.TenantId,
		SymbolId:     item.SymbolId,
		MakerFeeRate: conv.FloatString(item.MakerFeeRate),
		TakerFeeRate: conv.FloatString(item.TakerFeeRate),
		BuyEnabled:   enableToProto(item.BuyEnabled),
		SellEnabled:  enableToProto(item.SellEnabled),
		CreateTimes:  item.CreateTimes,
		UpdateTimes:  item.UpdateTimes,
	}
}

func contractSymbolToProto(item *models.TTradeSymbolContract) *trade.TradeSymbolContract {
	if item == nil {
		return nil
	}
	return &trade.TradeSymbolContract{
		Id:                       item.Id,
		TenantId:                 item.TenantId,
		SymbolId:                 item.SymbolId,
		ContractSize:             conv.FloatString(item.ContractSize),
		Multiplier:               conv.FloatString(item.Multiplier),
		MaintenanceMarginRate:    conv.FloatString(item.MaintenanceMarginRate),
		InitialMarginRate:        conv.FloatString(item.InitialMarginRate),
		MakerFeeRate:             conv.FloatString(item.MakerFeeRate),
		TakerFeeRate:             conv.FloatString(item.TakerFeeRate),
		FundingIntervalMinutes:   item.FundingIntervalMinutes,
		FundingRateCap:           conv.FloatString(item.FundingRateCap),
		FundingRateFloor:         conv.FloatString(item.FundingRateFloor),
		FundingRateSource:        item.FundingRateSource,
		IndexSymbol:              item.IndexSymbol,
		MarkPriceSource:          item.MarkPriceSource,
		SettlementPriceSource:    item.SettlementPriceSource,
		DeliveryTime:             item.DeliveryTime,
		OpenCutoffTime:           item.OpenCutoffTime,
		MatchingStopTime:         item.MatchingStopTime,
		SettlementWindowSeconds:  item.SettlementWindowSeconds,
		SettlementPriceAlgorithm: item.SettlementPriceAlgorithm,
		DeliveryFeeRate:          conv.FloatString(item.DeliveryFeeRate),
		LiquidationFeeRate:       conv.FloatString(item.LiquidationFeeRate),
		SupportCross:             item.SupportCross,
		SupportIsolated:          item.SupportIsolated,
		OpenLongEnabled:          enableToProto(item.OpenLongEnabled),
		OpenShortEnabled:         enableToProto(item.OpenShortEnabled),
		CloseLongEnabled:         enableToProto(item.CloseLongEnabled),
		CloseShortEnabled:        enableToProto(item.CloseShortEnabled),
		CreateTimes:              item.CreateTimes,
		UpdateTimes:              item.UpdateTimes,
	}
}

func secondsSymbolToProto(item *models.TTradeSymbolSeconds) *trade.TradeSymbolSeconds {
	if item == nil {
		return nil
	}
	return &trade.TradeSymbolSeconds{
		Id: item.Id, TenantId: item.TenantId, SymbolId: item.SymbolId,
		DurationSeconds: item.DurationSeconds, PayoutRate: conv.FloatString(item.PayoutRate),
		DrawRule: trade.SecondsDrawRule(item.DrawRule), StartPriceSource: item.StartPriceSource,
		SettlementPriceSource: item.SettlementPriceSource, QuoteValidityMs: item.QuoteValidityMs,
		MinStake: conv.FloatString(item.MinStake), MaxStake: conv.FloatString(item.MaxStake),
		FeeRate: conv.FloatString(item.FeeRate), SettlementWindowMs: item.SettlementWindowMs,
		SettlementPriceAlgorithm: item.SettlementPriceAlgorithm,
		DrawTolerance:            conv.FloatString(item.DrawTolerance), MaxExposureAmount: conv.FloatString(item.MaxExposureAmount),
		UpEnabled: enableToProto(item.UpEnabled), DownEnabled: enableToProto(item.DownEnabled),
		CreateTimes: item.CreateTimes, UpdateTimes: item.UpdateTimes,
	}
}

func symbolSessionToProto(item *models.TTradeSymbolSession) *trade.TradeSymbolSession {
	if item == nil {
		return nil
	}
	return &trade.TradeSymbolSession{
		Id:          item.Id,
		TenantId:    item.TenantId,
		SymbolId:    item.SymbolId,
		DayOfWeek:   item.DayOfWeek,
		StartSecond: item.StartSecond,
		EndSecond:   item.EndSecond,
		Timezone:    item.Timezone,
		Enabled:     enableToProto(item.Enabled),
	}
}

func userConfigToProto(item *models.TTradeUserConfig) *trade.TradeUserConfig {
	if item == nil {
		return nil
	}
	return &trade.TradeUserConfig{
		Id:           item.Id,
		TenantId:     item.TenantId,
		UserId:       item.UserId,
		ProductType:  common.ProductType(item.ProductType),
		SymbolId:     item.SymbolId,
		TradeEnabled: enableToProto(item.TradeEnabled),
		CreateTimes:  item.CreateTimes,
		UpdateTimes:  item.UpdateTimes,
	}
}

func contractUserConfigToProto(item *models.TContractUserConfig) *trade.ContractUserConfig {
	if item == nil {
		return nil
	}
	return &trade.ContractUserConfig{Id: item.Id, TenantId: item.TenantId, UserId: item.UserId, SymbolId: item.SymbolId, PositionMode: trade.PositionMode(item.PositionMode), MarginMode: trade.MarginMode(item.MarginMode), DefaultLeverage: item.DefaultLeverage, CreateTimes: item.CreateTimes, UpdateTimes: item.UpdateTimes}
}

func orderToProto(item *models.TTradeOrder) *trade.TradeOrder {
	if item == nil {
		return nil
	}
	return &trade.TradeOrder{
		Id:                item.Id,
		TenantId:          item.TenantId,
		OrderNo:           item.OrderNo,
		ClientOrderId:     item.ClientOrderId.String,
		RequestHash:       item.RequestHash,
		UserId:            item.UserId,
		SymbolId:          item.SymbolId,
		ProductType:       common.ProductType(item.ProductType),
		ContractType:      common.ContractType(item.ContractType),
		ContractValueType: trade.ContractValueType(item.ContractValueType),
		Side:              common.Side(item.Side),
		PositionSide:      trade.PositionSide(item.PositionSide),
		OrderType:         trade.OrderType(item.OrderType),
		TimeInForce:       trade.TimeInForce(item.TimeInForce),
		Status:            trade.OrderStatus(item.Status),
		DisplayStatus:     orderDisplayStatus(item),
		Price:             conv.FloatString(item.Price),
		Qty:               conv.FloatString(item.Qty),
		Amount:            conv.FloatString(item.Amount),
		FilledQty:         conv.FloatString(item.FilledQty),
		FilledAmount:      conv.FloatString(item.FilledAmount),
		CanceledQty:       conv.FloatString(item.CanceledQty),
		AvgPrice:          conv.FloatString(item.AvgPrice),
		Fee:               conv.FloatString(item.Fee),
		FeeAsset:          item.FeeAsset,
		Source:            trade.OrderSourceType(item.Source),
		IsReduceOnly:      common.YesNo(item.IsReduceOnly),
		IsClosePosition:   common.YesNo(item.IsClosePosition),
		TriggerPrice:      conv.FloatString(item.TriggerPrice),
		TriggerType:       trade.TriggerType(item.TriggerType),
		TriggerKind:       trade.TriggerKind(item.TriggerKind),
		OcoGroupNo:        item.OcoGroupNo,
		ExpireAt:          item.ExpireAt,
		TriggeredAt:       item.TriggeredAt,
		CompletionReason:  item.CompletionReason,
		CancelReason:      item.CancelReason,
		BizExt:            conv.NullStringValue(item.BizExt),
		CreateTimes:       item.CreateTimes,
		UpdateTimes:       item.UpdateTimes,
		Version:           item.Version,
	}
}

func orderDisplayStatus(order *models.TTradeOrder) trade.OrderDisplayStatus {
	if order == nil {
		return trade.OrderDisplayStatus_ORDER_DISPLAY_STATUS_UNKNOWN
	}
	if trade.OrderStatus(order.Status) == trade.OrderStatus_ORDER_STATUS_CANCELED &&
		(order.FilledQty.IsPositive() || order.FilledAmount.IsPositive()) {
		return trade.OrderDisplayStatus_ORDER_DISPLAY_STATUS_PART_FILLED
	}
	switch trade.OrderStatus(order.Status) {
	case trade.OrderStatus_ORDER_STATUS_FREEZING:
		return trade.OrderDisplayStatus_ORDER_DISPLAY_STATUS_FREEZING
	case trade.OrderStatus_ORDER_STATUS_TRIGGER_WAITING:
		return trade.OrderDisplayStatus_ORDER_DISPLAY_STATUS_TRIGGER_WAITING
	case trade.OrderStatus_ORDER_STATUS_PENDING:
		return trade.OrderDisplayStatus_ORDER_DISPLAY_STATUS_PENDING
	case trade.OrderStatus_ORDER_STATUS_PART_FILLED:
		return trade.OrderDisplayStatus_ORDER_DISPLAY_STATUS_PART_FILLED
	case trade.OrderStatus_ORDER_STATUS_SETTLEMENT_PENDING:
		return trade.OrderDisplayStatus_ORDER_DISPLAY_STATUS_SETTLING
	case trade.OrderStatus_ORDER_STATUS_FILLED:
		return trade.OrderDisplayStatus_ORDER_DISPLAY_STATUS_FILLED
	case trade.OrderStatus_ORDER_STATUS_CANCELING:
		return trade.OrderDisplayStatus_ORDER_DISPLAY_STATUS_CANCELING
	case trade.OrderStatus_ORDER_STATUS_CANCELED:
		return trade.OrderDisplayStatus_ORDER_DISPLAY_STATUS_CANCELED
	case trade.OrderStatus_ORDER_STATUS_EXPIRING:
		return trade.OrderDisplayStatus_ORDER_DISPLAY_STATUS_EXPIRING
	case trade.OrderStatus_ORDER_STATUS_EXPIRED:
		return trade.OrderDisplayStatus_ORDER_DISPLAY_STATUS_EXPIRED
	case trade.OrderStatus_ORDER_STATUS_REJECTED:
		return trade.OrderDisplayStatus_ORDER_DISPLAY_STATUS_REJECTED
	default:
		return trade.OrderDisplayStatus_ORDER_DISPLAY_STATUS_UNKNOWN
	}
}

func orderSpotToProto(item *models.TTradeOrderSpot) *trade.TradeOrderSpot {
	if item == nil {
		return nil
	}
	return &trade.TradeOrderSpot{
		Id:           item.Id,
		TenantId:     item.TenantId,
		OrderId:      item.OrderId,
		FrozenAsset:  item.FrozenAsset,
		FrozenAmount: conv.FloatString(item.FrozenAmount),
		SettleAsset:  item.SettleAsset,
		SettleAmount: conv.FloatString(item.SettleAmount),
		CreateTimes:  item.CreateTimes,
		UpdateTimes:  item.UpdateTimes,
	}
}

func orderContractToProto(item *models.TTradeOrderContract) *trade.TradeOrderContract {
	if item == nil {
		return nil
	}
	return &trade.TradeOrderContract{
		Id:                item.Id,
		TenantId:          item.TenantId,
		OrderId:           item.OrderId,
		MarginMode:        trade.MarginMode(item.MarginMode),
		Leverage:          item.Leverage,
		MarginAsset:       item.MarginAsset,
		MarginAmount:      conv.FloatString(item.MarginAmount),
		ReservedCloseQty:  conv.FloatString(item.ReservedCloseQty),
		RiskPrice:         conv.FloatString(item.RiskPrice),
		RiskTierId:        item.RiskTierId,
		ClosePositionType: trade.ClosePositionType(item.ClosePositionType),
		LiquidationPrice:  conv.FloatString(item.LiquidationPrice),
		TakeProfitPrice:   conv.FloatString(item.TakeProfitPrice),
		StopLossPrice:     conv.FloatString(item.StopLossPrice),
		CreateTimes:       item.CreateTimes,
		UpdateTimes:       item.UpdateTimes,
	}
}

func fillToProto(item *models.TTradeFill) *trade.TradeFill {
	if item == nil {
		return nil
	}
	return &trade.TradeFill{
		Id:                   item.Id,
		TenantId:             item.TenantId,
		FillNo:               item.FillNo,
		OrderId:              item.OrderId,
		OrderNo:              item.OrderNo,
		UserId:               item.UserId,
		SymbolId:             item.SymbolId,
		ProductType:          common.ProductType(item.ProductType),
		ContractType:         common.ContractType(item.ContractType),
		ContractValueType:    trade.ContractValueType(item.ContractValueType),
		MatchNo:              item.MatchNo,
		Side:                 common.Side(item.Side),
		PositionSide:         trade.PositionSide(item.PositionSide),
		Price:                conv.FloatString(item.Price),
		Qty:                  conv.FloatString(item.Qty),
		Amount:               conv.FloatString(item.Amount),
		Fee:                  conv.FloatString(item.Fee),
		FeeAsset:             item.FeeAsset,
		LiquidityType:        trade.LiquidityType(item.LiquidityType),
		RealizedPnl:          conv.FloatString(item.RealizedPnl),
		MatchTime:            item.MatchTime,
		CreateTimes:          item.CreateTimes,
		SettlementStatus:     trade.FillSettlementStatus(item.SettlementStatus),
		SettlementRetryCount: item.SettlementRetryCount,
		SettledAt:            item.SettledAt,
	}
}

func cancelLogToProto(item *models.TTradeCancelLog) *trade.TradeCancelLog {
	if item == nil {
		return nil
	}
	return &trade.TradeCancelLog{
		Id:           item.Id,
		TenantId:     item.TenantId,
		OrderId:      item.OrderId,
		OrderNo:      item.OrderNo,
		UserId:       item.UserId,
		CancelSource: trade.CancelSource(item.CancelSource),
		CancelReason: item.CancelReason,
		CreateTimes:  item.CreateTimes,
	}
}

func positionToProto(item *models.TContractPosition) *trade.ContractPosition {
	if item == nil {
		return nil
	}
	return &trade.ContractPosition{
		Id:                item.Id,
		TenantId:          item.TenantId,
		UserId:            item.UserId,
		SymbolId:          item.SymbolId,
		ContractType:      common.ContractType(item.ContractType),
		ContractValueType: trade.ContractValueType(item.ContractValueType),
		PositionSide:      trade.PositionSide(item.PositionSide),
		MarginMode:        trade.MarginMode(item.MarginMode),
		Status:            trade.PositionStatus(item.Status),
		Leverage:          item.Leverage,
		Qty:               conv.FloatString(item.Qty),
		AvailQty:          conv.FloatString(item.AvailQty),
		FrozenQty:         conv.FloatString(item.FrozenQty),
		OpenAvgPrice:      conv.FloatString(item.OpenAvgPrice),
		MarkPrice:         conv.FloatString(item.MarkPrice),
		MarkSnapshotId:    item.MarkSnapshotId,
		MarginAsset:       item.MarginAsset,
		PositionMargin:    conv.FloatString(item.PositionMargin),
		MaintenanceMargin: conv.FloatString(item.MaintenanceMargin),
		IsolatedMargin:    conv.FloatString(item.IsolatedMargin),
		UnrealizedPnl:     conv.FloatString(item.UnrealizedPnl),
		RealizedPnl:       conv.FloatString(item.RealizedPnl),
		LiquidationPrice:  conv.FloatString(item.LiquidationPrice),
		BankruptcyPrice:   conv.FloatString(item.BankruptcyPrice),
		RiskRate:          conv.FloatString(item.RiskRate),
		AdlRank:           item.AdlRank,
		Version:           item.Version,
		LastFundingTime:   item.LastFundingTime,
		ClosedAt:          item.ClosedAt,
		CreateTimes:       item.CreateTimes,
		UpdateTimes:       item.UpdateTimes,
	}
}

func positionHistoryToProto(item *models.TContractPositionHistory) *trade.ContractPositionHistory {
	if item == nil {
		return nil
	}
	return &trade.ContractPositionHistory{
		Id:                   item.Id,
		TenantId:             item.TenantId,
		PositionId:           item.PositionId,
		UserId:               item.UserId,
		SymbolId:             item.SymbolId,
		ContractType:         common.ContractType(item.ContractType),
		ContractValueType:    trade.ContractValueType(item.ContractValueType),
		PositionSide:         trade.PositionSide(item.PositionSide),
		ActionType:           trade.PositionActionType(item.ActionType),
		BeforeQty:            conv.FloatString(item.BeforeQty),
		AfterQty:             conv.FloatString(item.AfterQty),
		BeforeAvailQty:       conv.FloatString(item.BeforeAvailQty),
		AfterAvailQty:        conv.FloatString(item.AfterAvailQty),
		BeforeFrozenQty:      conv.FloatString(item.BeforeFrozenQty),
		AfterFrozenQty:       conv.FloatString(item.AfterFrozenQty),
		BeforeOpenAvgPrice:   conv.FloatString(item.BeforeOpenAvgPrice),
		AfterOpenAvgPrice:    conv.FloatString(item.AfterOpenAvgPrice),
		BeforePositionMargin: conv.FloatString(item.BeforePositionMargin),
		AfterPositionMargin:  conv.FloatString(item.AfterPositionMargin),
		BeforeIsolatedMargin: conv.FloatString(item.BeforeIsolatedMargin),
		AfterIsolatedMargin:  conv.FloatString(item.AfterIsolatedMargin),
		BeforeUnrealizedPnl:  conv.FloatString(item.BeforeUnrealizedPnl),
		AfterUnrealizedPnl:   conv.FloatString(item.AfterUnrealizedPnl),
		RealizedPnlDelta:     conv.FloatString(item.RealizedPnlDelta),
		FeeDelta:             conv.FloatString(item.FeeDelta),
		FeeAsset:             item.FeeAsset,
		MarkPrice:            conv.FloatString(item.MarkPrice),
		RefOrderId:           item.RefOrderId,
		RefFillId:            item.RefFillId,
		OperatorId:           item.OperatorId,
		Source:               trade.SourceType(item.Source),
		Remark:               item.Remark,
		CreateTimes:          item.CreateTimes,
	}
}

func marginSnapshotToProto(item *models.TContractMarginSnapshot) *trade.ContractMarginSnapshot {
	if item == nil {
		return nil
	}
	return &trade.ContractMarginSnapshot{
		Id:               item.Id,
		TenantId:         item.TenantId,
		UserId:           item.UserId,
		MarginAsset:      item.MarginAsset,
		WalletBalance:    conv.FloatString(item.WalletBalance),
		AvailableBalance: conv.FloatString(item.AvailableBalance),
		FrozenBalance:    conv.FloatString(item.FrozenBalance),
		PositionMargin:   conv.FloatString(item.PositionMargin),
		OrderMargin:      conv.FloatString(item.OrderMargin),
		UnrealizedPnl:    conv.FloatString(item.UnrealizedPnl),
		RealizedPnl:      conv.FloatString(item.RealizedPnl),
		Version:          item.Version,
		CreateTimes:      item.CreateTimes,
		UpdateTimes:      item.UpdateTimes,
		SourceEventNo:    item.SourceEventNo.String,
		SnapshotTime:     item.SnapshotTime,
	}
}

func leverageConfigToProto(item *models.TContractLeverageConfig) *trade.ContractLeverageConfig {
	if item == nil {
		return nil
	}
	return &trade.ContractLeverageConfig{
		Id:            item.Id,
		TenantId:      item.TenantId,
		UserId:        item.UserId,
		SymbolId:      item.SymbolId,
		MarginMode:    trade.MarginMode(item.MarginMode),
		LongLeverage:  item.LongLeverage,
		ShortLeverage: item.ShortLeverage,
		OperatorId:    item.OperatorId,
		Source:        trade.SourceType(item.Source),
		Enabled:       enableToProto(item.Enabled),
		Remark:        item.Remark,
		CreateTimes:   item.CreateTimes,
		UpdateTimes:   item.UpdateTimes,
	}
}

func symbolLeverageConfigToProto(item *models.TTradeSymbolLeverageConfig, defaultLeverage ...int64) *trade.TradeSymbolLeverageConfig {
	if item == nil {
		return nil
	}
	var defaultValue int64
	if len(defaultLeverage) > 0 {
		defaultValue = defaultLeverage[0]
	}
	return &trade.TradeSymbolLeverageConfig{
		Id:              item.Id,
		TenantId:        item.TenantId,
		SymbolId:        item.SymbolId,
		MarginMode:      trade.MarginMode(item.MarginMode),
		LeverageValues:  parseLeverageValues(item.LeverageValues),
		DefaultLeverage: defaultValue,
		Enabled:         enableToProto(item.Enabled),
		Sort:            item.Sort,
		Remark:          item.Remark,
		CreateTimes:     item.CreateTimes,
		UpdateTimes:     item.UpdateTimes,
	}
}

func findDefaultLeverage(ctx context.Context, model models.TTradeSymbolLeverageDefaultModel, tenantId, symbolId, marginMode int64) (int64, error) {
	if model == nil {
		return 0, nil
	}
	item, err := model.FindOneByTenantIdSymbolIdMarginMode(ctx, tenantId, symbolId, marginMode)
	if errors.Is(err, models.ErrNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return item.Leverage, nil
}

func parseLeverageValues(value string) []int64 {
	var jsonValues []int64
	if json.Unmarshal([]byte(value), &jsonValues) == nil {
		return normalizeLeverageValues(jsonValues)
	}
	parts := strings.Split(value, ",")
	values := make([]int64, 0, len(parts))
	for _, part := range parts {
		next, err := strconv.ParseInt(strings.TrimSpace(part), 10, 64)
		if err != nil || next <= 0 {
			continue
		}
		values = append(values, next)
	}
	return normalizeLeverageValues(values)
}

func normalizeLeverageValues(values []int64) []int64 {
	result := make([]int64, 0, len(values))
	seen := make(map[int64]struct{}, len(values))
	for _, value := range values {
		if value <= 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i] < result[j]
	})
	return result
}

func joinLeverageValues(values []int64, maxLeverage int64) (string, []int64) {
	if maxLeverage <= 0 {
		maxLeverage = 1
	}
	seen := make(map[int64]struct{}, len(values))
	result := make([]int64, 0, len(values))
	for _, value := range values {
		if value <= 0 || value > maxLeverage {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	if len(result) == 0 {
		result = append(result, 1)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i] < result[j]
	})

	parts := make([]string, 0, len(result))
	for _, value := range result {
		parts = append(parts, strconv.FormatInt(value, 10))
	}
	return strings.Join(parts, ","), result
}

func containsLeverage(values []int64, leverage int64) bool {
	for _, value := range values {
		if value == leverage {
			return true
		}
	}
	return false
}

func isDerivativeProduct(productType common.ProductType) bool {
	return productType == common.ProductType_PRODUCT_TYPE_DERIVATIVE
}

func riskUserTradeLimitToProto(item *models.TRiskUserTradeLimit) *trade.RiskUserTradeLimit {
	if item == nil {
		return nil
	}
	return &trade.RiskUserTradeLimit{
		Id:                   item.Id,
		TenantId:             item.TenantId,
		UserId:               item.UserId,
		ProductType:          common.ProductType(item.ProductType),
		CanOpen:              item.CanOpen,
		CanClose:             item.CanClose,
		CanCancel:            item.CanCancel,
		CanTriggerOrder:      item.CanTriggerOrder,
		CanApiTrade:          item.CanApiTrade,
		TradeEnabled:         enableToProto(item.TradeEnabled),
		OnlyReduceOnly:       common.Enable(item.OnlyReduceOnly),
		MaxOpenOrderCount:    item.MaxOpenOrderCount,
		MaxOrderCountPerDay:  item.MaxOrderCountPerDay,
		MaxCancelCountPerDay: item.MaxCancelCountPerDay,
		MaxOpenNotional:      conv.FloatString(item.MaxOpenNotional),
		MaxPositionNotional:  conv.FloatString(item.MaxPositionNotional),
		RiskLevel:            trade.RiskLevel(item.RiskLevel),
		OperatorId:           item.OperatorId,
		Source:               trade.SourceType(item.Source),
		Enabled:              enableToProto(item.Enabled),
		EffectiveStartTime:   item.EffectiveStartTime,
		EffectiveEndTime:     item.EffectiveEndTime,
		Remark:               item.Remark,
		CreateTimes:          item.CreateTimes,
		UpdateTimes:          item.UpdateTimes,
	}
}

func riskUserSymbolLimitToProto(item *models.TRiskUserSymbolLimit) *trade.RiskUserSymbolLimit {
	if item == nil {
		return nil
	}
	return &trade.RiskUserSymbolLimit{
		Id:                  item.Id,
		TenantId:            item.TenantId,
		UserId:              item.UserId,
		SymbolId:            item.SymbolId,
		MaxPositionQty:      conv.FloatString(item.MaxPositionQty),
		MaxPositionNotional: conv.FloatString(item.MaxPositionNotional),
		MaxOpenOrders:       item.MaxOpenOrders,
		MaxOrderQty:         conv.FloatString(item.MaxOrderQty),
		MaxOrderNotional:    conv.FloatString(item.MaxOrderNotional),
		MinOrderQty:         conv.FloatString(item.MinOrderQty),
		MinOrderNotional:    conv.FloatString(item.MinOrderNotional),
		MaxLongPositionQty:  conv.FloatString(item.MaxLongPositionQty),
		MaxShortPositionQty: conv.FloatString(item.MaxShortPositionQty),
		PriceDeviationRate:  conv.FloatString(item.PriceDeviationRate),
		OperatorId:          item.OperatorId,
		Source:              trade.SourceType(item.Source),
		Enabled:             enableToProto(item.Enabled),
		EffectiveStartTime:  item.EffectiveStartTime,
		EffectiveEndTime:    item.EffectiveEndTime,
		Remark:              item.Remark,
		CreateTimes:         item.CreateTimes,
		UpdateTimes:         item.UpdateTimes,
	}
}

func riskOrderCheckLogToProto(item *models.TRiskOrderCheckLog) *trade.RiskOrderCheckLog {
	if item == nil {
		return nil
	}
	return &trade.RiskOrderCheckLog{
		Id:            item.Id,
		TenantId:      item.TenantId,
		OrderNo:       item.OrderNo,
		ClientOrderId: item.ClientOrderId,
		UserId:        item.UserId,
		SymbolId:      item.SymbolId,
		ProductType:   common.ProductType(item.ProductType),
		CheckType:     trade.RiskCheckType(item.CheckType),
		CheckResult:   trade.RiskCheckResult(item.CheckResult),
		RejectCode:    item.RejectCode,
		RejectMsg:     item.RejectMsg,
		RequestPrice:  conv.FloatString(item.RequestPrice),
		RequestQty:    conv.FloatString(item.RequestQty),
		RequestAmount: conv.FloatString(item.RequestAmount),
		OperatorId:    item.OperatorId,
		Source:        trade.SourceType(item.Source),
		CheckSnapshot: conv.NullStringValue(item.CheckSnapshot),
		CreateTimes:   item.CreateTimes,
	}
}

func tradeEventToProto(item *models.TBizTradeEvent) *trade.BizTradeEvent {
	if item == nil {
		return nil
	}
	return &trade.BizTradeEvent{
		Id:             item.Id,
		TenantId:       item.TenantId,
		EventNo:        item.EventNo,
		EventType:      item.EventType,
		BizId:          item.BizId,
		BizType:        item.BizType,
		UserId:         item.UserId,
		SymbolId:       item.SymbolId,
		ProductType:    common.ProductType(item.ProductType),
		OperatorId:     item.OperatorId,
		Source:         trade.SourceType(item.Source),
		Consumer:       item.Consumer,
		EventStatus:    trade.EventStatus(item.EventStatus),
		RetryCount:     item.RetryCount,
		MaxRetryCount:  item.MaxRetryCount,
		NextRetryAt:    item.NextRetryAt,
		LastErrorMsg:   item.LastErrorMsg,
		PayloadVersion: item.PayloadVersion,
		ClaimedBy:      item.ClaimedBy,
		ClaimedAt:      item.ClaimedAt,
		DeliveredAt:    item.DeliveredAt,
		Payload:        item.Payload,
		ExtData:        conv.NullStringValue(item.ExtData),
		CreateTimes:    item.CreateTimes,
		UpdateTimes:    item.UpdateTimes,
	}
}

func ensureLeverage(leverage int64) int64 {
	if leverage <= 0 {
		return 1
	}
	return leverage
}

func ensureConfiguredLeverage(ctx context.Context, model models.TTradeSymbolLeverageConfigModel, defaultModel models.TTradeSymbolLeverageDefaultModel, tenantId int64, symbol *models.TTradeSymbol, marginMode trade.MarginMode, leverage int64) (int64, bool, error) {
	if symbol == nil || model == nil || marginMode == trade.MarginMode_MARGIN_MODE_UNKNOWN || !isDerivativeProduct(common.ProductType(symbol.ProductType)) {
		return ensureLeverage(leverage), true, nil
	}

	configs, _, err := model.FindPage(ctx, models.TradeSymbolLeverageConfigPageFilter{
		TenantId: tenantId, SymbolId: symbol.Id, MarginMode: int64(marginMode), Enabled: 1,
	}, 0, 200)
	if errors.Is(err, models.ErrNotFound) || len(configs) == 0 {
		return ensureLeverage(leverage), true, nil
	}
	if err != nil {
		return 0, false, err
	}

	effective := leverage
	if effective <= 0 {
		if defaultModel != nil {
			cfg, findErr := defaultModel.FindOneByTenantIdSymbolIdMarginMode(ctx, tenantId, symbol.Id, int64(marginMode))
			if findErr != nil && !errors.Is(findErr, models.ErrNotFound) {
				return 0, false, findErr
			}
			if cfg != nil {
				effective = cfg.Leverage
			}
		}
		if effective <= 0 {
			values := parseLeverageValues(configs[0].LeverageValues)
			if len(values) > 0 {
				effective = values[0]
			}
		}
	}
	for _, cfg := range configs {
		for _, configured := range parseLeverageValues(cfg.LeverageValues) {
			if configured == effective {
				return effective, true, nil
			}
		}
	}
	return 0, false, nil
}

func marginAssetForSymbol(symbol *models.TTradeSymbol) string {
	if symbol == nil {
		return ""
	}
	if symbol.SettleAsset != "" {
		return symbol.SettleAsset
	}
	if symbol.QuoteAsset != "" {
		return symbol.QuoteAsset
	}
	return symbol.BaseAsset
}

func orderCancelReason(operator string) string {
	if operator == "" {
		return "canceled"
	}
	return fmt.Sprintf("canceled by %s", operator)
}

const (
	immediateOrderExpireDelayMillis = int64(60 * 1000)
	freezingOrderRecoverDelayMillis = int64(60 * 1000)
)

func toTradeMinorAmount(amount decimal.Decimal) decimal.Decimal {
	return amount
}

func fromTradeMinorAmount(amount decimal.Decimal) decimal.Decimal {
	return amount
}

func tradeMinorAmountAtPrice(price, qty decimal.Decimal) decimal.Decimal {
	return toTradeMinorAmount(price.Mul(qty))
}

func tradeQtyFromMinorAmount(amount, price decimal.Decimal) decimal.Decimal {
	if !price.IsPositive() {
		return decimal.Zero
	}
	return fromTradeMinorAmount(amount).Div(price)
}

func openOrderStatuses() []int64 {
	return []int64{
		int64(trade.OrderStatus_ORDER_STATUS_PENDING),
		int64(trade.OrderStatus_ORDER_STATUS_PART_FILLED),
		int64(trade.OrderStatus_ORDER_STATUS_TRIGGER_WAITING),
	}
}

func isOpenOrderStatus(status int64) bool {
	switch trade.OrderStatus(status) {
	case trade.OrderStatus_ORDER_STATUS_PENDING,
		trade.OrderStatus_ORDER_STATUS_PART_FILLED,
		trade.OrderStatus_ORDER_STATUS_TRIGGER_WAITING:
		return true
	default:
		return false
	}
}

func matchableOrderStatuses() []int64 {
	return []int64{
		int64(trade.OrderStatus_ORDER_STATUS_PENDING),
		int64(trade.OrderStatus_ORDER_STATUS_PART_FILLED),
	}
}

func isMatchableOrderStatus(status int64) bool {
	switch trade.OrderStatus(status) {
	case trade.OrderStatus_ORDER_STATUS_PENDING, trade.OrderStatus_ORDER_STATUS_PART_FILLED:
		return true
	default:
		return false
	}
}

func triggerWaitingOrderStatuses() []int64 {
	return []int64{int64(trade.OrderStatus_ORDER_STATUS_TRIGGER_WAITING)}
}

func freezingOrderStatuses() []int64 {
	return []int64{int64(trade.OrderStatus_ORDER_STATUS_FREEZING)}
}

func terminatingOrderStatuses() []int64 {
	return []int64{int64(trade.OrderStatus_ORDER_STATUS_CANCELING), int64(trade.OrderStatus_ORDER_STATUS_EXPIRING)}
}

func isTriggerWaitingOrderStatus(status int64) bool {
	return trade.OrderStatus(status) == trade.OrderStatus_ORDER_STATUS_TRIGGER_WAITING
}

func isTerminalOrderStatus(status int64) bool {
	switch trade.OrderStatus(status) {
	case trade.OrderStatus_ORDER_STATUS_FILLED,
		trade.OrderStatus_ORDER_STATUS_CANCELED,
		trade.OrderStatus_ORDER_STATUS_REJECTED,
		trade.OrderStatus_ORDER_STATUS_EXPIRED:
		return true
	default:
		return false
	}
}

func orderStatusAfterFill(order *models.TTradeOrder) int64 {
	if order == nil {
		return int64(trade.OrderStatus_ORDER_STATUS_UNKNOWN)
	}
	if !order.FilledQty.IsPositive() && !order.FilledAmount.IsPositive() {
		return int64(trade.OrderStatus_ORDER_STATUS_PENDING)
	}
	if reachedFillTarget(order.FilledQty, order.Qty) {
		return int64(trade.OrderStatus_ORDER_STATUS_SETTLEMENT_PENDING)
	}
	if !order.Qty.IsPositive() && reachedFillTarget(order.FilledAmount, order.Amount) {
		return int64(trade.OrderStatus_ORDER_STATUS_SETTLEMENT_PENDING)
	}
	return int64(trade.OrderStatus_ORDER_STATUS_PART_FILLED)
}

func reachedFillTarget(filled, target decimal.Decimal) bool {
	return target.IsPositive() && filled.GreaterThanOrEqual(target)
}

func shouldExpireOrder(order *models.TTradeOrder, now int64) bool {
	if order == nil || !isMatchableOrderStatus(order.Status) {
		return false
	}
	switch trade.TimeInForce(order.TimeInForce) {
	case trade.TimeInForce_TIME_IN_FORCE_IOC, trade.TimeInForce_TIME_IN_FORCE_FOK:
	default:
		return false
	}
	activeAt := orderImmediateActiveAt(order)
	if activeAt <= 0 {
		return true
	}
	return now-activeAt >= immediateOrderExpireDelayMillis
}

func orderImmediateActiveAt(order *models.TTradeOrder) int64 {
	if order == nil {
		return 0
	}
	ext, err := parseOrderAssetExt(conv.NullStringValue(order.BizExt))
	if err == nil && ext.TriggeredAt > 0 {
		return ext.TriggeredAt
	}
	return order.CreateTimes
}

func shouldRecoverFreezingOrder(order *models.TTradeOrder, now int64) bool {
	if order == nil || trade.OrderStatus(order.Status) != trade.OrderStatus_ORDER_STATUS_FREEZING {
		return false
	}
	if order.CreateTimes <= 0 {
		return true
	}
	return now-order.CreateTimes >= freezingOrderRecoverDelayMillis
}

func orderExpireReason(order *models.TTradeOrder) string {
	switch trade.TimeInForce(order.TimeInForce) {
	case trade.TimeInForce_TIME_IN_FORCE_IOC:
		return "expired by IOC"
	case trade.TimeInForce_TIME_IN_FORCE_FOK:
		return "expired by FOK"
	default:
		return "expired"
	}
}

func marshalOrderAssetExt(ext orderAssetExt) (string, error) {
	if ext.FreezeNo == "" && ext.OriginalOrderType == 0 && ext.TriggeredAt == 0 && ext.TriggerPrice == "" && ext.TriggerSource == "" {
		return "", nil
	}
	buf, err := json.Marshal(ext)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func parseOrderAssetExt(raw string) (orderAssetExt, error) {
	if raw == "" {
		return orderAssetExt{}, nil
	}
	var ext orderAssetExt
	if err := json.Unmarshal([]byte(raw), &ext); err != nil {
		return orderAssetExt{}, err
	}
	return ext, nil
}

func spotFrozenAssetAndAmount(symbol *models.TTradeSymbol, side common.Side, qty, amount decimal.Decimal) (string, decimal.Decimal) {
	if symbol == nil {
		return "", decimal.Zero
	}
	if side == common.Side_SIDE_SELL {
		return symbol.BaseAsset, toTradeMinorAmount(qty)
	}
	return symbol.QuoteAsset, amount
}

const (
	legacyOrderTypeConditional = 3
	legacyOrderTypeTakeProfit  = 4
	legacyOrderTypeStopLoss    = 5
)

func normalizeOrderTypeAndTriggerKind(orderType trade.OrderType, triggerKind trade.TriggerKind, price decimal.Decimal) (trade.OrderType, trade.TriggerKind) {
	switch int32(orderType) {
	case legacyOrderTypeConditional:
		return executionOrderTypeFromPrice(price), trade.TriggerKind_TRIGGER_KIND_CONDITIONAL
	case legacyOrderTypeTakeProfit:
		return executionOrderTypeFromPrice(price), trade.TriggerKind_TRIGGER_KIND_TAKE_PROFIT
	case legacyOrderTypeStopLoss:
		return executionOrderTypeFromPrice(price), trade.TriggerKind_TRIGGER_KIND_STOP_LOSS
	default:
		return orderType, triggerKind
	}
}

func executionOrderTypeFromPrice(price decimal.Decimal) trade.OrderType {
	if price.IsPositive() {
		return trade.OrderType_ORDER_TYPE_LIMIT
	}
	return trade.OrderType_ORDER_TYPE_MARKET
}

func isTriggerKind(triggerKind trade.TriggerKind) bool {
	switch triggerKind {
	case trade.TriggerKind_TRIGGER_KIND_CONDITIONAL,
		trade.TriggerKind_TRIGGER_KIND_TAKE_PROFIT,
		trade.TriggerKind_TRIGGER_KIND_STOP_LOSS:
		return true
	default:
		return false
	}
}

func isSupportedOrderType(orderType trade.OrderType) bool {
	switch orderType {
	case trade.OrderType_ORDER_TYPE_LIMIT,
		trade.OrderType_ORDER_TYPE_MARKET:
		return true
	default:
		return false
	}
}

func isSupportedTriggerKind(triggerKind trade.TriggerKind) bool {
	switch triggerKind {
	case trade.TriggerKind_TRIGGER_KIND_NONE,
		trade.TriggerKind_TRIGGER_KIND_CONDITIONAL,
		trade.TriggerKind_TRIGGER_KIND_TAKE_PROFIT,
		trade.TriggerKind_TRIGGER_KIND_STOP_LOSS:
		return true
	default:
		return false
	}
}

func isValidOrderPrice(orderType trade.OrderType, price decimal.Decimal) bool {
	if orderType == trade.OrderType_ORDER_TYPE_LIMIT {
		return price.IsPositive()
	}
	return true
}

func hasNegativeOrderInput(price, qty, amount, triggerPrice decimal.Decimal) bool {
	return price.IsNegative() || qty.IsNegative() || amount.IsNegative() || triggerPrice.IsNegative()
}

func isValidOrderTimeInForce(orderType trade.OrderType, triggerKind trade.TriggerKind, timeInForce trade.TimeInForce) bool {
	if orderType == trade.OrderType_ORDER_TYPE_MARKET && timeInForce == trade.TimeInForce_TIME_IN_FORCE_POST_ONLY {
		return false
	}
	if isTriggerKind(triggerKind) && timeInForce == trade.TimeInForce_TIME_IN_FORCE_POST_ONLY {
		return false
	}
	return true
}

func normalizeOrderTimeInForce(orderType trade.OrderType, timeInForce trade.TimeInForce) trade.TimeInForce {
	switch orderType {
	case trade.OrderType_ORDER_TYPE_MARKET:
		if timeInForce == trade.TimeInForce_TIME_IN_FORCE_UNKNOWN ||
			timeInForce == trade.TimeInForce_TIME_IN_FORCE_GTC {
			return trade.TimeInForce_TIME_IN_FORCE_IOC
		}
	case trade.OrderType_ORDER_TYPE_LIMIT:
		if timeInForce == trade.TimeInForce_TIME_IN_FORCE_UNKNOWN {
			return trade.TimeInForce_TIME_IN_FORCE_GTC
		}
	}
	return timeInForce
}

func statusAfterFreeze(triggerKind trade.TriggerKind) int64 {
	if isTriggerKind(triggerKind) {
		return int64(trade.OrderStatus_ORDER_STATUS_TRIGGER_WAITING)
	}
	return int64(trade.OrderStatus_ORDER_STATUS_PENDING)
}

func triggeredOrderExecutionType(order *models.TTradeOrder) int64 {
	if order == nil {
		return int64(trade.OrderType_ORDER_TYPE_UNKNOWN)
	}
	if order.OrderType == int64(trade.OrderType_ORDER_TYPE_LIMIT) || order.OrderType == int64(trade.OrderType_ORDER_TYPE_MARKET) {
		return order.OrderType
	}
	if order.Price.IsPositive() {
		return int64(trade.OrderType_ORDER_TYPE_LIMIT)
	}
	return int64(trade.OrderType_ORDER_TYPE_MARKET)
}

func triggeredTimeInForce(order *models.TTradeOrder) int64 {
	if order == nil {
		return int64(trade.TimeInForce_TIME_IN_FORCE_UNKNOWN)
	}
	if triggeredOrderExecutionType(order) == int64(trade.OrderType_ORDER_TYPE_MARKET) {
		if order.TimeInForce == int64(trade.TimeInForce_TIME_IN_FORCE_UNKNOWN) ||
			order.TimeInForce == int64(trade.TimeInForce_TIME_IN_FORCE_GTC) ||
			order.TimeInForce == int64(trade.TimeInForce_TIME_IN_FORCE_POST_ONLY) {
			return int64(trade.TimeInForce_TIME_IN_FORCE_IOC)
		}
	}
	return order.TimeInForce
}

func triggerKindForOrder(order *models.TTradeOrder) trade.TriggerKind {
	if order == nil {
		return trade.TriggerKind_TRIGGER_KIND_NONE
	}
	if order.TriggerKind != 0 {
		return trade.TriggerKind(order.TriggerKind)
	}
	switch order.OrderType {
	case legacyOrderTypeConditional:
		return trade.TriggerKind_TRIGGER_KIND_CONDITIONAL
	case legacyOrderTypeTakeProfit:
		return trade.TriggerKind_TRIGGER_KIND_TAKE_PROFIT
	case legacyOrderTypeStopLoss:
		return trade.TriggerKind_TRIGGER_KIND_STOP_LOSS
	default:
		return trade.TriggerKind_TRIGGER_KIND_NONE
	}
}

func shouldTriggerOrder(order *models.TTradeOrder, triggerPrice decimal.Decimal) bool {
	if order == nil || !isTriggerWaitingOrderStatus(order.Status) || !order.TriggerPrice.IsPositive() || !triggerPrice.IsPositive() {
		return false
	}
	switch triggerKindForOrder(order) {
	case trade.TriggerKind_TRIGGER_KIND_TAKE_PROFIT:
		if order.Side == int64(common.Side_SIDE_BUY) {
			return triggerPrice.LessThanOrEqual(order.TriggerPrice)
		}
		return triggerPrice.GreaterThanOrEqual(order.TriggerPrice)
	case trade.TriggerKind_TRIGGER_KIND_STOP_LOSS:
		if order.Side == int64(common.Side_SIDE_BUY) {
			return triggerPrice.GreaterThanOrEqual(order.TriggerPrice)
		}
		return triggerPrice.LessThanOrEqual(order.TriggerPrice)
	case trade.TriggerKind_TRIGGER_KIND_CONDITIONAL:
		if order.Side == int64(common.Side_SIDE_BUY) {
			return triggerPrice.GreaterThanOrEqual(order.TriggerPrice)
		}
		return triggerPrice.LessThanOrEqual(order.TriggerPrice)
	default:
		return false
	}
}

func walletTypeForProduct(productType common.ProductType) common.WalletType {
	switch productType {
	case common.ProductType_PRODUCT_TYPE_SPOT:
		return common.WalletType_WALLET_TYPE_SPOT
	default:
		return common.WalletType_WALLET_TYPE_CONTRACT
	}
}

func adminTenantID(ctx context.Context, requested int64) int64 {
	// 管理端请求只能访问网关注入的当前租户，不能通过请求体切换租户。
	if tenantID, err := utils.GetTenantIdFromMd(ctx); err == nil && tenantID > 0 {
		return tenantID
	}
	return requested
}
