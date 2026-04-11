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

type MyRewardLogListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMyRewardLogListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MyRewardLogListLogic {
	return &MyRewardLogListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MyRewardLogListLogic) MyRewardLogList(req *types.AppMyRewardLogListReq) (resp *types.AppMyRewardLogListResp, err error) {
	return logicutil.Proxy[types.AppMyRewardLogListResp](l.ctx, req, l.svcCtx.StakingCli.MyRewardLogList)
}
