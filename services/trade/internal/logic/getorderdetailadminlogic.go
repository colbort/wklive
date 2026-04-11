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
		return &trade.GetOrderDetailAdminResp{Base: helper.GetErrResp(404, i18n.Translate(i18n.OrderNotFound, l.ctx))}, nil
	}
	if err != nil {
		return nil, err
	}

	return &trade.GetOrderDetailAdminResp{Base: helper.OkResp(), Data: orderToProto(item)}, nil
}
