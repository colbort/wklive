package logic

import (
	"context"
	"errors"
	"time"

	"wklive/common/helper"
	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
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
			return &option.AppCommonResp{Base: helper.GetErrResp(404, "订单不存在")}, nil
		}
		return nil, err
	}
	if item.Uid != in.Uid || item.AccountId != in.AccountId {
		return &option.AppCommonResp{Base: helper.GetErrResp(403, "无权操作该订单")}, nil
	}
	if item.Status != int64(option.OrderStatus_ORDER_STATUS_PENDING) && item.Status != int64(option.OrderStatus_ORDER_STATUS_PART_FILLED) {
		return &option.AppCommonResp{Base: helper.GetErrResp(400, "当前状态不可撤单")}, nil
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
