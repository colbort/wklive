package logic

import (
	"context"
	"errors"

	commonconv "wklive/common/conv"
	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"
)

func findContractByCodeOrID(ctx context.Context, svcCtx *svc.ServiceContext, tenantId, id int64, contractCode string) (*models.TOptionContract, error) {
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

func findOrderByNoOrID(ctx context.Context, svcCtx *svc.ServiceContext, tenantId, id int64, orderNo string) (*models.TOptionOrder, error) {
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

func findTradeByNoOrID(ctx context.Context, svcCtx *svc.ServiceContext, tenantId, id int64, tradeNo string) (*models.TOptionTrade, error) {
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

func findExerciseByNoOrID(ctx context.Context, svcCtx *svc.ServiceContext, tenantId, id int64, exerciseNo string) (*models.TOptionExercise, error) {
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

func findSettlementByNoOrID(ctx context.Context, svcCtx *svc.ServiceContext, tenantId, id int64, settlementNo string) (*models.TOptionSettlement, error) {
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

func findContractIgnoreNotFound(ctx context.Context, svcCtx *svc.ServiceContext, tenantId, contractId int64) (*models.TOptionContract, error) {
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

func findMarketIgnoreNotFound(ctx context.Context, svcCtx *svc.ServiceContext, tenantId, contractId int64) (*models.TOptionMarket, error) {
	item, err := svcCtx.OptionMarketModel.FindOneByTenantIdContractId(ctx, tenantId, contractId)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return item, nil
}

func toContractProto(item *models.TOptionContract) *option.OptionContract {
	if item == nil {
		return nil
	}
	return &option.OptionContract{
		Id:               item.Id,
		TenantId:         item.TenantId,
		ContractCode:     item.ContractCode,
		UnderlyingSymbol: item.UnderlyingSymbol,
		SettleCoin:       item.SettleCoin,
		QuoteCoin:        item.QuoteCoin,
		OptionType:       option.OptionType(item.OptionType),
		ExerciseStyle:    option.ExerciseStyle(item.ExerciseStyle),
		SettlementType:   option.SettlementType(item.SettlementType),
		StrikePrice:      commonconv.FloatString(item.StrikePrice),
		ContractUnit:     commonconv.FloatString(item.ContractUnit),
		MinOrderQty:      commonconv.FloatString(item.MinOrderQty),
		MaxOrderQty:      commonconv.FloatString(item.MaxOrderQty),
		PriceTick:        commonconv.FloatString(item.PriceTick),
		QtyStep:          commonconv.FloatString(item.QtyStep),
		Multiplier:       commonconv.FloatString(item.Multiplier),
		ListTime:         item.ListTime,
		ExpireTime:       item.ExpireTime,
		DeliverTime:      item.DeliverTime,
		IsAutoExercise:   option.YesNo(item.IsAutoExercise),
		Status:           option.ContractStatus(item.Status),
		Sort:             int32(item.Sort),
		Remark:           item.Remark,
		IsDeleted:        option.YesNo(item.IsDeleted),
		CreateTimes:      item.CreateTimes,
		UpdateTimes:      item.UpdateTimes,
	}
}

func toMarketProto(item *models.TOptionMarket) *option.OptionMarket {
	if item == nil {
		return nil
	}
	return &option.OptionMarket{
		Id:               item.Id,
		TenantId:         item.TenantId,
		ContractId:       item.ContractId,
		UnderlyingPrice:  commonconv.FloatString(item.UnderlyingPrice),
		MarkPrice:        commonconv.FloatString(item.MarkPrice),
		LastPrice:        commonconv.FloatString(item.LastPrice),
		BidPrice:         commonconv.FloatString(item.BidPrice),
		AskPrice:         commonconv.FloatString(item.AskPrice),
		TheoreticalPrice: commonconv.FloatString(item.TheoreticalPrice),
		IntrinsicValue:   commonconv.FloatString(item.IntrinsicValue),
		TimeValue:        commonconv.FloatString(item.TimeValue),
		Iv:               commonconv.FloatString(item.Iv),
		Delta:            commonconv.FloatString(item.Delta),
		Gamma:            commonconv.FloatString(item.Gamma),
		Theta:            commonconv.FloatString(item.Theta),
		Vega:             commonconv.FloatString(item.Vega),
		Rho:              commonconv.FloatString(item.Rho),
		RiskFreeRate:     commonconv.FloatString(item.RiskFreeRate),
		PricingModel:     item.PricingModel,
		SnapshotTime:     item.SnapshotTime,
		CreateTimes:      item.CreateTimes,
		UpdateTimes:      item.UpdateTimes,
	}
}

func toMarketSnapshotProto(item *models.TOptionMarketSnapshot) *option.OptionMarketSnapshot {
	if item == nil {
		return nil
	}
	return &option.OptionMarketSnapshot{
		Id:               item.Id,
		TenantId:         item.TenantId,
		ContractId:       item.ContractId,
		UnderlyingPrice:  commonconv.FloatString(item.UnderlyingPrice),
		MarkPrice:        commonconv.FloatString(item.MarkPrice),
		LastPrice:        commonconv.FloatString(item.LastPrice),
		BidPrice:         commonconv.FloatString(item.BidPrice),
		AskPrice:         commonconv.FloatString(item.AskPrice),
		TheoreticalPrice: commonconv.FloatString(item.TheoreticalPrice),
		Iv:               commonconv.FloatString(item.Iv),
		Delta:            commonconv.FloatString(item.Delta),
		Gamma:            commonconv.FloatString(item.Gamma),
		Theta:            commonconv.FloatString(item.Theta),
		Vega:             commonconv.FloatString(item.Vega),
		Rho:              commonconv.FloatString(item.Rho),
		SnapshotTime:     item.SnapshotTime,
		CreateTimes:      item.CreateTimes,
	}
}

func toOrderProto(item *models.TOptionOrder) *option.OptionOrder {
	if item == nil {
		return nil
	}
	return &option.OptionOrder{
		Id:               item.Id,
		TenantId:         item.TenantId,
		OrderNo:          item.OrderNo,
		Uid:              item.Uid,
		AccountId:        item.AccountId,
		ContractId:       item.ContractId,
		UnderlyingSymbol: item.UnderlyingSymbol,
		Side:             option.Side(item.Side),
		PositionEffect:   option.PositionEffect(item.PositionEffect),
		OrderType:        option.OrderType(item.OrderType),
		Price:            commonconv.FloatString(item.Price),
		Qty:              commonconv.FloatString(item.Qty),
		FilledQty:        commonconv.FloatString(item.FilledQty),
		UnfilledQty:      commonconv.FloatString(item.UnfilledQty),
		AvgPrice:         commonconv.FloatString(item.AvgPrice),
		Turnover:         commonconv.FloatString(item.Turnover),
		Fee:              commonconv.FloatString(item.Fee),
		FeeCoin:          item.FeeCoin,
		MarginAmount:     commonconv.FloatString(item.MarginAmount),
		Source:           option.OrderSource(item.Source),
		ClientOrderId:    item.ClientOrderId,
		ReduceOnly:       option.YesNo(item.ReduceOnly),
		Mmp:              option.YesNo(item.Mmp),
		Status:           option.OrderStatus(item.Status),
		CancelReason:     item.CancelReason,
		MatchTime:        item.MatchTime,
		CancelTime:       item.CancelTime,
		CreateTimes:      item.CreateTimes,
		UpdateTimes:      item.UpdateTimes,
	}
}

func toTradeProto(item *models.TOptionTrade) *option.OptionTrade {
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
		BuyUid:           item.BuyUid,
		BuyAccountId:     item.BuyAccountId,
		SellOrderId:      item.SellOrderId,
		SellOrderNo:      item.SellOrderNo,
		SellUid:          item.SellUid,
		SellAccountId:    item.SellAccountId,
		Price:            commonconv.FloatString(item.Price),
		Qty:              commonconv.FloatString(item.Qty),
		Turnover:         commonconv.FloatString(item.Turnover),
		BuyFee:           commonconv.FloatString(item.BuyFee),
		SellFee:          commonconv.FloatString(item.SellFee),
		FeeCoin:          item.FeeCoin,
		MakerSide:        option.MakerSide(item.MakerSide),
		TradeTime:        item.TradeTime,
		CreateTimes:      item.CreateTimes,
	}
}

func toPositionProto(item *models.TOptionPosition) *option.OptionPosition {
	if item == nil {
		return nil
	}
	return &option.OptionPosition{
		Id:                item.Id,
		TenantId:          item.TenantId,
		Uid:               item.Uid,
		AccountId:         item.AccountId,
		ContractId:        item.ContractId,
		UnderlyingSymbol:  item.UnderlyingSymbol,
		Side:              option.PositionSide(item.Side),
		PositionQty:       commonconv.FloatString(item.PositionQty),
		AvailableQty:      commonconv.FloatString(item.AvailableQty),
		FrozenQty:         commonconv.FloatString(item.FrozenQty),
		OpenAvgPrice:      commonconv.FloatString(item.OpenAvgPrice),
		MarkPrice:         commonconv.FloatString(item.MarkPrice),
		PositionValue:     commonconv.FloatString(item.PositionValue),
		MarginAmount:      commonconv.FloatString(item.MarginAmount),
		MaintenanceMargin: commonconv.FloatString(item.MaintenanceMargin),
		UnrealizedPnl:     commonconv.FloatString(item.UnrealizedPnl),
		RealizedPnl:       commonconv.FloatString(item.RealizedPnl),
		ExerciseableQty:   commonconv.FloatString(item.ExerciseableQty),
		Status:            option.PositionStatus(item.Status),
		LastCalcTime:      item.LastCalcTime,
		CreateTimes:       item.CreateTimes,
		UpdateTimes:       item.UpdateTimes,
	}
}

func toExerciseProto(item *models.TOptionExercise) *option.OptionExercise {
	if item == nil {
		return nil
	}
	return &option.OptionExercise{
		Id:              item.Id,
		TenantId:        item.TenantId,
		ExerciseNo:      item.ExerciseNo,
		Uid:             item.Uid,
		AccountId:       item.AccountId,
		ContractId:      item.ContractId,
		PositionId:      item.PositionId,
		ExerciseType:    option.ExerciseType(item.ExerciseType),
		ExerciseQty:     commonconv.FloatString(item.ExerciseQty),
		StrikePrice:     commonconv.FloatString(item.StrikePrice),
		SettlementPrice: commonconv.FloatString(item.SettlementPrice),
		ExerciseAmount:  commonconv.FloatString(item.ExerciseAmount),
		ProfitAmount:    commonconv.FloatString(item.ProfitAmount),
		Fee:             commonconv.FloatString(item.Fee),
		FeeCoin:         item.FeeCoin,
		Status:          option.ExerciseStatus(item.Status),
		Remark:          item.Remark,
		ExerciseTime:    item.ExerciseTime,
		FinishTime:      item.FinishTime,
		CreateTimes:     item.CreateTimes,
		UpdateTimes:     item.UpdateTimes,
	}
}

func toSettlementProto(item *models.TOptionSettlement) *option.OptionSettlement {
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
		DeliveryPrice:    commonconv.FloatString(item.DeliveryPrice),
		TheoreticalPrice: commonconv.FloatString(item.TheoreticalPrice),
		Iv:               commonconv.FloatString(item.Iv),
		IsItm:            option.YesNo(item.IsItm),
		ExerciseResult:   option.ExerciseResult(item.ExerciseResult),
		Status:           option.SettlementStatus(item.Status),
		Remark:           item.Remark,
		CreateTimes:      item.CreateTimes,
		UpdateTimes:      item.UpdateTimes,
	}
}

func toAccountProto(item *models.TOptionAccount) *option.OptionAccount {
	if item == nil {
		return nil
	}
	return &option.OptionAccount{
		Id:               item.Id,
		TenantId:         item.TenantId,
		Uid:              item.Uid,
		AccountId:        item.AccountId,
		MarginCoin:       item.MarginCoin,
		Balance:          commonconv.FloatString(item.Balance),
		AvailableBalance: commonconv.FloatString(item.AvailableBalance),
		FrozenBalance:    commonconv.FloatString(item.FrozenBalance),
		PositionMargin:   commonconv.FloatString(item.PositionMargin),
		OrderMargin:      commonconv.FloatString(item.OrderMargin),
		UnrealizedPnl:    commonconv.FloatString(item.UnrealizedPnl),
		RealizedPnl:      commonconv.FloatString(item.RealizedPnl),
		RiskRate:         commonconv.FloatString(item.RiskRate),
		Status:           option.AccountStatus(item.Status),
		CreateTimes:      item.CreateTimes,
		UpdateTimes:      item.UpdateTimes,
	}
}

func toBillProto(item *models.TOptionBill) *option.OptionBill {
	if item == nil {
		return nil
	}
	return &option.OptionBill{
		Id:            item.Id,
		TenantId:      item.TenantId,
		Uid:           item.Uid,
		AccountId:     item.AccountId,
		BizNo:         item.BizNo,
		RefType:       option.BillRefType(item.RefType),
		RefId:         item.RefId,
		Coin:          item.Coin,
		ChangeAmount:  commonconv.FloatString(item.ChangeAmount),
		BalanceBefore: commonconv.FloatString(item.BalanceBefore),
		BalanceAfter:  commonconv.FloatString(item.BalanceAfter),
		Remark:        item.Remark,
		CreateTimes:   item.CreateTimes,
	}
}

func buildContractDetail(ctx context.Context, svcCtx *svc.ServiceContext, contract *models.TOptionContract) (*option.OptionContractDetail, error) {
	market, err := findMarketIgnoreNotFound(ctx, svcCtx, contract.TenantId, contract.Id)
	if err != nil {
		return nil, err
	}
	return &option.OptionContractDetail{
		Contract: toContractProto(contract),
		Market:   toMarketProto(market),
	}, nil
}

func buildOrderDetail(ctx context.Context, svcCtx *svc.ServiceContext, item *models.TOptionOrder) (*option.OptionOrderDetail, error) {
	contract, err := findContractIgnoreNotFound(ctx, svcCtx, item.TenantId, item.ContractId)
	if err != nil {
		return nil, err
	}
	return &option.OptionOrderDetail{Order: toOrderProto(item), Contract: toContractProto(contract)}, nil
}

func buildTradeDetail(ctx context.Context, svcCtx *svc.ServiceContext, item *models.TOptionTrade) (*option.OptionTradeDetail, error) {
	contract, err := findContractIgnoreNotFound(ctx, svcCtx, item.TenantId, item.ContractId)
	if err != nil {
		return nil, err
	}
	return &option.OptionTradeDetail{Trade: toTradeProto(item), Contract: toContractProto(contract)}, nil
}

func buildPositionDetail(ctx context.Context, svcCtx *svc.ServiceContext, item *models.TOptionPosition) (*option.OptionPositionDetail, error) {
	contract, err := findContractIgnoreNotFound(ctx, svcCtx, item.TenantId, item.ContractId)
	if err != nil {
		return nil, err
	}
	market, err := findMarketIgnoreNotFound(ctx, svcCtx, item.TenantId, item.ContractId)
	if err != nil {
		return nil, err
	}
	return &option.OptionPositionDetail{
		Position: toPositionProto(item),
		Contract: toContractProto(contract),
		Market:   toMarketProto(market),
	}, nil
}

func buildExerciseDetail(ctx context.Context, svcCtx *svc.ServiceContext, item *models.TOptionExercise) (*option.OptionExerciseDetail, error) {
	contract, err := findContractIgnoreNotFound(ctx, svcCtx, item.TenantId, item.ContractId)
	if err != nil {
		return nil, err
	}
	return &option.OptionExerciseDetail{Exercise: toExerciseProto(item), Contract: toContractProto(contract)}, nil
}

func buildSettlementDetail(ctx context.Context, svcCtx *svc.ServiceContext, item *models.TOptionSettlement) (*option.OptionSettlementDetail, error) {
	contract, err := findContractIgnoreNotFound(ctx, svcCtx, item.TenantId, item.ContractId)
	if err != nil {
		return nil, err
	}
	return &option.OptionSettlementDetail{Settlement: toSettlementProto(item), Contract: toContractProto(contract)}, nil
}
