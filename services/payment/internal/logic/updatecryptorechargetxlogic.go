package logic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateCryptoRechargeTxLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateCryptoRechargeTxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCryptoRechargeTxLogic {
	return &UpdateCryptoRechargeTxLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新链上充值交易
func (l *UpdateCryptoRechargeTxLogic) UpdateCryptoRechargeTx(in *payment.UpdateCryptoRechargeTxReq) (*payment.AdminCommonResp, error) {
	return updateCryptoRechargeTx(l.ctx, l.svcCtx, in)
}
