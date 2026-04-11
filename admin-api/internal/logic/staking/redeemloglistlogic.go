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

type RedeemLogListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRedeemLogListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RedeemLogListLogic {
	return &RedeemLogListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RedeemLogListLogic) RedeemLogList(req *types.AdminRedeemLogListReq) (resp *types.AdminRedeemLogListResp, err error) {
	return logicutil.Proxy[types.AdminRedeemLogListResp](l.ctx, req, l.svcCtx.StakingCli.RedeemLogList)
}
