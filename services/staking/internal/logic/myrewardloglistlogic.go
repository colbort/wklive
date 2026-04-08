package logic

import (
	"context"

	"wklive/proto/staking"
	"wklive/services/staking/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type MyRewardLogListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMyRewardLogListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MyRewardLogListLogic {
	return &MyRewardLogListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取我的收益记录列表
func (l *MyRewardLogListLogic) MyRewardLogList(in *staking.AppMyRewardLogListReq) (*staking.AppMyRewardLogListResp, error) {
	// todo: add your logic here and delete this line

	return &staking.AppMyRewardLogListResp{}, nil
}
