package helpers

import (
	"context"
	"errors"

	"wklive/common/conv"
	"wklive/proto/common"
	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"
)

func FindContractByCodeOrID(ctx context.Context, svcCtx *svc.ServiceContext, tenantId, id int64, contractCode string) (*models.TOptionContract, error) {
	if id != 0 {
		item, err := svcCtx.OptionContractModel.FindOne(ctx, id)
		if err != nil {
			return nil, err
		}
		if tenantId != 0 && item.TenantId != tenantId {
			return nil, models.ErrNotFound
		}
		return item, nil
	}
	return svcCtx.OptionContractModel.FindOneByTenantIdContractCode(ctx, tenantId, contractCode)
}

func FindOrderByNoOrID(ctx context.Context, svcCtx *svc.ServiceContext, tenantId, id int64, orderNo string) (*models.TOptionOrder, error) {
	if id != 0 {
		item, err := svcCtx.OptionOrderModel.FindOne(ctx, id)
		if err != nil {
			return nil, err
		}
		if tenantId != 0 && item.TenantId != tenantId {
			return nil, models.ErrNotFound
		}
		return item, nil
	}
	return svcCtx.OptionOrderModel.FindOneByTenantIdOrderNo(ctx, tenantId, orderNo)
}

func FindTradeByNoOrID(ctx context.Context, svcCtx *svc.ServiceContext, tenantId, id int64, tradeNo string) (*models.TOptionTrade, error) {
	if id != 0 {
		item, err := svcCtx.OptionTradeModel.FindOne(ctx, id)
		if err != nil {
			return nil, err
		}
		if tenantId != 0 && item.TenantId != tenantId {
			return nil, models.ErrNotFound
		}
		return item, nil
	}
	return svcCtx.OptionTradeModel.FindOneByTenantIdTradeNo(ctx, tenantId, tradeNo)
}

func FindExerciseByNoOrID(ctx context.Context, svcCtx *svc.ServiceContext, tenantId, id int64, exerciseNo string) (*models.TOptionExercise, error) {
	if id != 0 {
		item, err := svcCtx.OptionExerciseModel.FindOne(ctx, id)
		if err != nil {
			return nil, err
		}
		if tenantId != 0 && item.TenantId != tenantId {
			return nil, models.ErrNotFound
		}
		return item, nil
	}
	return svcCtx.OptionExerciseModel.FindOneByTenantIdExerciseNo(ctx, tenantId, exerciseNo)
}

func FindSettlementByNoOrID(ctx context.Context, svcCtx *svc.ServiceContext, tenantId, id int64, settlementNo string) (*models.TOptionSettlement, error) {
	if id != 0 {
		item, err := svcCtx.OptionSettlementModel.FindOne(ctx, id)
		if err != nil {
			return nil, err
		}
		if tenantId != 0 && item.TenantId != tenantId {
			return nil, models.ErrNotFound
		}
		return item, nil
	}
	return svcCtx.OptionSettlementModel.FindOneByTenantIdSettlementNo(ctx, tenantId, settlementNo)
}

func FindContractIgnoreNotFound(ctx context.Context, svcCtx *svc.ServiceContext, tenantId, contractId int64) (*models.TOptionContract, error) {
	item, err := svcCtx.OptionContractModel.FindOne(ctx, contractId)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	if tenantId != 0 && item.TenantId != tenantId {
		return nil, nil
	}
	return item, nil
}

func FindMarketIgnoreNotFound(ctx context.Context, svcCtx *svc.ServiceContext, tenantId, contractId int64) (*models.TOptionMarket, error) {
	item, err := svcCtx.OptionMarketModel.FindOneByTenantIdContractId(ctx, tenantId, contractId)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return item, nil
}

