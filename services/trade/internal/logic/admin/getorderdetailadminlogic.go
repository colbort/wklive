package adminlogic

import (
	"context"
	"errors"
	"wklive/proto/common"
	helpers "wklive/services/trade/internal/logic/helpers"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOrderDetailAdminLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetOrderDetailAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrderDetailAdminLogic {
	return &GetOrderDetailAdminLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取订单详情
func (l *GetOrderDetailAdminLogic) GetOrderDetailAdmin(in *trade.GetOrderDetailAdminReq) (*trade.GetOrderDetailAdminResp, error) {
	item, err := l.svcCtx.TradeOrderModel.FindOne(l.ctx, in.Id)
	if errors.Is(err, models.ErrNotFound) || (err == nil && item.TenantId != in.TenantId) {
		return &trade.GetOrderDetailAdminResp{Base: helper.ErrResp(i18n.OrderNotFound, i18n.Translate(i18n.OrderNotFound, l.ctx))}, nil
	}
	if err != nil {
		return nil, err
	}

	order := helpers.OrderToProto(item)
	data := &trade.GetOrderDetailData{Order: order}
	switch common.ProductType(item.ProductType) {
	case common.ProductType_PRODUCT_TYPE_SPOT:
		spot, findErr := l.svcCtx.TradeOrderSpotModel.FindOneByTenantIdOrderId(l.ctx, item.TenantId, item.Id)
		if findErr != nil && !errors.Is(findErr, models.ErrNotFound) {
			return nil, findErr
		}
		if findErr == nil {
			data.Spot = helpers.OrderSpotToProto(spot)
		}
	case common.ProductType_PRODUCT_TYPE_DERIVATIVE:
		contract, findErr := l.svcCtx.TradeOrderContractModel.FindOneByTenantIdOrderId(l.ctx, item.TenantId, item.Id)
		if findErr != nil && !errors.Is(findErr, models.ErrNotFound) {
			return nil, findErr
		}
		if findErr == nil {
			data.Contract = helpers.OrderContractToProto(contract)
		}
	case common.ProductType_PRODUCT_TYPE_SECONDS:
		seconds, findErr := l.svcCtx.TradeOrderSecondsModel.FindOneByTenantIdOrderId(l.ctx, item.TenantId, item.Id)
		if findErr != nil && !errors.Is(findErr, models.ErrNotFound) {
			return nil, findErr
		}
		if findErr == nil {
			order.SecondsDirection = trade.SecondsDirection(seconds.Direction)
			order.DurationSeconds = seconds.DurationSeconds
			order.DisplayStatus = helpers.SecondsOrderDisplayStatus(seconds.SettlementStatus)
			data.Seconds = &trade.TradeOrderSeconds{
				Id: seconds.Id, TenantId: seconds.TenantId, OrderId: seconds.OrderId,
				Direction: trade.SecondsDirection(seconds.Direction), DurationSeconds: seconds.DurationSeconds,
				StakeAsset: seconds.StakeAsset, StakeAmount: seconds.StakeAmount.String(),
				PayoutRate: seconds.PayoutRate.String(), FeeRate: seconds.FeeRate.String(),
				FrozenAt: seconds.FrozenAt, ActivatedAt: seconds.ActivatedAt,
				StartPrice: seconds.StartPrice.String(), StartPriceTime: seconds.StartPriceTime,
				StartPriceSource: seconds.StartPriceSource, ExpireTime: seconds.ExpireTime,
				SettlementPrice: seconds.SettlementPrice.String(), SettlementPriceTime: seconds.SettlementPriceTime,
				SettlementPriceSource: seconds.SettlementPriceSource, PriceAlgorithm: seconds.PriceAlgorithm,
				Result: trade.SecondsResult(seconds.Result), ProfitAmount: seconds.ProfitAmount.String(),
				FeeAmount: seconds.FeeAmount.String(), ReturnAmount: seconds.ReturnAmount.String(),
				SettlementStatus: trade.SecondsSettlementStatus(seconds.SettlementStatus),
				ReservationNo:    seconds.ReservationNo, SettlementReason: seconds.SettlementReason,
				SettledAt: seconds.SettledAt, Version: seconds.Version,
				CreateTimes: seconds.CreateTimes, UpdateTimes: seconds.UpdateTimes,
			}
		}
	}
	return &trade.GetOrderDetailAdminResp{Base: helper.OkResp(), Data: data}, nil
}
