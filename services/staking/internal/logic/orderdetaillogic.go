package logic

import (
	"context"

	"wklive/proto/staking"
	"wklive/services/staking/internal/svc"

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
	// todo: add your logic here and delete this line

	return &staking.AdminOrderDetailResp{}, nil
}
