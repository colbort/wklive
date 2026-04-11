package logic

import (
	"encoding/json"
	"fmt"

	"wklive/common/conv"
	"wklive/proto/asset"
	"wklive/proto/trade"
	"wklive/services/trade/models"
)

func mustParseFloat(v string) float64 {
	value, _ := conv.ParseFloatField(v)
	return value
}

type orderAssetExt struct {
	FreezeNo string `json:"freezeNo,omitempty"`
}

func symbolToProto(item *models.TTradeSymbol) *trade.TradeSymbol {
	if item == nil {
		return nil
	}
	return &trade.TradeSymbol{
		Id:            item.Id,
		TenantId:      item.TenantId,
		Symbol:        item.Symbol,
		DisplaySymbol: item.DisplaySymbol,
		MarketType:    trade.MarketType(item.MarketType),
		BaseAsset:     item.BaseAsset,
		QuoteAsset:    item.QuoteAsset,
		SettleAsset:   item.SettleAsset,
		ContractType:  trade.ContractType(item.ContractType),
		Status:        trade.SymbolStatus(item.Status),
		PriceScale:    uint32(item.PriceScale),
		QtyScale:      uint32(item.QtyScale),
		MinPrice:      conv.FloatString(item.MinPrice),
		MaxPrice:      conv.FloatString(item.MaxPrice),
		PriceTick:     conv.FloatString(item.PriceTick),
		MinQty:        conv.FloatString(item.MinQty),
		MaxQty:        conv.FloatString(item.MaxQty),
		QtyStep:       conv.FloatString(item.QtyStep),
		MinNotional:   conv.FloatString(item.MinNotional),
		MaxLeverage:   uint32(item.MaxLeverage),
		OpenTime:      item.OpenTime,
		CloseTime:     item.CloseTime,
		Sort:          int32(item.Sort),
		Remark:        item.Remark,
		CreateTimes:   item.CreateTimes,
		UpdateTimes:   item.UpdateTimes,
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
		BuyEnabled:   conv.Int64Bool(item.BuyEnabled),
		SellEnabled:  conv.Int64Bool(item.SellEnabled),
		CreateTimes:  item.CreateTimes,
		UpdateTimes:  item.UpdateTimes,
	}
}

func contractSymbolToProto(item *models.TTradeSymbolContract) *trade.TradeSymbolContract {
	if item == nil {
		return nil
	}
	return &trade.TradeSymbolContract{
		Id:                     item.Id,
		TenantId:               item.TenantId,
		SymbolId:               item.SymbolId,
		ContractSize:           conv.FloatString(item.ContractSize),
		Multiplier:             conv.FloatString(item.Multiplier),
		MaintenanceMarginRate:  conv.FloatString(item.MaintenanceMarginRate),
		InitialMarginRate:      conv.FloatString(item.InitialMarginRate),
		MakerFeeRate:           conv.FloatString(item.MakerFeeRate),
		TakerFeeRate:           conv.FloatString(item.TakerFeeRate),
		FundingIntervalMinutes: uint32(item.FundingIntervalMinutes),
		DeliveryTime:           item.DeliveryTime,
		SupportCross:           conv.Int64Bool(item.SupportCross),
		SupportIsolated:        conv.Int64Bool(item.SupportIsolated),
		BuyEnabled:             conv.Int64Bool(item.BuyEnabled),
		SellEnabled:            conv.Int64Bool(item.SellEnabled),
		CreateTimes:            item.CreateTimes,
		UpdateTimes:            item.UpdateTimes,
	}
}

