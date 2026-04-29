package logic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateCryptoRechargeTxLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateCryptoRechargeTxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCryptoRechargeTxLogic {
	return &CreateCryptoRechargeTxLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建链上充值交易
func (l *CreateCryptoRechargeTxLogic) CreateCryptoRechargeTx(in *payment.CreateCryptoRechargeTxReq) (*payment.AdminCommonResp, error) {
	return createCryptoRechargeTx(l.ctx, l.svcCtx, in)
}
