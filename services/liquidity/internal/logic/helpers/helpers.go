package helpers

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"wklive/proto/common"
	"wklive/proto/liquidity"
	"wklive/services/liquidity/models"
)

const (
	DefaultLimit = 20
	MaxLimit     = 200
)

func Limit(value int32) int64 {
	if value <= 0 {
		return DefaultLimit
	}
	if value > MaxLimit {
		return MaxLimit
	}
	return int64(value)
}

func TrimPage[T any](rows []*T, limit int64, id func(*T) int64) ([]*T, *liquidity.PageMeta) {
	hasMore := int64(len(rows)) > limit
	if hasMore {
		rows = rows[:limit]
	}
	var next int64
	if len(rows) > 0 {
		next = id(rows[len(rows)-1])
	}
	return rows, &liquidity.PageMeta{NextCursor: next, HasMore: hasMore}
}

func RequireID(name string, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%s is required", name)
	}
	return nil
}

func ProviderToProto(row *models.TLiquidityProvider) *liquidity.LiquidityProvider {
	if row == nil {
		return nil
	}
	return &liquidity.LiquidityProvider{
		Id:                   row.Id,
		ProviderCode:         row.ProviderCode,
		ProviderName:         row.ProviderName,
		ProviderType:         liquidity.ProviderType(row.ProviderType),
		TradeUserId:          row.TradeUserId,
		VenueCode:            row.VenueCode,
		Environment:          liquidity.ProviderEnvironment(row.Environment),
		AccountRef:           row.AccountRef,
		RateLimitPerSecond:   int32(row.RateLimitPerSecond),
		Status:               liquidity.ProviderStatus(row.Status),
		LastHealthStatus:     liquidity.HealthStatus(row.LastHealthStatus),
		LastHealthAt:         row.LastHealthAt,
		LastErrorMsg:         row.LastErrorMsg,
		Version:              row.Version,
		Remark:               row.Remark,
		CreateTimes:          row.CreateTimes,
		UpdateTimes:          row.UpdateTimes,
		CredentialConfigured: strings.TrimSpace(row.CredentialRef) != "",
	}
}

func SymbolConfigToProto(row *models.TLiquiditySymbolConfig) *liquidity.LiquiditySymbolConfig {
	if row == nil {
		return nil
	}
	f := func(v float64) string { return fmt.Sprintf("%.8f", v) }
	return &liquidity.LiquiditySymbolConfig{
		Id: row.Id, SymbolId: row.SymbolId, Symbol: row.Symbol,
		ProductType: common.ProductType(row.ProductType), ContractType: common.ContractType(row.ContractType),
		LiquidityMode:      liquidity.LiquidityMode(row.LiquidityMode),
		InternalProviderId: row.InternalProviderId, ExternalProviderId: row.ExternalProviderId,
		ExternalSymbol: row.ExternalSymbol, ReferencePriceSource: row.ReferencePriceSource,
		ReferencePriceKind: row.ReferencePriceKind, QuoteValidityMs: int32(row.QuoteValidityMs),
		RefreshIntervalMs: int32(row.RefreshIntervalMs), QuoteTtlMs: int32(row.QuoteTtlMs),
		RepriceThresholdBps: f(row.RepriceThresholdBps), BaseSpreadBps: f(row.BaseSpreadBps),
		MaxSpreadBps: f(row.MaxSpreadBps), MaxPriceDeviationBps: f(row.MaxPriceDeviationBps),
		PriceTick: f(row.PriceTick), QtyStep: f(row.QtyStep), MinQuoteQty: f(row.MinQuoteQty),
		MaxQuoteQty: f(row.MaxQuoteQty), MaxQuoteNotional: f(row.MaxQuoteNotional),
		TargetBaseInventory: f(row.TargetBaseInventory), MinBaseInventory: f(row.MinBaseInventory),
		MaxBaseInventory: f(row.MaxBaseInventory), MaxNetExposure: f(row.MaxNetExposure),
		MaxDailyNotional: f(row.MaxDailyNotional), InventorySkewBps: f(row.InventorySkewBps),
		HedgeThreshold: f(row.HedgeThreshold), HedgeRatio: f(row.HedgeRatio),
		SelfTradePrevention: common.YesNo(row.SelfTradePrevention),
		Status:              liquidity.SymbolLiquidityStatus(row.Status), PauseReason: row.PauseReason,
		Version: row.Version, CreateTimes: row.CreateTimes, UpdateTimes: row.UpdateTimes,
	}
}

