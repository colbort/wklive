package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCryptoWalletAccountLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCryptoWalletAccountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCryptoWalletAccountLogic {
	return &GetCryptoWalletAccountLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取链上钱包账号详情
func (l *GetCryptoWalletAccountLogic) GetCryptoWalletAccount(in *payment.GetCryptoWalletAccountReq) (*payment.GetCryptoWalletAccountResp, error) {
	item, err := l.svcCtx.CryptoWalletAccountModel.FindOne(l.ctx, in.Id)
	if err != nil {
		if isNotFound(err) {
			return &payment.GetCryptoWalletAccountResp{Base: helper.ErrResp(i18n.CryptoWalletAccountNotFound, i18n.Translate(i18n.CryptoWalletAccountNotFound, l.ctx))}, nil
		}
		return nil, err
	}
	if in.TenantId > 0 && item.TenantId != in.TenantId {
		return &payment.GetCryptoWalletAccountResp{Base: helper.ErrResp(i18n.CryptoWalletAccountNotFound, i18n.Translate(i18n.CryptoWalletAccountNotFound, l.ctx))}, nil
	}
	return &payment.GetCryptoWalletAccountResp{Base: helper.OkResp(), Data: toCryptoWalletAccountProto(item)}, nil
}
