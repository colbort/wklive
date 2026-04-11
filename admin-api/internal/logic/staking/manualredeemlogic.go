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

type ManualRedeemLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewManualRedeemLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ManualRedeemLogic {
	return &ManualRedeemLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ManualRedeemLogic) ManualRedeem(req *types.AdminManualRedeemReq) (resp *types.AdminManualRedeemResp, err error) {
	return logicutil.Proxy[types.AdminManualRedeemResp](l.ctx, req, l.svcCtx.StakingCli.ManualRedeem)
}