func ToContractProto(item *models.TOptionContract) *option.OptionContract {
	if item == nil {
		return nil
	}
	return &option.OptionContract{
		Id:                      item.Id,
		TenantId:                item.TenantId,
		ContractCode:            item.ContractCode,
		UnderlyingSymbol:        item.UnderlyingSymbol,
		UnderlyingCoin:          item.UnderlyingCoin,
		SettleCoin:              item.SettleCoin,
		QuoteCoin:               item.QuoteCoin,
		OptionType:              option.OptionType(item.OptionType),
		ExerciseStyle:           option.ExerciseStyle(item.ExerciseStyle),
		SettlementType:          option.SettlementType(item.SettlementType),
		StrikePrice:             conv.FloatString(item.StrikePrice),
		ContractUnit:            conv.FloatString(item.ContractUnit),
		MinOrderQty:             conv.FloatString(item.MinOrderQty),
		MaxOrderQty:             conv.FloatString(item.MaxOrderQty),
		PriceTick:               conv.FloatString(item.PriceTick),
		QtyStep:                 conv.FloatString(item.QtyStep),
		Multiplier:              conv.FloatString(item.Multiplier),
		ListTime:                item.ListTime,
		ExpireTime:              item.ExpireTime,
		DeliverTime:             item.DeliverTime,
		TradingCalendarCode:     item.TradingCalendarCode,
		ExerciseCutoffTime:      item.ExerciseCutoffTime,
		AutoExerciseThreshold:   conv.FloatString(item.AutoExerciseThreshold),
		MaxUserLongQty:          conv.FloatString(item.MaxUserLongQty),
		MaxUserShortQty:         conv.FloatString(item.MaxUserShortQty),
		MaxOpenInterest:         conv.FloatString(item.MaxOpenInterest),
		OrderPriceBandRatio:     conv.FloatString(item.OrderPriceBandRatio),
		CircuitBreakerRatio:     conv.FloatString(item.CircuitBreakerRatio),
		GreeksMaxAgeSeconds:     item.GreeksMaxAgeSeconds,
		SettlementPriceSource:   item.SettlementPriceSource,
		SettlementPriceMethod:   item.SettlementPriceMethod,
		SettlementWindowSeconds: item.SettlementWindowSeconds,
		SettlementMinSamples:    item.SettlementMinSamples,
		IsAutoExercise:          common.YesNo(item.IsAutoExercise),
		Status:                  option.ContractStatus(item.Status),
		Sort:                    item.Sort,
		Remark:                  item.Remark,
		IsDeleted:               common.YesNo(item.IsDeleted),
		CreateTimes:             item.CreateTimes,
		UpdateTimes:             item.UpdateTimes,
		MakerFeeRate:            conv.FloatString(item.MakerFeeRate),
		TakerFeeRate:            conv.FloatString(item.TakerFeeRate),
		ExerciseFeeRate:         conv.FloatString(item.ExerciseFeeRate),
		FeeUserId:               item.FeeUserId,
		FeeAccountId:            item.FeeAccountId,
		SellerMarginMode:        option.SellerMarginMode(item.SellerMarginMode),
		InitialMarginRate:       conv.FloatString(item.InitialMarginRate),
		MaintenanceMarginRate:   conv.FloatString(item.MaintenanceMarginRate),
		MinMarginRate:           conv.FloatString(item.MinMarginRate),
		LiquidationFeeRate:      conv.FloatString(item.LiquidationFeeRate),
		InsuranceUserId:         item.InsuranceUserId,
		InsuranceAccountId:      item.InsuranceAccountId,
		LiquidationDeficitPolicy: option.LiquidationDeficitPolicy(
			item.LiquidationDeficitPolicy,
		),
		PhysicalDeliveryPolicy:      option.PhysicalDeliveryPolicy(item.PhysicalDeliveryPolicy),
		PhysicalDeliveryCureSeconds: item.PhysicalDeliveryCureSeconds,
	}
}

func ToTradingCalendarProto(
	item *models.TOptionTradingCalendar,
	sessions []*models.TOptionTradingCalendarSession,
	exceptions []*models.TOptionTradingCalendarException,
) *option.OptionTradingCalendar {
	if item == nil {
		return nil
	}
	result := &option.OptionTradingCalendar{
		Id: item.Id, TenantId: item.TenantId, CalendarCode: item.CalendarCode,
		Version: item.Version, Status: option.TradingCalendarStatus(item.Status),
		Timezone: item.Timezone, EffectiveFrom: item.EffectiveFrom, EffectiveUntil: item.EffectiveUntil,
		SupersedesId: item.SupersedesId, ChangeReason: item.ChangeReason, EvidenceRef: item.EvidenceRef,
		CreatedBy: item.CreatedBy, ReviewedBy: item.ReviewedBy, ReviewReason: item.ReviewReason,
		ReviewedAt: item.ReviewedAt, CreateTimes: item.CreateTimes, UpdateTimes: item.UpdateTimes,
		Sessions:   make([]*option.OptionTradingCalendarSession, 0, len(sessions)),
		Exceptions: make([]*option.OptionTradingCalendarException, 0, len(exceptions)),
	}
	for _, session := range sessions {
		result.Sessions = append(result.Sessions, &option.OptionTradingCalendarSession{
			Id: session.Id, TenantId: session.TenantId, CalendarId: session.CalendarId,
			Weekday: int32(session.Weekday), OpenSecond: int32(session.OpenSecond),
			CloseSecond: int32(session.CloseSecond), CreateTimes: session.CreateTimes,
		})
	}
	for _, exception := range exceptions {
		result.Exceptions = append(result.Exceptions, &option.OptionTradingCalendarException{
			Id: exception.Id, TenantId: exception.TenantId, CalendarId: exception.CalendarId,
			ExceptionType: option.TradingCalendarExceptionType(exception.ExceptionType),
			StartTime:     exception.StartTime, EndTime: exception.EndTime, Reason: exception.Reason,
			AnnouncementRef: exception.AnnouncementRef, CreateTimes: exception.CreateTimes,
		})
	}
	return result
}

func ToTradingHaltProto(item *models.TOptionTradingHalt) *option.OptionTradingHalt {
	if item == nil {
		return nil
	}
	return &option.OptionTradingHalt{
		Id: item.Id, TenantId: item.TenantId, HaltNo: item.HaltNo, ContractId: item.ContractId,
		Source: option.TradingHaltSource(item.Source), Status: option.TradingHaltStatus(item.Status),
		Reason: item.Reason, EvidenceRef: item.EvidenceRef, StartedAt: item.StartedAt,
		CreatedBy: item.CreatedBy, CancelTotal: item.CancelTotal, CancelSuccess: item.CancelSuccess,
		CancelFailed: item.CancelFailed, LastErrorMsg: item.LastErrorMsg, LiftedAt: item.LiftedAt,
		LiftedBy: item.LiftedBy, LiftReason: item.LiftReason, CreateTimes: item.CreateTimes,
		UpdateTimes: item.UpdateTimes,
	}
}

