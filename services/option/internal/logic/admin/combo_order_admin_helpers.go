package adminlogic

import (
	"context"

	"wklive/common/conv"
	"wklive/proto/common"
	"wklive/proto/option"
	logichelpers "wklive/services/option/internal/logic/helpers"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"
)

const adminComboRelatedRowLimit int64 = 100

func findAdminComboOrder(
	ctx context.Context,
	svcCtx *svc.ServiceContext,
	tenantID, id int64,
	comboNo string,
) (*models.TOptionComboOrder, error) {
	if id > 0 {
		item, err := svcCtx.OptionComboOrderModel.FindOne(ctx, id)
		if err != nil {
			return nil, err
		}
		if item.TenantId != tenantID || (comboNo != "" && item.ComboNo != comboNo) {
			return nil, models.ErrNotFound
		}
		return item, nil
	}
	if comboNo == "" {
		return nil, models.ErrNotFound
	}
	return svcCtx.OptionComboOrderModel.FindOneByTenantIdComboNo(ctx, tenantID, comboNo)
}

func toAdminComboOrderProto(item *models.TOptionComboOrder) *option.OptionComboOrder {
	if item == nil {
		return nil
	}
	return &option.OptionComboOrder{
		Id: item.Id, TenantId: item.TenantId, ComboNo: item.ComboNo,
		UserId: item.UserId, AccountId: item.AccountId,
		ClientComboId: item.ClientComboId, StrategyKey: item.StrategyKey,
		InverseStrategyKey: item.InverseStrategyKey,
		UnderlyingSymbol:   item.UnderlyingSymbol, ExpireTime: item.ExpireTime,
		SettleCoin: item.SettleCoin, QuoteCoin: item.QuoteCoin,
		OrderType: option.ComboOrderType(item.OrderType),
		NetPrice:  conv.FloatString(item.NetPrice), Qty: conv.FloatString(item.Qty),
		FilledQty:   conv.FloatString(item.FilledQty),
		UnfilledQty: conv.FloatString(item.UnfilledQty),
		Status:      option.ComboOrderStatus(item.Status), PayloadHash: item.PayloadHash,
		CancelReason: item.CancelReason, CancelTime: item.CancelTime,
		CreateTimes: item.CreateTimes, UpdateTimes: item.UpdateTimes,
	}
}

func toAdminComboLegProto(item *models.TOptionComboOrderLeg) *option.OptionComboOrderLeg {
	if item == nil {
		return nil
	}
	return &option.OptionComboOrderLeg{
		Id: item.Id, TenantId: item.TenantId, ComboOrderId: item.ComboOrderId,
		LegNo: item.LegNo, ContractId: item.ContractId, Side: common.Side(item.Side),
		PositionEffect: option.PositionEffect(item.PositionEffect), Ratio: item.Ratio,
		Price: conv.FloatString(item.Price), Qty: conv.FloatString(item.Qty),
		FilledQty:    conv.FloatString(item.FilledQty),
		UnfilledQty:  conv.FloatString(item.UnfilledQty),
		ChildOrderId: item.ChildOrderId, CreateTimes: item.CreateTimes,
		UpdateTimes: item.UpdateTimes,
	}
}

func buildAdminComboSummary(
	ctx context.Context,
	svcCtx *svc.ServiceContext,
	item *models.TOptionComboOrder,
) (*option.OptionComboOrderDetail, error) {
	legs, err := svcCtx.OptionComboOrderLegModel.FindByComboOrderID(ctx, item.TenantId, item.Id)
	if err != nil {
		return nil, err
	}
	result := &option.OptionComboOrderDetail{
		ComboOrder: toAdminComboOrderProto(item),
		Legs:       make([]*option.OptionComboOrderLeg, 0, len(legs)),
	}
	for _, leg := range legs {
		result.Legs = append(result.Legs, toAdminComboLegProto(leg))
	}
	return result, nil
}

func buildAdminComboDetail(
	ctx context.Context,
	svcCtx *svc.ServiceContext,
	item *models.TOptionComboOrder,
) (*option.OptionAdminComboOrderDetail, error) {
	summary, err := buildAdminComboSummary(ctx, svcCtx, item)
	if err != nil {
		return nil, err
	}
	result := &option.OptionAdminComboOrderDetail{
		ComboOrder: summary.ComboOrder,
		Legs:       summary.Legs,
	}
	children, err := svcCtx.OptionOrderModel.FindComboChildren(ctx, item.TenantId, item.Id)
	if err != nil {
		return nil, err
	}
	for _, child := range children {
		detail, buildErr := logichelpers.BuildOrderDetail(ctx, svcCtx, child)
		if buildErr != nil {
			return nil, buildErr
		}
		result.ChildOrders = append(result.ChildOrders, detail)
	}
	trades, tradeTotal, err := svcCtx.OptionTradeModel.FindByComboOrderID(
		ctx, item.TenantId, item.Id, adminComboRelatedRowLimit,
	)
	if err != nil {
		return nil, err
	}
	result.TradeTotal = tradeTotal
	for _, trade := range trades {
		detail, buildErr := logichelpers.BuildTradeDetail(ctx, svcCtx, trade)
		if buildErr != nil {
			return nil, buildErr
		}
		result.Trades = append(result.Trades, detail)
	}
	instructions, instructionTotal, err := svcCtx.OptionAssetInstructionModel.FindByComboOrderID(
		ctx, item.TenantId, item.Id, adminComboRelatedRowLimit,
	)
	if err != nil {
		return nil, err
	}
	result.AssetInstructionTotal = instructionTotal
	for _, instruction := range instructions {
		result.AssetInstructions = append(
			result.AssetInstructions, logichelpers.ToAssetInstructionProto(instruction),
		)
	}
	result.DataTruncated = tradeTotal > int64(len(trades)) ||
		instructionTotal > int64(len(instructions))
	return result, nil
}
