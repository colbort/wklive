// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"

	"wklive/app-api/internal/logicutil"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMyRechargeStatLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMyRechargeStatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMyRechargeStatLogic {
	return &GetMyRechargeStatLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMyRechargeStatLogic) GetMyRechargeStat(req *types.GetMyRechargeStatReq) (resp *types.GetMyRechargeStatResp, err error) {
	return logicutil.Proxy[types.GetMyRechargeStatResp](l.ctx, req, l.svcCtx.PaymentCli.GetMyRechargeStat)
}
