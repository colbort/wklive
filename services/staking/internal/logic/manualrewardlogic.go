package logic

import (
	"context"

	"wklive/proto/staking"
	"wklive/services/staking/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ManualRewardLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewManualRewardLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ManualRewardLogic {
	return &ManualRewardLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 手动发放收益
func (l *ManualRewardLogic) ManualReward(in *staking.AdminManualRewardReq) (*staking.AdminManualRewardResp, error) {
	// todo: add your logic here and delete this line

	return &staking.AdminManualRewardResp{}, nil
}
