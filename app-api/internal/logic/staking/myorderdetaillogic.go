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

type MyOrderDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMyOrderDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MyOrderDetailLogic {
	return &MyOrderDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MyOrderDetailLogic) MyOrderDetail(req *types.AppMyOrderDetailReq) (resp *types.AppMyOrderDetailResp, err error) {
	return logicutil.Proxy[types.AppMyOrderDetailResp](l.ctx, req, l.svcCtx.StakingCli.MyOrderDetail)
}
