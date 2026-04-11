package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOrderDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetOrderDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrderDetailLogic {
	return &GetOrderDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取订单详情
func (l *GetOrderDetailLogic) GetOrderDetail(in *trade.GetOrderDetailReq) (*trade.GetOrderDetailResp, error) {
	var (
		item *models.TTradeOrder
		err  error
	)
	if in.OrderId > 0 {
		item, err = l.svcCtx.TradeOrderModel.FindOne(l.ctx, in.OrderId)
	} else {
		item, err = l.svcCtx.TradeOrderModel.FindOneByTenantIdOrderNo(l.ctx, in.TenantId, in.OrderNo)
	}
	if errors.Is(err, models.ErrNotFound) || (err == nil && (item.TenantId != in.TenantId || item.UserId != in.UserId)) {
		return &trade.GetOrderDetailResp{Base: helper.GetErrResp(404, i18n.Translate(i18n.OrderNotFound, l.ctx))}, nil
	}
	if err != nil {
		return nil, err
	}

	resp := &trade.GetOrderDetailResp{Base: helper.OkResp(), Order: orderToProto(item)}
	spot, err := l.svcCtx.TradeOrderSpotModel.FindOneByTenantIdOrderId(l.ctx, in.TenantId, item.Id)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if spot != nil {
		resp.Spot = orderSpotToProto(spot)
	}
	contractCfg, err := l.svcCtx.TradeOrderContractModel.FindOneByTenantIdOrderId(l.ctx, in.TenantId, item.Id)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if contractCfg != nil {
		resp.Contract = orderContractToProto(contractCfg)
	}

	return resp, nil
}
