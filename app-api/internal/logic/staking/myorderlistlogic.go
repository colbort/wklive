// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package staking

import (
	"context"

	"wklive/app-api/internal/logicutil"
	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MyOrderListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMyOrderListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MyOrderListLogic {
	return &MyOrderListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MyOrderListLogic) MyOrderList(req *types.AppMyOrderListReq) (resp *types.AppMyOrderListResp, err error) {
	return logicutil.Proxy[types.AppMyOrderListResp](l.ctx, req, l.svcCtx.StakingCli.MyOrderList)
}
