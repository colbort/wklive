package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
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
	userId, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	tenantId, err := utils.GetTenantIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	if in.OrderId > 0 {
		item, err = l.svcCtx.TradeOrderModel.FindOne(l.ctx, in.OrderId)
	} else {
		item, err = l.svcCtx.TradeOrderModel.FindOneByTenantIdOrderNo(l.ctx, tenantId, in.OrderNo)
	}
	if errors.Is(err, models.ErrNotFound) || (err == nil && (item.TenantId != tenantId || item.UserId != userId)) {
		return &trade.GetOrderDetailResp{Base: helper.ErrResp(i18n.OrderNotFound, i18n.Translate(i18n.OrderNotFound, l.ctx))}, nil
	}
	if err != nil {
		return nil, err
	}

	resp := &trade.GetOrderDetailResp{
		Base: helper.OkResp(),
		Data: &trade.GetOrderDetailData{
			Order: orderToProto(item),
		},
	}
	spot, err := l.svcCtx.TradeOrderSpotModel.FindOneByTenantIdOrderId(l.ctx, tenantId, item.Id)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if spot != nil {
		resp.Data.Spot = orderSpotToProto(spot)
	}
	contractCfg, err := l.svcCtx.TradeOrderContractModel.FindOneByTenantIdOrderId(l.ctx, tenantId, item.Id)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if contractCfg != nil {
		resp.Data.Contract = orderContractToProto(contractCfg)
	}

	return resp, nil
}
