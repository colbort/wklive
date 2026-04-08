package logic

import (
	"context"

	"wklive/proto/staking"
	"wklive/services/staking/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type MyRedeemLogListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMyRedeemLogListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MyRedeemLogListLogic {
	return &MyRedeemLogListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取我的赎回记录列表
func (l *MyRedeemLogListLogic) MyRedeemLogList(in *staking.AppMyRedeemLogListReq) (*staking.AppMyRedeemLogListResp, error) {
	// todo: add your logic here and delete this line

	return &staking.AppMyRedeemLogListResp{}, nil
}
