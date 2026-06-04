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

type CreateCryptoRechargeOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateCryptoRechargeOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCryptoRechargeOrderLogic {
	return &CreateCryptoRechargeOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateCryptoRechargeOrderLogic) CreateCryptoRechargeOrder(req *types.CreateCryptoRechargeOrderReq) (*types.CreateCryptoRechargeOrderResp, error) {
	return logicutil.Proxy[types.CreateCryptoRechargeOrderResp](l.ctx, req, l.svcCtx.PaymentCli.CreateCryptoRechargeOrder)
}
