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

type GetCryptoWalletAccountLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetCryptoWalletAccountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCryptoWalletAccountLogic {
	return &GetCryptoWalletAccountLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCryptoWalletAccountLogic) GetCryptoWalletAccount(req *types.GetCryptoWalletAccountReq) (resp *types.GetCryptoWalletAccountResp, err error) {
	return logicutil.Proxy[types.GetCryptoWalletAccountResp](l.ctx, req, l.svcCtx.PaymentCli.GetCryptoWalletAccount)
}
