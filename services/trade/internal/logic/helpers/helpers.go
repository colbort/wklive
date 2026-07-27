package helpers

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

func MustParseFloat(v string) decimal.Decimal {
	value, _ := conv.ParseDecimalField(v)
	return value
}

func EnableToProto(value int64) common.Enable {
	return common.Enable(value)
}

func EnableToModel(value common.Enable, defaultValue int64) int64 {
	if value == common.Enable_ENABLE_UNKNOWN {
		return defaultValue
	}
	return int64(value)
}

func YesNoToModel(value common.YesNo, defaultValue int64) int64 {
	if value == common.YesNo_YES_NO_UNKNOWN {
		return defaultValue
	}
	return int64(value)
}

type OrderAssetExt struct {
	FreezeNo          string `json:"freezeNo,omitempty"`
	OriginalOrderType int64  `json:"originalOrderType,omitempty"`
	TriggeredAt       int64  `json:"triggeredAt,omitempty"`
	TriggerPrice      string `json:"triggerPrice,omitempty"`
	TriggerSource     string `json:"triggerSource,omitempty"`
}

func SymbolToProto(item *models.TTradeSymbol) *trade.TradeSymbol {
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

func SpotSymbolToProto(item *models.TTradeSymbolSpot) *trade.TradeSymbolSpot {
	if item == nil {
		return nil
	}
	return &trade.TradeSymbolSpot{
		Id:           item.Id,
		TenantId:     item.TenantId,
		SymbolId:     item.SymbolId,
		MakerFeeRate: conv.FloatString(item.MakerFeeRate),
		TakerFeeRate: conv.FloatString(item.TakerFeeRate),
		BuyEnabled:   EnableToProto(item.BuyEnabled),
		SellEnabled:  EnableToProto(item.SellEnabled),
		CreateTimes:  item.CreateTimes,
		UpdateTimes:  item.UpdateTimes,
	}
}

func ContractSymbolToProto(item *models.TTradeSymbolContract) *trade.TradeSymbolContract {
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
		OpenLongEnabled:          EnableToProto(item.OpenLongEnabled),
		OpenShortEnabled:         EnableToProto(item.OpenShortEnabled),
		CloseLongEnabled:         EnableToProto(item.CloseLongEnabled),
		CloseShortEnabled:        EnableToProto(item.CloseShortEnabled),
		CreateTimes:              item.CreateTimes,
		UpdateTimes:              item.UpdateTimes,
	}
}

func SecondsSymbolToProto(item *models.TTradeSymbolSeconds) *trade.TradeSymbolSeconds {
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
		UpEnabled: EnableToProto(item.UpEnabled), DownEnabled: EnableToProto(item.DownEnabled),
		CreateTimes: item.CreateTimes, UpdateTimes: item.UpdateTimes,
	}
}

func SymbolSessionToProto(item *models.TTradeSymbolSession) *trade.TradeSymbolSession {
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
		Enabled:     EnableToProto(item.Enabled),
	}
}

func UserConfigToProto(item *models.TTradeUserConfig) *trade.TradeUserConfig {
	if item == nil {
		return nil
	}
	return &trade.TradeUserConfig{
		Id:           item.Id,
		TenantId:     item.TenantId,
		UserId:       item.UserId,
		ProductType:  common.ProductType(item.ProductType),
		SymbolId:     item.SymbolId,
		TradeEnabled: EnableToProto(item.TradeEnabled),
		CreateTimes:  item.CreateTimes,
		UpdateTimes:  item.UpdateTimes,
	}
}

func ContractUserConfigToProto(item *models.TContractUserConfig) *trade.ContractUserConfig {
	if item == nil {
		return nil
	}
	return &trade.ContractUserConfig{Id: item.Id, TenantId: item.TenantId, UserId: item.UserId, SymbolId: item.SymbolId, PositionMode: trade.PositionMode(item.PositionMode), MarginMode: trade.MarginMode(item.MarginMode), DefaultLeverage: item.DefaultLeverage, CreateTimes: item.CreateTimes, UpdateTimes: item.UpdateTimes}
}

