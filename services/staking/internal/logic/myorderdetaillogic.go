package logic

import (
	"context"

	"wklive/proto/staking"
	"wklive/services/staking/internal/svc"

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
	// todo: add your logic here and delete this line

	return &staking.AppMyOrderDetailResp{}, nil
}
