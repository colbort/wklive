package adminlogic

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
		return &trade.GetOrderDetailAdminResp{Base: helper.ErrResp(i18n.OrderNotFound, i18n.Translate(i18n.OrderNotFound, l.ctx))}, nil
	}
	if err != nil {
		return nil, err
	}

	order := orderToProto(item)
	if item.ProductType == int64(trade.ProductType_PRODUCT_TYPE_SECONDS) {
		seconds, findErr := l.svcCtx.TradeOrderSecondsModel.FindOneByTenantIdOrderId(l.ctx, item.TenantId, item.Id)
		if findErr != nil && !errors.Is(findErr, models.ErrNotFound) {
			return nil, findErr
		}
		if findErr == nil {
			order.SecondsDirection = trade.SecondsDirection(seconds.Direction)
			order.DurationSeconds = seconds.DurationSeconds
		}
	}
	return &trade.GetOrderDetailAdminResp{Base: helper.OkResp(), Data: order}, nil
}
