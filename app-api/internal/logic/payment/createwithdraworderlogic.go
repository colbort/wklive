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

type CreateWithdrawOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateWithdrawOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateWithdrawOrderLogic {
	return &CreateWithdrawOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateWithdrawOrderLogic) CreateWithdrawOrder(req *types.CreateWithdrawOrderReq) (resp *types.CreateWithdrawOrderResp, err error) {
	return logicutil.Proxy[types.CreateWithdrawOrderResp](l.ctx, req, l.svcCtx.PaymentCli.CreateWithdrawOrder)
}
