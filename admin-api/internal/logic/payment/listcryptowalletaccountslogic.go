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

type ListCryptoWalletAccountsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListCryptoWalletAccountsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListCryptoWalletAccountsLogic {
	return &ListCryptoWalletAccountsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListCryptoWalletAccountsLogic) ListCryptoWalletAccounts(req *types.ListCryptoWalletAccountsReq) (resp *types.ListCryptoWalletAccountsResp, err error) {
	return logicutil.Proxy[types.ListCryptoWalletAccountsResp](l.ctx, req, l.svcCtx.PaymentCli.ListCryptoWalletAccounts)
}
