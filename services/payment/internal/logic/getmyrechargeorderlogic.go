package logic

import (
	"context"
	"errors"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"
	"wklive/services/payment/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMyRechargeOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMyRechargeOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMyRechargeOrderLogic {
	return &GetMyRechargeOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询我的订单详情
func (l *GetMyRechargeOrderLogic) GetMyRechargeOrder(in *payment.GetMyRechargeOrderReq) (*payment.GetMyRechargeOrderResp, error) {
	order, err := l.svcCtx.RechargeOrderModel.FindOneByOrderNo(l.ctx, in.OrderNo)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if order == nil {
		return &payment.GetMyRechargeOrderResp{
			Base: helper.GetErrResp(404, i18n.Translate(i18n.OrderNotFound, l.ctx)),
		}, nil
	}

	// Check permission - user can only see their own orders
	if order.UserId != in.UserId || order.TenantId != in.TenantId {
		return &payment.GetMyRechargeOrderResp{
			Base: helper.GetErrResp(403, i18n.Translate(i18n.NoPermissionAccessOrder, l.ctx)),
		}, nil
	}

	return &payment.GetMyRechargeOrderResp{
		Base:  helper.OkResp(),
		Order: toRechargeOrderProto(order),
	}, nil
}
