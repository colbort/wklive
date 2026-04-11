// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package staking

import (
	"context"

	"wklive/admin-api/internal/logicutil"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ManualRewardLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewManualRewardLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ManualRewardLogic {
	return &ManualRewardLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ManualRewardLogic) ManualReward(req *types.AdminManualRewardReq) (resp *types.AdminManualRewardResp, err error) {
	return logicutil.Proxy[types.AdminManualRewardResp](l.ctx, req, l.svcCtx.StakingCli.ManualReward)
}
