package logic

import (
	"context"
	"errors"
	"github.com/zeromicro/go-zero/core/logx"
	"time"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"
)

type AppCancelOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAppCancelOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppCancelOrderLogic {
	return &AppCancelOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 撤销期权委托订单
func (l *AppCancelOrderLogic) AppCancelOrder(in *option.AppCancelOrderReq) (*option.AppCommonResp, error) {
	item, err := findOrderByNoOrID(l.ctx, l.svcCtx, in.TenantId, in.OrderId, in.OrderNo)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return &option.AppCommonResp{Base: helper.GetErrResp(404, i18n.Translate(i18n.OrderNotFound, l.ctx))}, nil
		}
		return nil, err
	}
	if item.Uid != in.Uid || item.AccountId != in.AccountId {
		return &option.AppCommonResp{Base: helper.GetErrResp(403, i18n.Translate(i18n.NoPermissionOperateOrder, l.ctx))}, nil
	}
	if item.Status != int64(option.OrderStatus_ORDER_STATUS_PENDING) && item.Status != int64(option.OrderStatus_ORDER_STATUS_PART_FILLED) {
		return &option.AppCommonResp{Base: helper.GetErrResp(400, i18n.Translate(i18n.CurrentStatusCannotCancel, l.ctx))}, nil
	}

	now := time.Now().Unix()
	item.Status = int64(option.OrderStatus_ORDER_STATUS_CANCELED)
	item.CancelReason = "USER_CANCEL"
	item.CancelTime = now
	item.UpdateTimes = now
	if err := l.svcCtx.OptionOrderModel.Update(l.ctx, item); err != nil {
		return nil, err
	}

	return &option.AppCommonResp{Base: helper.OkResp()}, nil
}
