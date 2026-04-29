// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/admin-api/internal/logicutil"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCryptoRechargeAddressLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetCryptoRechargeAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCryptoRechargeAddressLogic {
	return &GetCryptoRechargeAddressLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCryptoRechargeAddressLogic) GetCryptoRechargeAddress(req *types.GetCryptoRechargeAddressReq) (resp *types.GetCryptoRechargeAddressResp, err error) {
	return logicutil.Proxy[types.GetCryptoRechargeAddressResp](l.ctx, req, l.svcCtx.PaymentCli.GetCryptoRechargeAddress)
}