func StrategyLevelToProto(row *models.TLiquidityStrategyLevel) *liquidity.LiquidityStrategyLevel {
	if row == nil {
		return nil
	}
	f := func(v float64) string { return fmt.Sprintf("%.8f", v) }
	return &liquidity.LiquidityStrategyLevel{
		Id: row.Id, ConfigId: row.ConfigId, LevelNo: int32(row.LevelNo),
		BidSpreadBps: f(row.BidSpreadBps), AskSpreadBps: f(row.AskSpreadBps),
		BidQty: f(row.BidQty), AskQty: f(row.AskQty), Enabled: common.Enable(row.Enabled),
		Version: row.Version, CreateTimes: row.CreateTimes, UpdateTimes: row.UpdateTimes,
	}
}

func QuoteCycleToProto(row *models.TLiquidityQuoteCycle) *liquidity.LiquidityQuoteCycle {
	if row == nil {
		return nil
	}
	return &liquidity.LiquidityQuoteCycle{
		Id: row.Id, CycleNo: row.CycleNo, ConfigId: row.ConfigId,
		SymbolId: row.SymbolId, ReferencePrice: number(row.ReferencePrice),
		ReferenceSource: row.ReferenceSource, ReferenceSnapshotId: row.ReferenceSnapshotId,
		ReferenceTime: row.ReferenceTime, TargetBidCount: int32(row.TargetBidCount),
		TargetAskCount: int32(row.TargetAskCount), PlacedBidCount: int32(row.PlacedBidCount),
		PlacedAskCount: int32(row.PlacedAskCount), Status: liquidity.QuoteCycleStatus(row.Status),
		LastErrorMsg: row.LastErrorMsg, StartedAt: row.StartedAt, FinishedAt: row.FinishedAt,
		CreateTimes: row.CreateTimes, UpdateTimes: row.UpdateTimes,
	}
}

func QuoteOrderToProto(row *models.TLiquidityQuoteOrder) *liquidity.LiquidityQuoteOrder {
	if row == nil {
		return nil
	}
	return &liquidity.LiquidityQuoteOrder{
		Id: row.Id, QuoteNo: row.QuoteNo, CycleId: row.CycleId,
		ConfigId: row.ConfigId, ProviderId: row.ProviderId, SymbolId: row.SymbolId,
		Side: common.Side(row.Side), LevelNo: int32(row.LevelNo), Price: number(row.Price),
		Qty: number(row.Qty), FilledQty: number(row.FilledQty), InternalOrderId: row.InternalOrderId,
		InternalOrderNo: row.InternalOrderNo, ClientOrderId: row.ClientOrderId,
		Status: liquidity.QuoteOrderStatus(row.Status), ExpireAt: row.ExpireAt,
		CancelReason: row.CancelReason, LastErrorMsg: row.LastErrorMsg, Version: row.Version,
		CreateTimes: row.CreateTimes, UpdateTimes: row.UpdateTimes,
	}
}

