package logic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateCryptoWalletAccountLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateCryptoWalletAccountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCryptoWalletAccountLogic {
	return &UpdateCryptoWalletAccountLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新链上钱包账号
func (l *UpdateCryptoWalletAccountLogic) UpdateCryptoWalletAccount(in *payment.UpdateCryptoWalletAccountReq) (*payment.AdminCommonResp, error) {
	return updateCryptoWalletAccount(l.ctx, l.svcCtx, in)
}
