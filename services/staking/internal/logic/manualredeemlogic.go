package logic

import (
	"context"

	"wklive/proto/staking"
	"wklive/services/staking/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ManualRedeemLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewManualRedeemLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ManualRedeemLogic {
	return &ManualRedeemLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 手动赎回
func (l *ManualRedeemLogic) ManualRedeem(in *staking.AdminManualRedeemReq) (*staking.AdminManualRedeemResp, error) {
	// todo: add your logic here and delete this line

	return &staking.AdminManualRedeemResp{}, nil
}