func ExternalOrderToProto(row *models.TLiquidityExternalOrder) *liquidity.LiquidityExternalOrder {
	if row == nil {
		return nil
	}
	return &liquidity.LiquidityExternalOrder{
		Id: row.Id, OrderNo: row.OrderNo, RequestNo: row.RequestNo,
		ProviderId: row.ProviderId, ConfigId: row.ConfigId, SymbolId: row.SymbolId,
		ExternalSymbol: row.ExternalSymbol, Purpose: liquidity.ExternalOrderPurpose(row.Purpose),
		ReferenceType: row.ReferenceType, ReferenceId: row.ReferenceId, Side: common.Side(row.Side),
		OrderType: liquidity.ExternalOrderType(row.OrderType), TimeInForce: liquidity.ExternalTimeInForce(row.TimeInForce),
		Price: number(row.Price), Qty: number(row.Qty), FilledQty: number(row.FilledQty),
		AvgPrice: number(row.AvgPrice), FeeAmount: number(row.FeeAmount), FeeAsset: row.FeeAsset,
		ExternalOrderId: nullString(row.ExternalOrderId), ExternalClientOrderId: row.ExternalClientOrderId,
		Status: liquidity.ExternalOrderStatus(row.Status), SubmittedAt: row.SubmittedAt,
		FinishedAt: row.FinishedAt, RetryCount: int32(row.RetryCount), NextRetryAt: row.NextRetryAt,
		LastErrorCode: row.LastErrorCode, LastErrorMsg: row.LastErrorMsg,
		RawResponse: nullString(row.RawResponse), Version: row.Version,
		CreateTimes: row.CreateTimes, UpdateTimes: row.UpdateTimes,
	}
}

func ExternalFillToProto(row *models.TLiquidityExternalFill) *liquidity.LiquidityExternalFill {
	if row == nil {
		return nil
	}
	return &liquidity.LiquidityExternalFill{
		Id: row.Id, ProviderId: row.ProviderId,
		ExternalOrderId: row.ExternalOrderId, FillNo: row.FillNo,
		ExternalTradeId: row.ExternalTradeId, Side: common.Side(row.Side),
		Price: number(row.Price), Qty: number(row.Qty), Amount: number(row.Amount),
		FeeAmount: number(row.FeeAmount), FeeAsset: row.FeeAsset,
		LiquidityType: int32(row.LiquidityType), TradeTime: row.TradeTime,
		SettlementStatus: liquidity.ExternalFillSettlementStatus(row.SettlementStatus),
		RetryCount:       int32(row.RetryCount), NextRetryAt: row.NextRetryAt,
		LastErrorMsg: row.LastErrorMsg, RawPayload: nullString(row.RawPayload),
		CreateTimes: row.CreateTimes, UpdateTimes: row.UpdateTimes,
	}
}

func HedgeTaskToProto(row *models.TLiquidityHedgeTask) *liquidity.LiquidityHedgeTask {
	if row == nil {
		return nil
	}
	return &liquidity.LiquidityHedgeTask{
		Id: row.Id, HedgeNo: row.HedgeNo, ConfigId: row.ConfigId,
		ProviderId: row.ProviderId, SymbolId: row.SymbolId,
		TriggerType:    liquidity.HedgeTriggerType(row.TriggerType),
		ExposureBefore: number(row.ExposureBefore), TargetExposure: number(row.TargetExposure),
		Side: common.Side(row.Side), TargetQty: number(row.TargetQty),
		ExecutedQty: number(row.ExecutedQty), AvgPrice: number(row.AvgPrice),
		Status: liquidity.HedgeStatus(row.Status), RetryCount: int32(row.RetryCount),
		NextRetryAt: row.NextRetryAt, LastErrorMsg: row.LastErrorMsg, Version: row.Version,
		CreateTimes: row.CreateTimes, UpdateTimes: row.UpdateTimes,
	}
}

