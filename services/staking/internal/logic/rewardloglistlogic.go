package logic

import (
	"context"

	"wklive/proto/staking"
	"wklive/services/staking/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type RewardLogListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRewardLogListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RewardLogListLogic {
	return &RewardLogListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取收益记录列表
func (l *RewardLogListLogic) RewardLogList(in *staking.AdminRewardLogListReq) (*staking.AdminRewardLogListResp, error) {
	// todo: add your logic here and delete this line

	return &staking.AdminRewardLogListResp{}, nil
}