func ToMarketProto(item *models.TOptionMarket) *option.OptionMarket {
	if item == nil {
		return nil
	}
	return &option.OptionMarket{
		Id:                     item.Id,
		TenantId:               item.TenantId,
		ContractId:             item.ContractId,
		UnderlyingPrice:        conv.FloatString(item.UnderlyingPrice),
		MarkPrice:              conv.FloatString(item.MarkPrice),
		LastPrice:              conv.FloatString(item.LastPrice),
		BidPrice:               conv.FloatString(item.BidPrice),
		AskPrice:               conv.FloatString(item.AskPrice),
		TheoreticalPrice:       conv.FloatString(item.TheoreticalPrice),
		IntrinsicValue:         conv.FloatString(item.IntrinsicValue),
		TimeValue:              conv.FloatString(item.TimeValue),
		Iv:                     conv.FloatString(item.Iv),
		Delta:                  conv.FloatString(item.Delta),
		Gamma:                  conv.FloatString(item.Gamma),
		Theta:                  conv.FloatString(item.Theta),
		Vega:                   conv.FloatString(item.Vega),
		Rho:                    conv.FloatString(item.Rho),
		RiskFreeRate:           conv.FloatString(item.RiskFreeRate),
		PricingModel:           item.PricingModel,
		SnapshotTime:           item.SnapshotTime,
		UnderlyingSnapshotTime: item.UnderlyingSnapshotTime,
		MarkSnapshotTime:       item.MarkSnapshotTime,
		GreeksSnapshotTime:     item.GreeksSnapshotTime,
		CreateTimes:            item.CreateTimes,
		UpdateTimes:            item.UpdateTimes,
	}
}

func ToUserTradingControlProto(item *models.TOptionUserTradingControl) *option.OptionUserTradingControl {
	if item == nil {
		return nil
	}
	return &option.OptionUserTradingControl{
		TenantId: item.TenantId, UserId: item.UserId,
		KillSwitch: common.YesNo(item.KillSwitch), Reason: item.Reason,
		ActivatedAt: item.ActivatedAt, ReleasedAt: item.ReleasedAt, UpdateTimes: item.UpdateTimes,
	}
}

func ToTradingControlEventProto(item *models.TOptionTradingControlEvent) *option.OptionTradingControlEvent {
	if item == nil {
		return nil
	}
	return &option.OptionTradingControlEvent{
		Id: item.Id, TenantId: item.TenantId, UserId: item.UserId,
		ContractId: item.ContractId, OrderId: item.OrderId,
		EventType: item.EventType, Reason: item.Reason, Detail: item.Detail,
		OperatorId: item.OperatorId, CreateTimes: item.CreateTimes,
	}
}

func ToMMPConfigProto(item *models.TOptionMmpConfig) *option.OptionMMPConfig {
	if item == nil {
		return nil
	}
	return &option.OptionMMPConfig{
		Id: item.Id, TenantId: item.TenantId, UserId: item.UserId,
		ContractId: item.ContractId, GroupCode: item.GroupCode,
		Enabled:             common.YesNo(item.Enabled),
		QtyThreshold:        conv.FloatString(item.QtyThreshold),
		TradeCountThreshold: item.TradeCountThreshold,
		LossThreshold:       conv.FloatString(item.LossThreshold),
		WindowSeconds:       item.WindowSeconds, CooldownSeconds: item.CooldownSeconds,
		Status: option.MMPStatus(item.Status), WindowStart: item.WindowStart,
		AccumulatedQty:  conv.FloatString(item.AccumulatedQty),
		TradeCount:      item.TradeCount,
		AccumulatedLoss: conv.FloatString(item.AccumulatedLoss),
		TriggeredAt:     item.TriggeredAt, CooldownUntil: item.CooldownUntil,
		TriggerReason: item.TriggerReason, LastErrorMsg: item.LastErrorMsg,
		CreatedBy: item.CreatedBy, UpdatedBy: item.UpdatedBy,
		CreateTimes: item.CreateTimes, UpdateTimes: item.UpdateTimes,
	}
}

func ToPortfolioRiskConfigProto(item *models.TOptionPortfolioRiskConfig) *option.OptionPortfolioRiskConfig {
	if item == nil {
		return nil
	}
	return &option.OptionPortfolioRiskConfig{
		Id: item.Id, TenantId: item.TenantId, SettleCoin: item.SettleCoin,
		Version: item.Version, Status: option.PortfolioRiskConfigStatus(item.Status),
		ModelMethod:            option.PortfolioRiskMethod(item.ModelMethod),
		InitialShockRate:       conv.FloatString(item.InitialShockRate),
		MaintenanceShockRate:   conv.FloatString(item.MaintenanceShockRate),
		ScenarioShocks:         item.ScenarioShocks,
		ConcentrationThreshold: conv.FloatString(item.ConcentrationThreshold),
		ConcentrationAddonRate: conv.FloatString(item.ConcentrationAddonRate),
		LiquidityAddonRate:     conv.FloatString(item.LiquidityAddonRate),
		EffectiveFrom:          item.EffectiveFrom,
		EffectiveUntil:         item.EffectiveUntil,
		SupersedesId:           item.SupersedesId,
		ChangeReason:           item.ChangeReason,
		EvidenceRef:            item.EvidenceRef,
		CreatedBy:              item.CreatedBy,
		ReviewedBy:             item.ReviewedBy,
		ReviewReason:           item.ReviewReason,
		ReviewedAt:             item.ReviewedAt,
		CreateTimes:            item.CreateTimes,
		UpdateTimes:            item.UpdateTimes,
	}
}

