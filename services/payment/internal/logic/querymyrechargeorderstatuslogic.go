package logic

import (
	"context"
	"errors"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
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
	userId, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	tenantId, err := utils.GetTenantIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	order, err := l.svcCtx.RechargeOrderModel.FindOneByOrderNo(l.ctx, in.OrderNo)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if order == nil {
		return &payment.QueryMyRechargeOrderStatusResp{
			Base: helper.ErrResp(i18n.OrderNotFound, i18n.Translate(i18n.OrderNotFound, l.ctx)),
		}, nil
	}

	// Check permission
	if order.UserId != userId || order.TenantId != tenantId {
		return &payment.QueryMyRechargeOrderStatusResp{
			Base: helper.ErrResp(i18n.NoPermissionQueryOrder, i18n.Translate(i18n.NoPermissionQueryOrder, l.ctx)),
		}, nil
	}

	return &payment.QueryMyRechargeOrderStatusResp{
		Base: helper.OkResp(),
		Data: toRechargeOrderProto(order),
	}, nil
}
