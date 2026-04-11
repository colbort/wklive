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

type OrderDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewOrderDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrderDetailLogic {
	return &OrderDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取质押订单详情
func (l *OrderDetailLogic) OrderDetail(in *staking.AdminOrderDetailReq) (*staking.AdminOrderDetailResp, error) {
	item, err := l.svcCtx.StakeOrderModel.FindOne(l.ctx, in.Id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return &staking.AdminOrderDetailResp{Page: helper.GetErrResp(404, i18n.Translate(i18n.OrderNotFound, l.ctx))}, nil
		}
		return nil, err
	}
	if item.TenantId != in.TenantId {
		return &staking.AdminOrderDetailResp{Page: helper.GetErrResp(404, i18n.Translate(i18n.OrderNotFound, l.ctx))}, nil
	}

	return &staking.AdminOrderDetailResp{Page: helper.OkResp(), Data: orderToProto(item)}, nil
}
