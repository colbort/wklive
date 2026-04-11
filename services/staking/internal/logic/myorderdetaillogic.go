package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/staking"
	"wklive/services/staking/internal/svc"
	"wklive/services/staking/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type MyOrderDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMyOrderDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MyOrderDetailLogic {
	return &MyOrderDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取我的质押订单详情
func (l *MyOrderDetailLogic) MyOrderDetail(in *staking.AppMyOrderDetailReq) (*staking.AppMyOrderDetailResp, error) {
	item, err := l.svcCtx.StakeOrderModel.FindOne(l.ctx, in.Id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return &staking.AppMyOrderDetailResp{Base: helper.GetErrResp(404, i18n.Translate(i18n.OrderNotFound, l.ctx))}, nil
		}
		return nil, err
	}
	if item.TenantId != in.TenantId {
		return &staking.AppMyOrderDetailResp{Base: helper.GetErrResp(404, i18n.Translate(i18n.OrderNotFound, l.ctx))}, nil
	}
	if item.Uid != in.Uid {
		return &staking.AppMyOrderDetailResp{Base: helper.GetErrResp(403, i18n.Translate(i18n.NoPermissionAccessOrder, l.ctx))}, nil
	}

	return &staking.AppMyOrderDetailResp{Base: helper.OkResp(), Data: orderToProto(item)}, nil
}