func userConfigToProto(item *models.TTradeUserConfig) *trade.TradeUserConfig {
	if item == nil {
		return nil
	}
	return &trade.TradeUserConfig{
		Id:                item.Id,
		TenantId:          item.TenantId,
		UserId:            item.UserId,
		MarketType:        trade.MarketType(item.MarketType),
		SymbolId:          item.SymbolId,
		PositionMode:      trade.PositionMode(item.PositionMode),
		MarginMode:        trade.MarginMode(item.MarginMode),
		DefaultLeverage:   uint32(item.DefaultLeverage),
		TradeEnabled:      conv.Int64Bool(item.TradeEnabled),
		ReduceOnlyEnabled: conv.Int64Bool(item.ReduceOnlyEnabled),
		CreateTimes:       item.CreateTimes,
		UpdateTimes:       item.UpdateTimes,
	}
}

func orderToProto(item *models.TTradeOrder) *trade.TradeOrder {
	if item == nil {
		return nil
	}
	return &trade.TradeOrder{
		Id:            item.Id,
		TenantId:      item.TenantId,
		OrderNo:       item.OrderNo,
		ClientOrderId: item.ClientOrderId,
		UserId:        item.UserId,
		SymbolId:      item.SymbolId,
		MarketType:    trade.MarketType(item.MarketType),
		Side:          trade.TradeSide(item.Side),
		PositionSide:  trade.PositionSide(item.PositionSide),
		OrderType:     trade.OrderType(item.OrderType),
		TimeInForce:   trade.TimeInForce(item.TimeInForce),
		Status:        trade.OrderStatus(item.Status),
		Price:         conv.FloatString(item.Price),
		Qty:           conv.FloatString(item.Qty),
		Amount:        conv.FloatString(item.Amount),
		FilledQty:     conv.FloatString(item.FilledQty),
		FilledAmount:  conv.FloatString(item.FilledAmount),
		AvgPrice:      conv.FloatString(item.AvgPrice),
		Fee:           conv.FloatString(item.Fee),
		FeeAsset:      item.FeeAsset,
		Source:        trade.OrderSourceType(item.Source),
		IsReduceOnly:  conv.Int64Bool(item.IsReduceOnly),
		IsCloseOnly:   conv.Int64Bool(item.IsCloseOnly),
		TriggerPrice:  conv.FloatString(item.TriggerPrice),
		TriggerType:   uint32(item.TriggerType),
		CancelReason:  item.CancelReason,
		BizExt:        conv.NullStringValue(item.BizExt),
		CreateTimes:   item.CreateTimes,
		UpdateTimes:   item.UpdateTimes,
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
		Leverage:          uint32(item.Leverage),
		MarginAsset:       item.MarginAsset,
		MarginAmount:      conv.FloatString(item.MarginAmount),
		ClosePositionType: uint32(item.ClosePositionType),
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
		Id:            item.Id,
		TenantId:      item.TenantId,
		FillNo:        item.FillNo,
		OrderId:       item.OrderId,
		OrderNo:       item.OrderNo,
		UserId:        item.UserId,
		SymbolId:      item.SymbolId,
		MarketType:    trade.MarketType(item.MarketType),
		Side:          trade.TradeSide(item.Side),
		PositionSide:  trade.PositionSide(item.PositionSide),
		Price:         conv.FloatString(item.Price),
		Qty:           conv.FloatString(item.Qty),
		Amount:        conv.FloatString(item.Amount),
		Fee:           conv.FloatString(item.Fee),
		FeeAsset:      item.FeeAsset,
		LiquidityType: trade.LiquidityType(item.LiquidityType),
		RealizedPnl:   conv.FloatString(item.RealizedPnl),
		MatchTime:     item.MatchTime,
		CreateTimes:   item.CreateTimes,
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
		CancelSource: uint32(item.CancelSource),
		CancelReason: item.CancelReason,
		CreateTimes:  item.CreateTimes,
	}
}

