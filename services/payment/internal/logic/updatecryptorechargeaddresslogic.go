package logic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateCryptoRechargeAddressLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateCryptoRechargeAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCryptoRechargeAddressLogic {
	return &UpdateCryptoRechargeAddressLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新链上充值地址
func (l *UpdateCryptoRechargeAddressLogic) UpdateCryptoRechargeAddress(in *payment.UpdateCryptoRechargeAddressReq) (*payment.AdminCommonResp, error) {
	return updateCryptoRechargeAddress(l.ctx, l.svcCtx, in)
}
