package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCryptoRechargeTxLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCryptoRechargeTxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCryptoRechargeTxLogic {
	return &GetCryptoRechargeTxLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取链上充值交易详情
func (l *GetCryptoRechargeTxLogic) GetCryptoRechargeTx(in *payment.GetCryptoRechargeTxReq) (*payment.GetCryptoRechargeTxResp, error) {
	item, err := l.svcCtx.CryptoRechargeTxModel.FindOneByIdOrHash(l.ctx, in.TenantId, in.Id, int64(in.ChainCode), in.TxHash)
	if err != nil {
		if isNotFound(err) {
			return &payment.GetCryptoRechargeTxResp{Base: helper.ErrResp(i18n.CryptoRechargeTxNotFound, i18n.Translate(i18n.CryptoRechargeTxNotFound, l.ctx))}, nil
		}
		return nil, err
	}
	return &payment.GetCryptoRechargeTxResp{Base: helper.OkResp(), Data: toCryptoRechargeTxProto(item)}, nil
}
