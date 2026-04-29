package logic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListCryptoWalletAccountsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListCryptoWalletAccountsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListCryptoWalletAccountsLogic {
	return &ListCryptoWalletAccountsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 链上钱包账号列表
func (l *ListCryptoWalletAccountsLogic) ListCryptoWalletAccounts(in *payment.ListCryptoWalletAccountsReq) (*payment.ListCryptoWalletAccountsResp, error) {
	return listCryptoWalletAccounts(l.ctx, l.svcCtx, in)
}