func ToTradeCorrectionProto(
	item *models.TOptionTradeCorrection,
	legs []*models.TOptionTradeCorrectionLeg,
) *option.OptionTradeCorrection {
	if item == nil {
		return nil
	}
	result := &option.OptionTradeCorrection{
		Id: item.Id, TenantId: item.TenantId, CaseNo: item.CaseNo,
		TradeId: item.TradeId, ContractId: item.ContractId,
		Action: option.TradeCorrectionAction(item.Action),
		Status: option.TradeCorrectionStatus(item.Status),
		Reason: item.Reason, EvidenceRef: item.EvidenceRef,
		RequestedBy: item.RequestedBy, ReviewedBy: item.ReviewedBy,
		ReviewReason: item.ReviewReason, ReviewedAt: item.ReviewedAt,
		CompletedAt: item.CompletedAt, LastErrorMsg: item.LastErrorMsg,
		CreateTimes: item.CreateTimes, UpdateTimes: item.UpdateTimes,
		Legs: make([]*option.OptionTradeCorrectionLeg, 0, len(legs)),
	}
	for _, leg := range legs {
		result.Legs = append(result.Legs, &option.OptionTradeCorrectionLeg{
			Id: leg.Id, LegNo: leg.LegNo, UserId: leg.UserId, AccountId: leg.AccountId,
			Coin: leg.Coin, Direction: option.TradeCorrectionLegDirection(leg.Direction),
			Amount: conv.FloatString(leg.Amount), InstructionNo: leg.InstructionNo,
		})
	}
	return result
}

func ToMarketSnapshotProto(item *models.TOptionMarketSnapshot) *option.OptionMarketSnapshot {
	if item == nil {
		return nil
	}
	return &option.OptionMarketSnapshot{
		Id:               item.Id,
		TenantId:         item.TenantId,
		ContractId:       item.ContractId,
		UnderlyingPrice:  conv.FloatString(item.UnderlyingPrice),
		MarkPrice:        conv.FloatString(item.MarkPrice),
		LastPrice:        conv.FloatString(item.LastPrice),
		BidPrice:         conv.FloatString(item.BidPrice),
		AskPrice:         conv.FloatString(item.AskPrice),
		TheoreticalPrice: conv.FloatString(item.TheoreticalPrice),
		Iv:               conv.FloatString(item.Iv),
		Delta:            conv.FloatString(item.Delta),
		Gamma:            conv.FloatString(item.Gamma),
		Theta:            conv.FloatString(item.Theta),
		Vega:             conv.FloatString(item.Vega),
		Rho:              conv.FloatString(item.Rho),
		SnapshotTime:     item.SnapshotTime,
		SourceType:       item.SourceType,
		SourceSnapshotId: item.SourceSnapshotId,
		CreateTimes:      item.CreateTimes,
	}
}

const (
	MarketSnapshotSourceUnknown    int64 = 0
	MarketSnapshotSourceUnderlying int64 = 1
	MarketSnapshotSourceAdmin      int64 = 2
	MarketSnapshotSourceSettlement int64 = 3
)

func InsertMarketSnapshot(
	ctx context.Context,
	model models.TOptionMarketSnapshotModel,
	market *models.TOptionMarket,
	sourceType int64,
	sourceSnapshotID string,
	now int64,
) error {
	if market == nil {
		return nil
	}
	snapshotTime := market.SnapshotTime
	if snapshotTime == 0 {
		snapshotTime = now
	}
	_, err := model.Insert(ctx, &models.TOptionMarketSnapshot{
		TenantId:         market.TenantId,
		ContractId:       market.ContractId,
		UnderlyingPrice:  market.UnderlyingPrice,
		MarkPrice:        market.MarkPrice,
		LastPrice:        market.LastPrice,
		BidPrice:         market.BidPrice,
		AskPrice:         market.AskPrice,
		TheoreticalPrice: market.TheoreticalPrice,
		Iv:               market.Iv,
		Delta:            market.Delta,
		Gamma:            market.Gamma,
		Theta:            market.Theta,
		Vega:             market.Vega,
		Rho:              market.Rho,
		SnapshotTime:     snapshotTime,
		SourceType:       sourceType,
		SourceSnapshotId: sourceSnapshotID,
		CreateTimes:      now,
	})
	return err
}

func ToOrderProto(item *models.TOptionOrder) *option.OptionOrder {
	if item == nil {
		return nil
	}
	return &option.OptionOrder{
		Id:                         item.Id,
		TenantId:                   item.TenantId,
		OrderNo:                    item.OrderNo,
		UserId:                     item.UserId,
		AccountId:                  item.AccountId,
		ContractId:                 item.ContractId,
		UnderlyingSymbol:           item.UnderlyingSymbol,
		Side:                       common.Side(item.Side),
		PositionEffect:             option.PositionEffect(item.PositionEffect),
		OrderType:                  option.OrderType(item.OrderType),
		Price:                      conv.FloatString(item.Price),
		Qty:                        conv.FloatString(item.Qty),
		FilledQty:                  conv.FloatString(item.FilledQty),
		UnfilledQty:                conv.FloatString(item.UnfilledQty),
		AvgPrice:                   conv.FloatString(item.AvgPrice),
		Turnover:                   conv.FloatString(item.Turnover),
		Fee:                        conv.FloatString(item.Fee),
		FeeCoin:                    item.FeeCoin,
		MarginAmount:               conv.FloatString(item.MarginAmount),
		MarginCoin:                 item.MarginCoin,
		Source:                     option.OrderSource(item.Source),
		ClientOrderId:              item.ClientOrderId,
		ReduceOnly:                 common.YesNo(item.ReduceOnly),
		Mmp:                        common.YesNo(item.Mmp),
		MmpGroup:                   item.MmpGroup,
		ComboOrderId:               item.ComboOrderId,
		ComboLegNo:                 item.ComboLegNo,
		PortfolioRiskConfigId:      item.PortfolioRiskConfigId,
		PortfolioRiskConfigVersion: item.PortfolioRiskConfigVersion,
		Status:                     option.OrderStatus(item.Status),
		CancelReason:               item.CancelReason,
		MatchTime:                  item.MatchTime,
		CancelTime:                 item.CancelTime,
		CreateTimes:                item.CreateTimes,
		UpdateTimes:                item.UpdateTimes,
	}
}

