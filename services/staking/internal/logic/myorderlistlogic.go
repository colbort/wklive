package logic

import (
	"context"

	"wklive/proto/staking"
	"wklive/services/staking/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type MyOrderListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMyOrderListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MyOrderListLogic {
	return &MyOrderListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取我的质押订单列表
func (l *MyOrderListLogic) MyOrderList(in *staking.AppMyOrderListReq) (*staking.AppMyOrderListResp, error) {
	// todo: add your logic here and delete this line

	return &staking.AppMyOrderListResp{}, nil
}
