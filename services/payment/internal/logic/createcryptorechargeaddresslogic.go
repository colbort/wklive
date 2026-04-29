package logic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateCryptoRechargeAddressLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateCryptoRechargeAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCryptoRechargeAddressLogic {
	return &CreateCryptoRechargeAddressLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建链上充值地址
func (l *CreateCryptoRechargeAddressLogic) CreateCryptoRechargeAddress(in *payment.CreateCryptoRechargeAddressReq) (*payment.AdminCommonResp, error) {
	return createCryptoRechargeAddress(l.ctx, l.svcCtx, in)
}