func ToTradeProto(item *models.TOptionTrade) *option.OptionTrade {
	if item == nil {
		return nil
	}
	return &option.OptionTrade{
		Id:               item.Id,
		TenantId:         item.TenantId,
		TradeNo:          item.TradeNo,
		ContractId:       item.ContractId,
		UnderlyingSymbol: item.UnderlyingSymbol,
		BuyOrderId:       item.BuyOrderId,
		BuyOrderNo:       item.BuyOrderNo,
		BuyUserId:        item.BuyUserId,
		BuyAccountId:     item.BuyAccountId,
		SellOrderId:      item.SellOrderId,
		SellOrderNo:      item.SellOrderNo,
		SellUserId:       item.SellUserId,
		SellAccountId:    item.SellAccountId,
		Price:            conv.FloatString(item.Price),
		Qty:              conv.FloatString(item.Qty),
		Turnover:         conv.FloatString(item.Turnover),
		BuyFee:           conv.FloatString(item.BuyFee),
		SellFee:          conv.FloatString(item.SellFee),
		FeeCoin:          item.FeeCoin,
		MakerSide:        common.Side(item.MakerSide),
		MatchSequence:    item.MatchSequence,
		ComboMatchNo:     item.ComboMatchNo,
		ComboLegNo:       item.ComboLegNo,
		TradeTime:        item.TradeTime,
		CreateTimes:      item.CreateTimes,
	}
}

func ToPositionProto(item *models.TOptionPosition) *option.OptionPosition {
	if item == nil {
		return nil
	}
	return &option.OptionPosition{
		Id:                item.Id,
		TenantId:          item.TenantId,
		UserId:            item.UserId,
		AccountId:         item.AccountId,
		ContractId:        item.ContractId,
		UnderlyingSymbol:  item.UnderlyingSymbol,
		Side:              common.PositionSide(item.Side),
		PositionQty:       conv.FloatString(item.PositionQty),
		AvailableQty:      conv.FloatString(item.AvailableQty),
		FrozenQty:         conv.FloatString(item.FrozenQty),
		OpenAvgPrice:      conv.FloatString(item.OpenAvgPrice),
		MarkPrice:         conv.FloatString(item.MarkPrice),
		PositionValue:     conv.FloatString(item.PositionValue),
		MarginAmount:      conv.FloatString(item.MarginAmount),
		MaintenanceMargin: conv.FloatString(item.MaintenanceMargin),
		UnrealizedPnl:     conv.FloatString(item.UnrealizedPnl),
		RealizedPnl:       conv.FloatString(item.RealizedPnl),
		TradeRealizedPnl:  conv.FloatString(item.TradeRealizedPnl),
		SettlementRealizedPnl: conv.FloatString(
			item.SettlementRealizedPnl,
		),
		FeePaid:         conv.FloatString(item.FeePaid),
		TotalReturn:     conv.FloatString(item.TotalReturn),
		ExerciseableQty: conv.FloatString(item.ExerciseableQty),
		Status:          option.PositionStatus(item.Status),
		LastCalcTime:    item.LastCalcTime,
		CreateTimes:     item.CreateTimes,
		UpdateTimes:     item.UpdateTimes,
	}
}

func ToExerciseProto(item *models.TOptionExercise) *option.OptionExercise {
	if item == nil {
		return nil
	}
	return &option.OptionExercise{
		Id:               item.Id,
		TenantId:         item.TenantId,
		ExerciseNo:       item.ExerciseNo,
		ClientExerciseId: item.ClientExerciseId,
		UserId:           item.UserId,
		AccountId:        item.AccountId,
		ContractId:       item.ContractId,
		PositionId:       item.PositionId,
		ExerciseType:     option.ExerciseType(item.ExerciseType),
		ExerciseQty:      conv.FloatString(item.ExerciseQty),
		StrikePrice:      conv.FloatString(item.StrikePrice),
		SettlementPrice:  conv.FloatString(item.SettlementPrice),
		ExerciseAmount:   conv.FloatString(item.ExerciseAmount),
		ProfitAmount:     conv.FloatString(item.ProfitAmount),
		Fee:              conv.FloatString(item.Fee),
		FeeCoin:          item.FeeCoin,
		Status:           option.ExerciseStatus(item.Status),
		Remark:           item.Remark,
		ExerciseTime:     item.ExerciseTime,
		FinishTime:       item.FinishTime,
		CreateTimes:      item.CreateTimes,
		UpdateTimes:      item.UpdateTimes,
	}
}

func ToExerciseInstructionProto(item *models.TOptionExerciseInstruction) *option.OptionExerciseInstruction {
	if item == nil {
		return nil
	}
	return &option.OptionExerciseInstruction{
		Id:                  item.Id,
		TenantId:            item.TenantId,
		UserId:              item.UserId,
		AccountId:           item.AccountId,
		ContractId:          item.ContractId,
		PositionId:          item.PositionId,
		ClientInstructionId: item.ClientInstructionId,
		InstructionType:     option.ExerciseInstructionType(item.InstructionType),
		Version:             item.Version,
		Status:              option.ExerciseInstructionStatus(item.Status),
		SupersedesId:        item.SupersedesId,
		CutoffTime:          item.CutoffTime,
		CreateTimes:         item.CreateTimes,
		UpdateTimes:         item.UpdateTimes,
	}
}

