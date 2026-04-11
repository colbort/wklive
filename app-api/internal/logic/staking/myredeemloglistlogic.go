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

type MyRedeemLogListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMyRedeemLogListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MyRedeemLogListLogic {
	return &MyRedeemLogListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MyRedeemLogListLogic) MyRedeemLogList(req *types.AppMyRedeemLogListReq) (resp *types.AppMyRedeemLogListResp, err error) {
	return logicutil.Proxy[types.AppMyRedeemLogListResp](l.ctx, req, l.svcCtx.StakingCli.MyRedeemLogList)
}
