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

type GetWithdrawOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetWithdrawOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetWithdrawOrderLogic {
	return &GetWithdrawOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetWithdrawOrderLogic) GetWithdrawOrder(req *types.GetWithdrawOrderReq) (resp *types.GetWithdrawOrderResp, err error) {
	return logicutil.Proxy[types.GetWithdrawOrderResp](l.ctx, req, l.svcCtx.PaymentCli.GetWithdrawOrder)
}