func ToExerciseAssignmentProto(item *models.TOptionExerciseAssignment) *option.OptionExerciseAssignment {
	if item == nil {
		return nil
	}
	return &option.OptionExerciseAssignment{
		Id: item.Id, TenantId: item.TenantId, ExerciseId: item.ExerciseId,
		ExerciseNo: item.ExerciseNo, LongPositionId: item.LongPositionId,
		ShortPositionId: item.ShortPositionId, ShortUserId: item.ShortUserId,
		ShortAccountId: item.ShortAccountId, Quantity: conv.FloatString(item.Quantity),
		Payoff: conv.FloatString(item.Payoff), Status: option.ExerciseAssignmentStatus(item.Status),
		InstructionNo: item.InstructionNo, CreateTimes: item.CreateTimes, UpdateTimes: item.UpdateTimes,
	}
}

func ToAssetInstructionProto(item *models.TOptionAssetInstruction) *option.OptionAssetInstruction {
	if item == nil {
		return nil
	}
	return &option.OptionAssetInstruction{
		Id: item.Id, TenantId: item.TenantId, InstructionNo: item.InstructionNo,
		BizNo: item.BizNo, OrderId: item.OrderId, TradeId: item.TradeId,
		PositionId: item.PositionId, UserId: item.UserId, AccountId: item.AccountId,
		Action: option.AssetInstructionAction(item.Action), TargetBizNo: item.TargetBizNo,
		Coin: item.Coin, Amount: conv.FloatString(item.Amount), StepNo: item.StepNo,
		Status: option.AssetInstructionStatus(item.Status), RetryCount: item.RetryCount,
		NextRetryAt: item.NextRetryAt, LastErrorMsg: item.LastErrorMsg,
		CreateTimes: item.CreateTimes, UpdateTimes: item.UpdateTimes,
		AssetFlowNo:          item.AssetFlowNo,
		ReconciliationStatus: option.AssetReconciliationStatus(item.ReconciliationStatus),
		ReconciledAt:         item.ReconciledAt, MarginLotId: item.MarginLotId,
		LiquidationId:  item.LiquidationId,
		DeliveryUnitId: item.DeliveryUnitId, ExecutionGroup: item.ExecutionGroup,
	}
}

func ToPhysicalDeliveryUnitProto(
	item *models.TOptionPhysicalDeliveryUnit,
	instructions []*models.TOptionAssetInstruction,
) *option.OptionPhysicalDeliveryUnit {
	if item == nil {
		return nil
	}
	result := &option.OptionPhysicalDeliveryUnit{
		Id: item.Id, TenantId: item.TenantId, DeliveryUnitNo: item.DeliveryUnitNo,
		BatchId: item.BatchId, BatchNo: item.BatchNo, ContractId: item.ContractId,
		LongPositionId: item.LongPositionId, LongUserId: item.LongUserId,
		LongAccountId: item.LongAccountId, ShortPositionId: item.ShortPositionId,
		ShortUserId: item.ShortUserId, ShortAccountId: item.ShortAccountId,
		Quantity: conv.FloatString(item.Quantity), DeliveryCoin: item.DeliveryCoin,
		DeliveryQuantity: conv.FloatString(item.DeliveryQuantity), PaymentCoin: item.PaymentCoin,
		PaymentAmount: conv.FloatString(item.PaymentAmount),
		Status:        option.PhysicalDeliveryUnitStatus(item.Status), CureDeadline: item.CureDeadline,
		FailedInstructionId: item.FailedInstructionId, LastErrorMsg: item.LastErrorMsg,
		CompletedAt: item.CompletedAt, CreateTimes: item.CreateTimes, UpdateTimes: item.UpdateTimes,
		ManualRetryCount: item.ManualRetryCount,
	}
	for _, instruction := range instructions {
		result.AssetInstructions = append(result.AssetInstructions, ToAssetInstructionProto(instruction))
	}
	return result
}

func ToSettlementProto(item *models.TOptionSettlement) *option.OptionSettlement {
	if item == nil {
		return nil
	}
	return &option.OptionSettlement{
		Id:               item.Id,
		TenantId:         item.TenantId,
		SettlementNo:     item.SettlementNo,
		ContractId:       item.ContractId,
		UnderlyingSymbol: item.UnderlyingSymbol,
		ExpireTime:       item.ExpireTime,
		SettlementTime:   item.SettlementTime,
		DeliveryPrice:    conv.FloatString(item.DeliveryPrice),
		TheoreticalPrice: conv.FloatString(item.TheoreticalPrice),
		Iv:               conv.FloatString(item.Iv),
		IsItm:            common.YesNo(item.IsItm),
		ExerciseResult:   option.ExerciseResult(item.ExerciseResult),
		Status:           option.SettlementStatus(item.Status),
		Remark:           item.Remark,
		CreateTimes:      item.CreateTimes,
		UpdateTimes:      item.UpdateTimes,
	}
}

func ToSettlementPriceProto(item *models.TOptionSettlementPrice) *option.OptionSettlementPrice {
	if item == nil {
		return nil
	}
	return &option.OptionSettlementPrice{
		Id: item.Id, TenantId: item.TenantId, ContractId: item.ContractId,
		PriceSource: item.PriceSource, WindowStart: item.WindowStart, WindowEnd: item.WindowEnd,
		SampleCount: item.SampleCount, CalculationMethod: item.CalculationMethod,
		DeliveryPrice: conv.FloatString(item.DeliveryPrice), SourceSnapshotIds: item.SourceSnapshotIds,
		Version: item.Version, Status: option.SettlementPriceStatus(item.Status),
		ConfirmedBy: item.ConfirmedBy, ConfirmedAt: item.ConfirmedAt,
		CreateTimes: item.CreateTimes, UpdateTimes: item.UpdateTimes,
		SupersedesId: item.SupersedesId, ChangeReason: item.ChangeReason, CreatedBy: item.CreatedBy,
	}
}