func InventoryToProto(row *models.TLiquidityInventorySnapshot) *liquidity.LiquidityInventorySnapshot {
	if row == nil {
		return nil
	}
	return &liquidity.LiquidityInventorySnapshot{
		Id: row.Id, SnapshotNo: row.SnapshotNo,
		ConfigId: row.ConfigId, ProviderId: row.ProviderId, SymbolId: row.SymbolId,
		BaseAsset: row.BaseAsset, QuoteAsset: row.QuoteAsset,
		BaseTotal: number(row.BaseTotal), BaseAvailable: number(row.BaseAvailable),
		BaseFrozen: number(row.BaseFrozen), QuoteTotal: number(row.QuoteTotal),
		QuoteAvailable: number(row.QuoteAvailable), QuoteFrozen: number(row.QuoteFrozen),
		PositionQty: number(row.PositionQty), PendingBuyQty: number(row.PendingBuyQty),
		PendingSellQty: number(row.PendingSellQty), NetExposure: number(row.NetExposure),
		ReferencePrice: number(row.ReferencePrice), ExposureNotional: number(row.ExposureNotional),
		Source: liquidity.InventorySource(row.Source), SnapshotTime: row.SnapshotTime,
		RawPayload: nullString(row.RawPayload), CreateTimes: row.CreateTimes,
	}
}

func RiskEventToProto(row *models.TLiquidityRiskEvent) *liquidity.LiquidityRiskEvent {
	if row == nil {
		return nil
	}
	return &liquidity.LiquidityRiskEvent{
		Id: row.Id, EventNo: row.EventNo, ConfigId: row.ConfigId,
		ProviderId: row.ProviderId, SymbolId: row.SymbolId, RiskType: row.RiskType,
		RiskLevel: liquidity.RiskLevel(row.RiskLevel), MetricValue: number(row.MetricValue),
		ThresholdValue: number(row.ThresholdValue), ActionType: liquidity.RiskActionType(row.ActionType),
		Status: liquidity.RiskEventStatus(row.Status), Message: row.Message,
		ContextJson: nullString(row.ContextJson), TriggeredAt: row.TriggeredAt,
		RecoveredAt: row.RecoveredAt, ClosedAt: row.ClosedAt, OperatorId: row.OperatorId,
		CreateTimes: row.CreateTimes, UpdateTimes: row.UpdateTimes,
	}
}

func ReconcileBatchToProto(row *models.TLiquidityReconcileBatch) *liquidity.LiquidityReconcileBatch {
	if row == nil {
		return nil
	}
	return &liquidity.LiquidityReconcileBatch{
		Id: row.Id, BatchNo: row.BatchNo, ProviderId: row.ProviderId,
		ReconcileType: liquidity.ReconcileType(row.ReconcileType), WindowStart: row.WindowStart,
		WindowEnd: row.WindowEnd, LocalCount: row.LocalCount, ExternalCount: row.ExternalCount,
		MatchedCount: row.MatchedCount, DifferenceCount: row.DifferenceCount,
		Status: liquidity.ReconcileStatus(row.Status), LastErrorMsg: row.LastErrorMsg,
		StartedAt: row.StartedAt, FinishedAt: row.FinishedAt,
		CreateTimes: row.CreateTimes, UpdateTimes: row.UpdateTimes,
	}
}

func ReconcileDetailToProto(row *models.TLiquidityReconcileDetail) *liquidity.LiquidityReconcileDetail {
	if row == nil {
		return nil
	}
	return &liquidity.LiquidityReconcileDetail{
		Id: row.Id, BatchId: row.BatchId,
		DifferenceNo: row.DifferenceNo, DifferenceType: liquidity.ReconcileDifferenceType(row.DifferenceType),
		BusinessType: row.BusinessType, LocalReference: row.LocalReference,
		ExternalReference: row.ExternalReference, LocalValue: nullString(row.LocalValue),
		ExternalValue: nullString(row.ExternalValue),
		Status:        liquidity.ReconcileDifferenceStatus(row.Status), Resolution: row.Resolution,
		OperatorId: row.OperatorId, ResolvedAt: row.ResolvedAt,
		CreateTimes: row.CreateTimes, UpdateTimes: row.UpdateTimes,
	}
}

func number(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func nullString(value sql.NullString) string {
	if !value.Valid {
		return ""
	}
	return value.String
}
