package logic

import (
	"context"

	"wklive/proto/staking"
	"wklive/services/staking/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type RedeemLogListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRedeemLogListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RedeemLogListLogic {
	return &RedeemLogListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取赎回记录列表
func (l *RedeemLogListLogic) RedeemLogList(in *staking.AdminRedeemLogListReq) (*staking.AdminRedeemLogListResp, error) {
	// todo: add your logic here and delete this line

	return &staking.AdminRedeemLogListResp{}, nil
}