func OrderToProto(item *models.TTradeOrder) *trade.TradeOrder {
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
		DisplayStatus:     OrderDisplayStatus(item),
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

func OrderDisplayStatus(order *models.TTradeOrder) trade.OrderDisplayStatus {
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

func SecondsOrderDisplayStatus(status int64) trade.OrderDisplayStatus {
	switch trade.SecondsSettlementStatus(status) {
	case trade.SecondsSettlementStatus_SECONDS_SETTLEMENT_STATUS_PENDING_FREEZE:
		return trade.OrderDisplayStatus_ORDER_DISPLAY_STATUS_FREEZING
	case trade.SecondsSettlementStatus_SECONDS_SETTLEMENT_STATUS_ACTIVATING:
		return trade.OrderDisplayStatus_ORDER_DISPLAY_STATUS_ACTIVATING
	case trade.SecondsSettlementStatus_SECONDS_SETTLEMENT_STATUS_ACTIVE:
		return trade.OrderDisplayStatus_ORDER_DISPLAY_STATUS_ACTIVE
	case trade.SecondsSettlementStatus_SECONDS_SETTLEMENT_STATUS_SETTLING:
		return trade.OrderDisplayStatus_ORDER_DISPLAY_STATUS_SETTLING
	case trade.SecondsSettlementStatus_SECONDS_SETTLEMENT_STATUS_SETTLED:
		return trade.OrderDisplayStatus_ORDER_DISPLAY_STATUS_SETTLED
	case trade.SecondsSettlementStatus_SECONDS_SETTLEMENT_STATUS_REFUNDING:
		return trade.OrderDisplayStatus_ORDER_DISPLAY_STATUS_REFUNDING
	case trade.SecondsSettlementStatus_SECONDS_SETTLEMENT_STATUS_REFUNDED:
		return trade.OrderDisplayStatus_ORDER_DISPLAY_STATUS_REFUNDED
	case trade.SecondsSettlementStatus_SECONDS_SETTLEMENT_STATUS_MANUAL_REVIEW:
		return trade.OrderDisplayStatus_ORDER_DISPLAY_STATUS_MANUAL_REVIEW
	default:
		return trade.OrderDisplayStatus_ORDER_DISPLAY_STATUS_UNKNOWN
	}
}

func OrderSpotToProto(item *models.TTradeOrderSpot) *trade.TradeOrderSpot {
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

func OrderContractToProto(item *models.TTradeOrderContract) *trade.TradeOrderContract {
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

func OrderSecondsToProto(item *models.TTradeOrderSeconds) *trade.TradeOrderSeconds {
	if item == nil {
		return nil
	}
	return &trade.TradeOrderSeconds{
		Id: item.Id, TenantId: item.TenantId, OrderId: item.OrderId,
		Direction: trade.SecondsDirection(item.Direction), DurationSeconds: item.DurationSeconds,
		StakeAsset: item.StakeAsset, StakeAmount: conv.FloatString(item.StakeAmount),
		PayoutRate: conv.FloatString(item.PayoutRate), FeeRate: conv.FloatString(item.FeeRate),
		FrozenAt: item.FrozenAt, ActivatedAt: item.ActivatedAt,
		StartPrice: conv.FloatString(item.StartPrice), StartPriceTime: item.StartPriceTime,
		StartPriceSource: item.StartPriceSource, ExpireTime: item.ExpireTime,
		SettlementPrice: conv.FloatString(item.SettlementPrice), SettlementPriceTime: item.SettlementPriceTime,
		SettlementPriceSource: item.SettlementPriceSource, PriceAlgorithm: item.PriceAlgorithm,
		Result: trade.SecondsResult(item.Result), ProfitAmount: conv.FloatString(item.ProfitAmount),
		FeeAmount: conv.FloatString(item.FeeAmount), ReturnAmount: conv.FloatString(item.ReturnAmount),
		SettlementStatus: trade.SecondsSettlementStatus(item.SettlementStatus),
		ReservationNo:    item.ReservationNo, SettlementReason: item.SettlementReason,
		SettledAt: item.SettledAt, Version: item.Version,
		CreateTimes: item.CreateTimes, UpdateTimes: item.UpdateTimes,
	}
}

func FillToProto(item *models.TTradeFill) *trade.TradeFill {
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

func CancelLogToProto(item *models.TTradeCancelLog) *trade.TradeCancelLog {
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

func PositionToProto(item *models.TContractPosition) *trade.ContractPosition {
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

func PositionHistoryToProto(item *models.TContractPositionHistory) *trade.ContractPositionHistory {
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

func MarginSnapshotToProto(item *models.TContractMarginSnapshot) *trade.ContractMarginSnapshot {
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

func LeverageConfigToProto(item *models.TContractLeverageConfig) *trade.ContractLeverageConfig {
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
		Enabled:       EnableToProto(item.Enabled),
		Remark:        item.Remark,
		CreateTimes:   item.CreateTimes,
		UpdateTimes:   item.UpdateTimes,
	}
}

func SymbolLeverageConfigToProto(item *models.TTradeSymbolLeverageConfig, defaultLeverage ...int64) *trade.TradeSymbolLeverageConfig {
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
		LeverageValues:  ParseLeverageValues(item.LeverageValues),
		DefaultLeverage: defaultValue,
		Enabled:         EnableToProto(item.Enabled),
		Sort:            item.Sort,
		Remark:          item.Remark,
		CreateTimes:     item.CreateTimes,
		UpdateTimes:     item.UpdateTimes,
	}
}

func FindDefaultLeverage(ctx context.Context, model models.TTradeSymbolLeverageDefaultModel, tenantId, symbolId, marginMode int64) (int64, error) {
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

func ParseLeverageValues(value string) []int64 {
	var jsonValues []int64
	if json.Unmarshal([]byte(value), &jsonValues) == nil {
		return NormalizeLeverageValues(jsonValues)
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
	return NormalizeLeverageValues(values)
}

func NormalizeLeverageValues(values []int64) []int64 {
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

func JoinLeverageValues(values []int64, maxLeverage int64) (string, []int64) {
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

func ContainsLeverage(values []int64, leverage int64) bool {
	for _, value := range values {
		if value == leverage {
			return true
		}
	}
	return false
}

func IsDerivativeProduct(productType common.ProductType) bool {
	return productType == common.ProductType_PRODUCT_TYPE_DERIVATIVE
}

func RiskUserTradeLimitToProto(item *models.TRiskUserTradeLimit) *trade.RiskUserTradeLimit {
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
		TradeEnabled:         EnableToProto(item.TradeEnabled),
		OnlyReduceOnly:       common.Enable(item.OnlyReduceOnly),
		MaxOpenOrderCount:    item.MaxOpenOrderCount,
		MaxOrderCountPerDay:  item.MaxOrderCountPerDay,
		MaxCancelCountPerDay: item.MaxCancelCountPerDay,
		MaxOpenNotional:      conv.FloatString(item.MaxOpenNotional),
		MaxPositionNotional:  conv.FloatString(item.MaxPositionNotional),
		RiskLevel:            trade.RiskLevel(item.RiskLevel),
		OperatorId:           item.OperatorId,
		Source:               trade.SourceType(item.Source),
		Enabled:              EnableToProto(item.Enabled),
		EffectiveStartTime:   item.EffectiveStartTime,
		EffectiveEndTime:     item.EffectiveEndTime,
		Remark:               item.Remark,
		CreateTimes:          item.CreateTimes,
		UpdateTimes:          item.UpdateTimes,
	}
}

func RiskUserSymbolLimitToProto(item *models.TRiskUserSymbolLimit) *trade.RiskUserSymbolLimit {
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
		Enabled:             EnableToProto(item.Enabled),
		EffectiveStartTime:  item.EffectiveStartTime,
		EffectiveEndTime:    item.EffectiveEndTime,
		Remark:              item.Remark,
		CreateTimes:         item.CreateTimes,
		UpdateTimes:         item.UpdateTimes,
	}
}

func RiskOrderCheckLogToProto(item *models.TRiskOrderCheckLog) *trade.RiskOrderCheckLog {
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

func TradeEventToProto(item *models.TBizTradeEvent) *trade.BizTradeEvent {
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

func EnsureLeverage(leverage int64) int64 {
	if leverage <= 0 {
		return 1
	}
	return leverage
}

func EnsureConfiguredLeverage(ctx context.Context, model models.TTradeSymbolLeverageConfigModel, defaultModel models.TTradeSymbolLeverageDefaultModel, tenantId int64, symbol *models.TTradeSymbol, marginMode trade.MarginMode, leverage int64) (int64, bool, error) {
	if symbol == nil || model == nil || marginMode == trade.MarginMode_MARGIN_MODE_UNKNOWN || !IsDerivativeProduct(common.ProductType(symbol.ProductType)) {
		return EnsureLeverage(leverage), true, nil
	}

	configs, _, err := model.FindPage(ctx, models.TradeSymbolLeverageConfigPageFilter{
		TenantId: tenantId, SymbolId: symbol.Id, MarginMode: int64(marginMode), Enabled: 1,
	}, 0, 200)
	if errors.Is(err, models.ErrNotFound) || len(configs) == 0 {
		return EnsureLeverage(leverage), true, nil
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
			values := ParseLeverageValues(configs[0].LeverageValues)
			if len(values) > 0 {
				effective = values[0]
			}
		}
	}
	for _, cfg := range configs {
		for _, configured := range ParseLeverageValues(cfg.LeverageValues) {
			if configured == effective {
				return effective, true, nil
			}
		}
	}
	return 0, false, nil
}

func MarginAssetForSymbol(symbol *models.TTradeSymbol) string {
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

func OrderCancelReason(operator string) string {
	if operator == "" {
		return "canceled"
	}
	return fmt.Sprintf("canceled by %s", operator)
}

const (
	ImmediateOrderExpireDelayMillis = int64(60 * 1000)
	FreezingOrderRecoverDelayMillis = int64(60 * 1000)
)

func ToTradeMinorAmount(amount decimal.Decimal) decimal.Decimal {
	return amount
}

func FromTradeMinorAmount(amount decimal.Decimal) decimal.Decimal {
	return amount
}

func TradeMinorAmountAtPrice(price, qty decimal.Decimal) decimal.Decimal {
	return ToTradeMinorAmount(price.Mul(qty))
}

func TradeQtyFromMinorAmount(amount, price decimal.Decimal) decimal.Decimal {
	if !price.IsPositive() {
		return decimal.Zero
	}
	return FromTradeMinorAmount(amount).Div(price)
}

func OpenOrderStatuses() []int64 {
	return []int64{
		int64(trade.OrderStatus_ORDER_STATUS_PENDING),
		int64(trade.OrderStatus_ORDER_STATUS_PART_FILLED),
		int64(trade.OrderStatus_ORDER_STATUS_TRIGGER_WAITING),
	}
}

func IsOpenOrderStatus(status int64) bool {
	switch trade.OrderStatus(status) {
	case trade.OrderStatus_ORDER_STATUS_PENDING,
		trade.OrderStatus_ORDER_STATUS_PART_FILLED,
		trade.OrderStatus_ORDER_STATUS_TRIGGER_WAITING:
		return true
	default:
		return false
	}
}

func MatchableOrderStatuses() []int64 {
	return []int64{
		int64(trade.OrderStatus_ORDER_STATUS_PENDING),
		int64(trade.OrderStatus_ORDER_STATUS_PART_FILLED),
	}
}

func IsMatchableOrderStatus(status int64) bool {
	switch trade.OrderStatus(status) {
	case trade.OrderStatus_ORDER_STATUS_PENDING, trade.OrderStatus_ORDER_STATUS_PART_FILLED:
		return true
	default:
		return false
	}
}

func TriggerWaitingOrderStatuses() []int64 {
	return []int64{int64(trade.OrderStatus_ORDER_STATUS_TRIGGER_WAITING)}
}

func FreezingOrderStatuses() []int64 {
	return []int64{int64(trade.OrderStatus_ORDER_STATUS_FREEZING)}
}

func TerminatingOrderStatuses() []int64 {
	return []int64{int64(trade.OrderStatus_ORDER_STATUS_CANCELING), int64(trade.OrderStatus_ORDER_STATUS_EXPIRING)}
}

func IsTriggerWaitingOrderStatus(status int64) bool {
	return trade.OrderStatus(status) == trade.OrderStatus_ORDER_STATUS_TRIGGER_WAITING
}

func IsTerminalOrderStatus(status int64) bool {
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

func OrderStatusAfterFill(order *models.TTradeOrder) int64 {
	if order == nil {
		return int64(trade.OrderStatus_ORDER_STATUS_UNKNOWN)
	}
	if !order.FilledQty.IsPositive() && !order.FilledAmount.IsPositive() {
		return int64(trade.OrderStatus_ORDER_STATUS_PENDING)
	}
	if OrderFillTargetReached(order) {
		return int64(trade.OrderStatus_ORDER_STATUS_SETTLEMENT_PENDING)
	}
	return int64(trade.OrderStatus_ORDER_STATUS_PART_FILLED)
}

func OrderFillTargetReached(order *models.TTradeOrder) bool {
	if order == nil {
		return false
	}
	if order.Qty.IsPositive() {
		return ReachedFillTarget(order.FilledQty, order.Qty)
	}
	return ReachedFillTarget(order.FilledAmount, order.Amount)
}

func ReachedFillTarget(filled, target decimal.Decimal) bool {
	if !target.IsPositive() {
		return false
	}
	if filled.GreaterThanOrEqual(target) {
		return true
	}
	// Amount-based spot orders calculate quantity by dividing amount by the
	// execution price. Sub-minor-unit arithmetic dust counts as fully filled.
	return target.Sub(filled).LessThanOrEqual(decimal.RequireFromString("0.00000001"))
}

func ShouldExpireOrder(order *models.TTradeOrder, now int64) bool {
	if order == nil || !IsMatchableOrderStatus(order.Status) {
		return false
	}
	switch trade.TimeInForce(order.TimeInForce) {
	case trade.TimeInForce_TIME_IN_FORCE_IOC, trade.TimeInForce_TIME_IN_FORCE_FOK:
	default:
		return false
	}
	activeAt := OrderImmediateActiveAt(order)
	if activeAt <= 0 {
		return true
	}
	return now-activeAt >= ImmediateOrderExpireDelayMillis
}

func OrderImmediateActiveAt(order *models.TTradeOrder) int64 {
	if order == nil {
		return 0
	}
	ext, err := ParseOrderAssetExt(conv.NullStringValue(order.BizExt))
	if err == nil && ext.TriggeredAt > 0 {
		return ext.TriggeredAt
	}
	return order.CreateTimes
}

func ShouldRecoverFreezingOrder(order *models.TTradeOrder, now int64) bool {
	if order == nil || trade.OrderStatus(order.Status) != trade.OrderStatus_ORDER_STATUS_FREEZING {
		return false
	}
	if order.CreateTimes <= 0 {
		return true
	}
	return now-order.CreateTimes >= FreezingOrderRecoverDelayMillis
}

func OrderExpireReason(order *models.TTradeOrder) string {
	switch trade.TimeInForce(order.TimeInForce) {
	case trade.TimeInForce_TIME_IN_FORCE_IOC:
		return "expired by IOC"
	case trade.TimeInForce_TIME_IN_FORCE_FOK:
		return "expired by FOK"
	default:
		return "expired"
	}
}

func MarshalOrderAssetExt(ext OrderAssetExt) (string, error) {
	if ext.FreezeNo == "" && ext.OriginalOrderType == 0 && ext.TriggeredAt == 0 && ext.TriggerPrice == "" && ext.TriggerSource == "" {
		return "", nil
	}
	buf, err := json.Marshal(ext)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func ParseOrderAssetExt(raw string) (OrderAssetExt, error) {
	if raw == "" {
		return OrderAssetExt{}, nil
	}
	var ext OrderAssetExt
	if err := json.Unmarshal([]byte(raw), &ext); err != nil {
		return OrderAssetExt{}, err
	}
	return ext, nil
}

func SpotFrozenAssetAndAmount(symbol *models.TTradeSymbol, side common.Side, qty, amount decimal.Decimal) (string, decimal.Decimal) {
	if symbol == nil {
		return "", decimal.Zero
	}
	if side == common.Side_SIDE_SELL {
		return symbol.BaseAsset, ToTradeMinorAmount(qty)
	}
	return symbol.QuoteAsset, amount
}

const (
	LegacyOrderTypeConditional = 3
	LegacyOrderTypeTakeProfit  = 4
	LegacyOrderTypeStopLoss    = 5
)

func NormalizeOrderTypeAndTriggerKind(orderType trade.OrderType, triggerKind trade.TriggerKind, price decimal.Decimal) (trade.OrderType, trade.TriggerKind) {
	switch int32(orderType) {
	case LegacyOrderTypeConditional:
		return ExecutionOrderTypeFromPrice(price), trade.TriggerKind_TRIGGER_KIND_CONDITIONAL
	case LegacyOrderTypeTakeProfit:
		return ExecutionOrderTypeFromPrice(price), trade.TriggerKind_TRIGGER_KIND_TAKE_PROFIT
	case LegacyOrderTypeStopLoss:
		return ExecutionOrderTypeFromPrice(price), trade.TriggerKind_TRIGGER_KIND_STOP_LOSS
	default:
		return orderType, triggerKind
	}
}

func ExecutionOrderTypeFromPrice(price decimal.Decimal) trade.OrderType {
	if price.IsPositive() {
		return trade.OrderType_ORDER_TYPE_LIMIT
	}
	return trade.OrderType_ORDER_TYPE_MARKET
}

func IsTriggerKind(triggerKind trade.TriggerKind) bool {
	switch triggerKind {
	case trade.TriggerKind_TRIGGER_KIND_CONDITIONAL,
		trade.TriggerKind_TRIGGER_KIND_TAKE_PROFIT,
		trade.TriggerKind_TRIGGER_KIND_STOP_LOSS:
		return true
	default:
		return false
	}
}

func IsSupportedOrderType(orderType trade.OrderType) bool {
	switch orderType {
	case trade.OrderType_ORDER_TYPE_LIMIT,
		trade.OrderType_ORDER_TYPE_MARKET:
		return true
	default:
		return false
	}
}

func IsSupportedTriggerKind(triggerKind trade.TriggerKind) bool {
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

func IsValidOrderPrice(orderType trade.OrderType, price decimal.Decimal) bool {
	if orderType == trade.OrderType_ORDER_TYPE_LIMIT {
		return price.IsPositive()
	}
	return true
}

func HasNegativeOrderInput(price, qty, amount, triggerPrice decimal.Decimal) bool {
	return price.IsNegative() || qty.IsNegative() || amount.IsNegative() || triggerPrice.IsNegative()
}

func IsValidOrderTimeInForce(orderType trade.OrderType, triggerKind trade.TriggerKind, timeInForce trade.TimeInForce) bool {
	if orderType == trade.OrderType_ORDER_TYPE_MARKET && timeInForce == trade.TimeInForce_TIME_IN_FORCE_POST_ONLY {
		return false
	}
	if IsTriggerKind(triggerKind) && timeInForce == trade.TimeInForce_TIME_IN_FORCE_POST_ONLY {
		return false
	}
	return true
}

func NormalizeOrderTimeInForce(orderType trade.OrderType, timeInForce trade.TimeInForce) trade.TimeInForce {
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

func StatusAfterFreeze(triggerKind trade.TriggerKind) int64 {
	if IsTriggerKind(triggerKind) {
		return int64(trade.OrderStatus_ORDER_STATUS_TRIGGER_WAITING)
	}
	return int64(trade.OrderStatus_ORDER_STATUS_PENDING)
}

func TriggeredOrderExecutionType(order *models.TTradeOrder) int64 {
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

func TriggeredTimeInForce(order *models.TTradeOrder) int64 {
	if order == nil {
		return int64(trade.TimeInForce_TIME_IN_FORCE_UNKNOWN)
	}
	if TriggeredOrderExecutionType(order) == int64(trade.OrderType_ORDER_TYPE_MARKET) {
		if order.TimeInForce == int64(trade.TimeInForce_TIME_IN_FORCE_UNKNOWN) ||
			order.TimeInForce == int64(trade.TimeInForce_TIME_IN_FORCE_GTC) ||
			order.TimeInForce == int64(trade.TimeInForce_TIME_IN_FORCE_POST_ONLY) {
			return int64(trade.TimeInForce_TIME_IN_FORCE_IOC)
		}
	}
	return order.TimeInForce
}

func TriggerKindForOrder(order *models.TTradeOrder) trade.TriggerKind {
	if order == nil {
		return trade.TriggerKind_TRIGGER_KIND_NONE
	}
	if order.TriggerKind != 0 {
		return trade.TriggerKind(order.TriggerKind)
	}
	switch order.OrderType {
	case LegacyOrderTypeConditional:
		return trade.TriggerKind_TRIGGER_KIND_CONDITIONAL
	case LegacyOrderTypeTakeProfit:
		return trade.TriggerKind_TRIGGER_KIND_TAKE_PROFIT
	case LegacyOrderTypeStopLoss:
		return trade.TriggerKind_TRIGGER_KIND_STOP_LOSS
	default:
		return trade.TriggerKind_TRIGGER_KIND_NONE
	}
}

func ShouldTriggerOrder(order *models.TTradeOrder, triggerPrice decimal.Decimal) bool {
	if order == nil || !IsTriggerWaitingOrderStatus(order.Status) || !order.TriggerPrice.IsPositive() || !triggerPrice.IsPositive() {
		return false
	}
	switch TriggerKindForOrder(order) {
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

func WalletTypeForProduct(productType common.ProductType) common.WalletType {
	switch productType {
	case common.ProductType_PRODUCT_TYPE_SPOT:
		return common.WalletType_WALLET_TYPE_SPOT
	default:
		return common.WalletType_WALLET_TYPE_CONTRACT
	}
}

func AdminTenantID(ctx context.Context, requested int64) int64 {
	// 管理端请求只能访问网关注入的当前租户，不能通过请求体切换租户。
	if tenantID, err := utils.GetTenantIdFromMd(ctx); err == nil && tenantID > 0 {
		return tenantID
	}
	return requested
}