func ToAccountProto(item *models.TOptionAccount) *option.OptionAccount {
	if item == nil {
		return nil
	}
	return &option.OptionAccount{
		Id:               item.Id,
		TenantId:         item.TenantId,
		UserId:           item.UserId,
		AccountId:        item.AccountId,
		MarginCoin:       item.MarginCoin,
		Balance:          conv.FloatString(item.Balance),
		AvailableBalance: conv.FloatString(item.AvailableBalance),
		FrozenBalance:    conv.FloatString(item.FrozenBalance),
		PositionMargin:   conv.FloatString(item.PositionMargin),
		OrderMargin:      conv.FloatString(item.OrderMargin),
		UnrealizedPnl:    conv.FloatString(item.UnrealizedPnl),
		RealizedPnl:      conv.FloatString(item.RealizedPnl),
		RiskRate:         conv.FloatString(item.RiskRate),
		Status:           option.AccountStatus(item.Status),
		CreateTimes:      item.CreateTimes,
		UpdateTimes:      item.UpdateTimes,
	}
}

func ToBillProto(item *models.TOptionBill) *option.OptionBill {
	if item == nil {
		return nil
	}
	return &option.OptionBill{
		Id:            item.Id,
		TenantId:      item.TenantId,
		UserId:        item.UserId,
		AccountId:     item.AccountId,
		BizNo:         item.BizNo,
		RefType:       option.BillRefType(item.RefType),
		RefId:         item.RefId,
		Coin:          item.Coin,
		ChangeAmount:  conv.FloatString(item.ChangeAmount),
		BalanceBefore: conv.FloatString(item.BalanceBefore),
		BalanceAfter:  conv.FloatString(item.BalanceAfter),
		Remark:        item.Remark,
		CreateTimes:   item.CreateTimes,
	}
}

func ToRiskAccountProto(item *models.TOptionRiskAccount) *option.OptionRiskAccount {
	if item == nil {
		return nil
	}
	return &option.OptionRiskAccount{
		Id: item.Id, TenantId: item.TenantId, UserId: item.UserId, AccountId: item.AccountId,
		SettleCoin: item.SettleCoin, Equity: conv.FloatString(item.Equity),
		NetOptionValue:    conv.FloatString(item.NetOptionValue),
		PositionMargin:    conv.FloatString(item.PositionMargin),
		MaintenanceMargin: conv.FloatString(item.MaintenanceMargin),
		UnrealizedPnl:     conv.FloatString(item.UnrealizedPnl),
		RiskRate:          conv.FloatString(item.RiskRate), Status: option.RiskAccountStatus(item.Status),
		PortfolioRiskMethod:         option.PortfolioRiskMethod(item.PortfolioRiskMethod),
		PortfolioRiskConfigId:       item.PortfolioRiskConfigId,
		PortfolioRiskConfigVersion:  item.PortfolioRiskConfigVersion,
		PortfolioScenarioLoss:       conv.FloatString(item.PortfolioScenarioLoss),
		PortfolioShortFloor:         conv.FloatString(item.PortfolioShortFloor),
		PortfolioConcentrationAddon: conv.FloatString(item.PortfolioConcentrationAddon),
		PortfolioLiquidityAddon:     conv.FloatString(item.PortfolioLiquidityAddon),
		LastCalcTime:                item.LastCalcTime, CreateTimes: item.CreateTimes, UpdateTimes: item.UpdateTimes,
	}
}

func ToLiquidationProto(item *models.TOptionLiquidation) *option.OptionLiquidation {
	if item == nil {
		return nil
	}
	return &option.OptionLiquidation{
		Id: item.Id, TenantId: item.TenantId, LiquidationNo: item.LiquidationNo,
		UserId: item.UserId, AccountId: item.AccountId, ContractId: item.ContractId,
		PositionId: item.PositionId, Quantity: conv.FloatString(item.Quantity),
		MarkPrice:         conv.FloatString(item.MarkPrice),
		MaintenanceMargin: conv.FloatString(item.MaintenanceMargin),
		Equity:            conv.FloatString(item.Equity), DeficitAmount: conv.FloatString(item.DeficitAmount),
		LiquidationFee: conv.FloatString(item.LiquidationFee),
		Status:         option.LiquidationStatus(item.Status), RetryCount: item.RetryCount,
		LastErrorMsg: item.LastErrorMsg, CreateTimes: item.CreateTimes, UpdateTimes: item.UpdateTimes,
		CollateralAmount:    conv.FloatString(item.CollateralAmount),
		InsuranceFundAmount: conv.FloatString(item.InsuranceFundAmount),
		RemainingDeficit:    conv.FloatString(item.RemainingDeficit),
		TakeoverPositionId:  item.TakeoverPositionId, CompletedAt: item.CompletedAt,
		InsuranceAttempt: item.InsuranceAttempt, BackstopAmount: conv.FloatString(item.BackstopAmount),
		DeficitResolution: option.LiquidationDeficitResolution(item.DeficitResolution),
	}
}

func BuildContractDetail(ctx context.Context, svcCtx *svc.ServiceContext, contract *models.TOptionContract) (*option.OptionContractDetail, error) {
	market, err := FindMarketIgnoreNotFound(ctx, svcCtx, contract.TenantId, contract.Id)
	if err != nil {
		return nil, err
	}
	return &option.OptionContractDetail{
		Contract: ToContractProto(contract),
		Market:   ToMarketProto(market),
	}, nil
}

