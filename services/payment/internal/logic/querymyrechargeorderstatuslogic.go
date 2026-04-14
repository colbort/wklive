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

type QueryMyRechargeOrderStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryMyRechargeOrderStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryMyRechargeOrderStatusLogic {
	return &QueryMyRechargeOrderStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 轮询订单状态
func (l *QueryMyRechargeOrderStatusLogic) QueryMyRechargeOrderStatus(in *payment.QueryMyRechargeOrderStatusReq) (*payment.QueryMyRechargeOrderStatusResp, error) {
	order, err := l.svcCtx.RechargeOrderModel.FindOneByOrderNo(l.ctx, in.OrderNo)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if order == nil {
		return &payment.QueryMyRechargeOrderStatusResp{
			Base: helper.GetErrResp(404, i18n.Translate(i18n.OrderNotFound, l.ctx)),
		}, nil
	}

	// Check permission
	if order.UserId != in.UserId || order.TenantId != in.TenantId {
		return &payment.QueryMyRechargeOrderStatusResp{
			Base: helper.GetErrResp(403, i18n.Translate(i18n.NoPermissionQueryOrder, l.ctx)),
		}, nil
	}

	return &payment.QueryMyRechargeOrderStatusResp{
		Base:  helper.OkResp(),
		Order: toRechargeOrderProto(order),
	}, nil
}