func positionToProto(item *models.TContractPosition) *trade.ContractPosition {
	if item == nil {
		return nil
	}
	return &trade.ContractPosition{
		Id:               item.Id,
		TenantId:         item.TenantId,
		UserId:           item.UserId,
		SymbolId:         item.SymbolId,
		MarketType:       trade.MarketType(item.MarketType),
		PositionSide:     trade.PositionSide(item.PositionSide),
		MarginMode:       trade.MarginMode(item.MarginMode),
		Leverage:         uint32(item.Leverage),
		Qty:              conv.FloatString(item.Qty),
		AvailQty:         conv.FloatString(item.AvailQty),
		FrozenQty:        conv.FloatString(item.FrozenQty),
		OpenAvgPrice:     conv.FloatString(item.OpenAvgPrice),
		MarkPrice:        conv.FloatString(item.MarkPrice),
		MarginAsset:      item.MarginAsset,
		PositionMargin:   conv.FloatString(item.PositionMargin),
		IsolatedMargin:   conv.FloatString(item.IsolatedMargin),
		UnrealizedPnl:    conv.FloatString(item.UnrealizedPnl),
		RealizedPnl:      conv.FloatString(item.RealizedPnl),
		LiquidationPrice: conv.FloatString(item.LiquidationPrice),
		AdlRank:          int32(item.AdlRank),
		Version:          item.Version,
		CreateTimes:      item.CreateTimes,
		UpdateTimes:      item.UpdateTimes,
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
		MarketType:           trade.MarketType(item.MarketType),
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

func marginAccountToProto(item *models.TContractMarginAccount) *trade.ContractMarginAccount {
	if item == nil {
		return nil
	}
	return &trade.ContractMarginAccount{
		Id:               item.Id,
		TenantId:         item.TenantId,
		UserId:           item.UserId,
		MarketType:       trade.MarketType(item.MarketType),
		MarginAsset:      item.MarginAsset,
		Balance:          conv.FloatString(item.Balance),
		AvailableBalance: conv.FloatString(item.AvailableBalance),
		FrozenBalance:    conv.FloatString(item.FrozenBalance),
		PositionMargin:   conv.FloatString(item.PositionMargin),
		OrderMargin:      conv.FloatString(item.OrderMargin),
		UnrealizedPnl:    conv.FloatString(item.UnrealizedPnl),
		RealizedPnl:      conv.FloatString(item.RealizedPnl),
		Version:          item.Version,
		CreateTimes:      item.CreateTimes,
		UpdateTimes:      item.UpdateTimes,
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
		MarketType:    trade.MarketType(item.MarketType),
		MarginMode:    trade.MarginMode(item.MarginMode),
		PositionMode:  trade.PositionMode(item.PositionMode),
		LongLeverage:  uint32(item.LongLeverage),
		ShortLeverage: uint32(item.ShortLeverage),
		MaxLeverage:   uint32(item.MaxLeverage),
		OperatorId:    item.OperatorId,
		Source:        trade.SourceType(item.Source),
		Status:        uint32(item.Status),
		Remark:        item.Remark,
		CreateTimes:   item.CreateTimes,
		UpdateTimes:   item.UpdateTimes,
	}
}

func riskUserTradeLimitToProto(item *models.TRiskUserTradeLimit) *trade.RiskUserTradeLimit {
	if item == nil {
		return nil
	}
	return &trade.RiskUserTradeLimit{
		Id:                   item.Id,
		TenantId:             item.TenantId,
		UserId:               item.UserId,
		MarketType:           trade.MarketType(item.MarketType),
		CanOpen:              conv.Int64Bool(item.CanOpen),
		CanClose:             conv.Int64Bool(item.CanClose),
		CanCancel:            conv.Int64Bool(item.CanCancel),
		CanTriggerOrder:      conv.Int64Bool(item.CanTriggerOrder),
		CanApiTrade:          conv.Int64Bool(item.CanApiTrade),
		TradeEnabled:         conv.Int64Bool(item.TradeEnabled),
		OnlyReduceOnly:       conv.Int64Bool(item.OnlyReduceOnly),
		MaxOpenOrderCount:    uint32(item.MaxOpenOrderCount),
		MaxOrderCountPerDay:  uint32(item.MaxOrderCountPerDay),
		MaxCancelCountPerDay: uint32(item.MaxCancelCountPerDay),
		MaxOpenNotional:      conv.FloatString(item.MaxOpenNotional),
		MaxPositionNotional:  conv.FloatString(item.MaxPositionNotional),
		RiskLevel:            uint32(item.RiskLevel),
		OperatorId:           item.OperatorId,
		Source:               trade.SourceType(item.Source),
		Status:               uint32(item.Status),
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
		MarketType:          trade.MarketType(item.MarketType),
		MaxPositionQty:      conv.FloatString(item.MaxPositionQty),
		MaxPositionNotional: conv.FloatString(item.MaxPositionNotional),
		MaxOpenOrders:       uint32(item.MaxOpenOrders),
		MaxOrderQty:         conv.FloatString(item.MaxOrderQty),
		MaxOrderNotional:    conv.FloatString(item.MaxOrderNotional),
		MinOrderQty:         conv.FloatString(item.MinOrderQty),
		MinOrderNotional:    conv.FloatString(item.MinOrderNotional),
		MaxLongPositionQty:  conv.FloatString(item.MaxLongPositionQty),
		MaxShortPositionQty: conv.FloatString(item.MaxShortPositionQty),
		PriceDeviationRate:  conv.FloatString(item.PriceDeviationRate),
		OperatorId:          item.OperatorId,
		Source:              trade.SourceType(item.Source),
		Status:              uint32(item.Status),
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
		MarketType:    trade.MarketType(item.MarketType),
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
		Id:            item.Id,
		TenantId:      item.TenantId,
		EventNo:       item.EventNo,
		EventType:     item.EventType,
		BizId:         item.BizId,
		BizType:       item.BizType,
		UserId:        item.UserId,
		SymbolId:      item.SymbolId,
		MarketType:    trade.MarketType(item.MarketType),
		OperatorId:    item.OperatorId,
		Source:        trade.SourceType(item.Source),
		EventStatus:   trade.EventStatus(item.EventStatus),
		RetryCount:    uint32(item.RetryCount),
		MaxRetryCount: uint32(item.MaxRetryCount),
		NextRetryAt:   item.NextRetryAt,
		LastErrorMsg:  item.LastErrorMsg,
		Payload:       item.Payload,
		ExtData:       conv.NullStringValue(item.ExtData),
		CreateTimes:   item.CreateTimes,
		UpdateTimes:   item.UpdateTimes,
	}
}

func ensureLeverage(symbol *models.TTradeSymbol, leverage uint32) uint32 {
	if leverage == 0 {
		if symbol != nil && symbol.MaxLeverage > 0 {
			return uint32(symbol.MaxLeverage)
		}
		return 1
	}
	if symbol != nil && symbol.MaxLeverage > 0 && leverage > uint32(symbol.MaxLeverage) {
		return uint32(symbol.MaxLeverage)
	}
	return leverage
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

func openOrderStatuses() []int64 {
	return []int64{
		int64(trade.OrderStatus_ORDER_STATUS_PENDING),
		int64(trade.OrderStatus_ORDER_STATUS_PART_FILLED),
	}
}

func marshalOrderAssetExt(ext orderAssetExt) (string, error) {
	if ext.FreezeNo == "" {
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

func spotFrozenAssetAndAmount(symbol *models.TTradeSymbol, side trade.TradeSide, qty, amount float64) (string, float64) {
	if symbol == nil {
		return "", 0
	}
	if side == trade.TradeSide_TRADE_SIDE_SELL {
		return symbol.BaseAsset, qty
	}
	return symbol.QuoteAsset, amount
}

func walletTypeForMarket(marketType trade.MarketType) asset.WalletType {
	switch marketType {
	case trade.MarketType_MARKET_TYPE_SPOT:
		return asset.WalletType_WALLET_TYPE_SPOT
	default:
		return asset.WalletType_WALLET_TYPE_CONTRACT
	}
}