func BuildOrderDetail(ctx context.Context, svcCtx *svc.ServiceContext, item *models.TOptionOrder) (*option.OptionOrderDetail, error) {
	contract, err := FindContractIgnoreNotFound(ctx, svcCtx, item.TenantId, item.ContractId)
	if err != nil {
		return nil, err
	}
	return &option.OptionOrderDetail{Order: ToOrderProto(item), Contract: ToContractProto(contract)}, nil
}

func BuildTradeDetail(ctx context.Context, svcCtx *svc.ServiceContext, item *models.TOptionTrade) (*option.OptionTradeDetail, error) {
	contract, err := FindContractIgnoreNotFound(ctx, svcCtx, item.TenantId, item.ContractId)
	if err != nil {
		return nil, err
	}
	return &option.OptionTradeDetail{Trade: ToTradeProto(item), Contract: ToContractProto(contract)}, nil
}

func BuildPositionDetail(ctx context.Context, svcCtx *svc.ServiceContext, item *models.TOptionPosition) (*option.OptionPositionDetail, error) {
	contract, err := FindContractIgnoreNotFound(ctx, svcCtx, item.TenantId, item.ContractId)
	if err != nil {
		return nil, err
	}
	market, err := FindMarketIgnoreNotFound(ctx, svcCtx, item.TenantId, item.ContractId)
	if err != nil {
		return nil, err
	}
	return &option.OptionPositionDetail{
		Position: ToPositionProto(item),
		Contract: ToContractProto(contract),
		Market:   ToMarketProto(market),
	}, nil
}

func BuildExerciseDetail(ctx context.Context, svcCtx *svc.ServiceContext, item *models.TOptionExercise) (*option.OptionExerciseDetail, error) {
	contract, err := FindContractIgnoreNotFound(ctx, svcCtx, item.TenantId, item.ContractId)
	if err != nil {
		return nil, err
	}
	assignments, err := svcCtx.OptionExerciseAssignmentModel.FindByExercise(ctx, item.TenantId, item.Id)
	if err != nil {
		return nil, err
	}
	instructions, err := svcCtx.OptionAssetInstructionModel.FindByBizNo(ctx, item.TenantId, item.ExerciseNo)
	if err != nil {
		return nil, err
	}
	result := &option.OptionExerciseDetail{
		Exercise: ToExerciseProto(item),
		Contract: ToContractProto(contract),
	}
	for _, assignment := range assignments {
		result.Assignments = append(result.Assignments, ToExerciseAssignmentProto(assignment))
	}
	for _, instruction := range instructions {
		result.AssetInstructions = append(result.AssetInstructions, ToAssetInstructionProto(instruction))
	}
	return result, nil
}

func BuildSettlementDetail(ctx context.Context, svcCtx *svc.ServiceContext, item *models.TOptionSettlement) (*option.OptionSettlementDetail, error) {
	contract, err := FindContractIgnoreNotFound(ctx, svcCtx, item.TenantId, item.ContractId)
	if err != nil {
		return nil, err
	}
	batch, err := svcCtx.OptionSettlementBatchModel.FindOneByTenantIdBatchNo(ctx, item.TenantId, item.SettlementNo)
	if err != nil {
		return nil, err
	}
	details, err := svcCtx.OptionSettlementDetailModel.FindByBatch(ctx, item.TenantId, batch.Id)
	if err != nil {
		return nil, err
	}
	instructions, err := svcCtx.OptionAssetInstructionModel.FindByBizNo(ctx, item.TenantId, item.SettlementNo)
	if err != nil {
		return nil, err
	}
	result := &option.OptionSettlementDetail{
		Settlement: ToSettlementProto(item),
		Contract:   ToContractProto(contract),
		Batch: &option.OptionSettlementBatch{
			Id: batch.Id, TenantId: batch.TenantId, BatchNo: batch.BatchNo,
			ContractId: batch.ContractId, SettlementPriceId: batch.SettlementPriceId,
			TotalCredit: conv.FloatString(batch.TotalCredit), TotalDebit: conv.FloatString(batch.TotalDebit),
			InstructionCount: batch.InstructionCount, SuccessCount: batch.SuccessCount,
			Status: option.SettlementBatchStatus(batch.Status), LastErrorMsg: batch.LastErrorMsg,
			CreateTimes: batch.CreateTimes, UpdateTimes: batch.UpdateTimes,
		},
	}
	for _, detail := range details {
		result.PositionDetails = append(result.PositionDetails, &option.OptionSettlementPositionDetail{
			Id: detail.Id, TenantId: detail.TenantId, BatchId: detail.BatchId,
			BatchNo: detail.BatchNo, ContractId: detail.ContractId, PositionId: detail.PositionId,
			UserId: detail.UserId, AccountId: detail.AccountId,
			Side: common.PositionSide(detail.Side), Quantity: conv.FloatString(detail.Quantity),
			Payoff:        conv.FloatString(detail.Payoff),
			Direction:     option.SettlementDetailDirection(detail.Direction),
			InstructionNo: detail.InstructionNo, CreateTimes: detail.CreateTimes,
			DeliveryCoin: detail.DeliveryCoin, DeliveryQuantity: conv.FloatString(detail.DeliveryQuantity),
			PaymentCoin: detail.PaymentCoin, PaymentAmount: conv.FloatString(detail.PaymentAmount),
		})
	}
	for _, instruction := range instructions {
		result.AssetInstructions = append(result.AssetInstructions, ToAssetInstructionProto(instruction))
	}
	return result, nil
}
