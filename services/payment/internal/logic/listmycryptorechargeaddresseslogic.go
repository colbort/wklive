package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/common/utils"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMyCryptoRechargeAddressesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListMyCryptoRechargeAddressesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMyCryptoRechargeAddressesLogic {
	return &ListMyCryptoRechargeAddressesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 我的链上充值地址列表
func (l *ListMyCryptoRechargeAddressesLogic) ListMyCryptoRechargeAddresses(in *payment.ListMyCryptoRechargeAddressesReq) (*payment.ListMyCryptoRechargeAddressesResp, error) {
	userId, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	tenantId, err := utils.GetTenantIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	items, _, err := l.svcCtx.CryptoRechargeAddressModel.FindPage(l.ctx, tenantId, userId, int64(in.WalletType), in.Coin, int64(in.ChainCode), "", int64(payment.CryptoRechargeAddressStatus_CRYPTO_RECHARGE_ADDRESS_STATUS_ENABLED), 0, 0, 100)
	if err != nil {
		return nil, err
	}
	data := make([]*payment.CryptoRechargeAddress, 0, len(items))
	for _, item := range items {
		data = append(data, toCryptoRechargeAddressProto(item))
	}
	return &payment.ListMyCryptoRechargeAddressesResp{Base: helper.OkResp(), Data: data}, nil
}
