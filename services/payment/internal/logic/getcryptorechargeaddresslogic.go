package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCryptoRechargeAddressLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCryptoRechargeAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCryptoRechargeAddressLogic {
	return &GetCryptoRechargeAddressLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取链上充值地址详情
func (l *GetCryptoRechargeAddressLogic) GetCryptoRechargeAddress(in *payment.GetCryptoRechargeAddressReq) (*payment.GetCryptoRechargeAddressResp, error) {
	item, err := getCryptoRechargeAddress(l.ctx, l.svcCtx, in.TenantId, in.Id)
	if err != nil {
		if isNotFound(err) {
			return &payment.GetCryptoRechargeAddressResp{Base: helper.ErrResp(i18n.CryptoRechargeAddressNotFound, i18n.Translate(i18n.CryptoRechargeAddressNotFound, l.ctx))}, nil
		}
		return nil, err
	}
	if item == nil {
		return &payment.GetCryptoRechargeAddressResp{Base: helper.ErrResp(i18n.CryptoRechargeAddressNotFound, i18n.Translate(i18n.CryptoRechargeAddressNotFound, l.ctx))}, nil
	}
	return &payment.GetCryptoRechargeAddressResp{Base: helper.OkResp(), Data: toCryptoRechargeAddressProto(item)}, nil
}
