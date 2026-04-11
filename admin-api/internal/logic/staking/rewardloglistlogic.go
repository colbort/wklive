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

type RewardLogListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRewardLogListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RewardLogListLogic {
	return &RewardLogListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RewardLogListLogic) RewardLogList(req *types.AdminRewardLogListReq) (resp *types.AdminRewardLogListResp, err error) {
	return logicutil.Proxy[types.AdminRewardLogListResp](l.ctx, req, l.svcCtx.StakingCli.RewardLogList)
}
