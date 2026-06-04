// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"wklive/admin-api/internal/logicutil"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserRechargeStatLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserRechargeStatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserRechargeStatLogic {
	return &GetUserRechargeStatLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserRechargeStatLogic) GetUserRechargeStat(req *types.GetUserRechargeStatReq) (resp *types.GetUserRechargeStatResp, err error) {
	return logicutil.Proxy[types.GetUserRechargeStatResp](l.ctx, req, l.svcCtx.PaymentCli.GetUserRechargeStat)
}
