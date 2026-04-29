package logic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListCryptoRechargeAddressesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListCryptoRechargeAddressesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListCryptoRechargeAddressesLogic {
	return &ListCryptoRechargeAddressesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 链上充值地址列表
func (l *ListCryptoRechargeAddressesLogic) ListCryptoRechargeAddresses(in *payment.ListCryptoRechargeAddressesReq) (*payment.ListCryptoRechargeAddressesResp, error) {
	return listCryptoRechargeAddresses(l.ctx, l.svcCtx, in)
}
