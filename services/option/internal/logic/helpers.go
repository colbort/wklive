package logic

import (
	"context"
	"errors"

	"wklive/common/conv"
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
		StrikePrice:      conv.FloatString(item.StrikePrice),
		ContractUnit:     conv.FloatString(item.ContractUnit),
		MinOrderQty:      conv.FloatString(item.MinOrderQty),
		MaxOrderQty:      conv.FloatString(item.MaxOrderQty),
		PriceTick:        conv.FloatString(item.PriceTick),
		QtyStep:          conv.FloatString(item.QtyStep),
		Multiplier:       conv.FloatString(item.Multiplier),
		ListTime:         item.ListTime,
		ExpireTime:       item.ExpireTime,
		DeliverTime:      item.DeliverTime,
		IsAutoExercise:   option.YesNo(item.IsAutoExercise),
		Status:           option.ContractStatus(item.Status),
		Sort:             item.Sort,
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
		UnderlyingPrice:  conv.FloatString(item.UnderlyingPrice),
		MarkPrice:        conv.FloatString(item.MarkPrice),
		LastPrice:        conv.FloatString(item.LastPrice),
		BidPrice:         conv.FloatString(item.BidPrice),
		AskPrice:         conv.FloatString(item.AskPrice),
		TheoreticalPrice: conv.FloatString(item.TheoreticalPrice),
		IntrinsicValue:   conv.FloatString(item.IntrinsicValue),
		TimeValue:        conv.FloatString(item.TimeValue),
		Iv:               conv.FloatString(item.Iv),
		Delta:            conv.FloatString(item.Delta),
		Gamma:            conv.FloatString(item.Gamma),
		Theta:            conv.FloatString(item.Theta),
		Vega:             conv.FloatString(item.Vega),
		Rho:              conv.FloatString(item.Rho),
		RiskFreeRate:     conv.FloatString(item.RiskFreeRate),
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
		Price:            conv.FloatString(item.Price),
		Qty:              conv.FloatString(item.Qty),
		FilledQty:        conv.FloatString(item.FilledQty),
		UnfilledQty:      conv.FloatString(item.UnfilledQty),
		AvgPrice:         conv.FloatString(item.AvgPrice),
		Turnover:         conv.FloatString(item.Turnover),
		Fee:              conv.FloatString(item.Fee),
		FeeCoin:          item.FeeCoin,
		MarginAmount:     conv.FloatString(item.MarginAmount),
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
		Price:            conv.FloatString(item.Price),
		Qty:              conv.FloatString(item.Qty),
		Turnover:         conv.FloatString(item.Turnover),
		BuyFee:           conv.FloatString(item.BuyFee),
		SellFee:          conv.FloatString(item.SellFee),
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
		ExerciseableQty:   conv.FloatString(item.ExerciseableQty),
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
		ExerciseQty:     conv.FloatString(item.ExerciseQty),
		StrikePrice:     conv.FloatString(item.StrikePrice),
		SettlementPrice: conv.FloatString(item.SettlementPrice),
		ExerciseAmount:  conv.FloatString(item.ExerciseAmount),
		ProfitAmount:    conv.FloatString(item.ProfitAmount),
		Fee:             conv.FloatString(item.Fee),
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
		DeliveryPrice:    conv.FloatString(item.DeliveryPrice),
		TheoreticalPrice: conv.FloatString(item.TheoreticalPrice),
		Iv:               conv.FloatString(item.Iv),
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
		ChangeAmount:  conv.FloatString(item.ChangeAmount),
		BalanceBefore: conv.FloatString(item.BalanceBefore),
		BalanceAfter:  conv.FloatString(item.BalanceAfter),
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
