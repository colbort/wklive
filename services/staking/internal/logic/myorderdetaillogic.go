package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
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
	userId, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	tenantId, err := utils.GetTenantIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	item, err := l.svcCtx.StakeOrderModel.FindOne(l.ctx, in.Id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return &staking.AppMyOrderDetailResp{Base: helper.ErrResp(i18n.OrderNotFound, i18n.Translate(i18n.OrderNotFound, l.ctx))}, nil
		}
		return nil, err
	}
	if item.TenantId != tenantId {
		return &staking.AppMyOrderDetailResp{Base: helper.ErrResp(i18n.OrderNotFound, i18n.Translate(i18n.OrderNotFound, l.ctx))}, nil
	}
	if item.UserId != userId {
		return &staking.AppMyOrderDetailResp{Base: helper.ErrResp(i18n.NoPermissionAccessOrder, i18n.Translate(i18n.NoPermissionAccessOrder, l.ctx))}, nil
	}

	return &staking.AppMyOrderDetailResp{Base: helper.OkResp(), Data: orderToProto(item)}, nil
}
