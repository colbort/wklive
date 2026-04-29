package logic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateCryptoWalletAccountLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateCryptoWalletAccountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCryptoWalletAccountLogic {
	return &CreateCryptoWalletAccountLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建链上钱包账号
func (l *CreateCryptoWalletAccountLogic) CreateCryptoWalletAccount(in *payment.CreateCryptoWalletAccountReq) (*payment.AdminCommonResp, error) {
	return createCryptoWalletAccount(l.ctx, l.svcCtx, in)
}
