// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/admin-api/internal/logicutil"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateCryptoWalletAccountLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateCryptoWalletAccountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCryptoWalletAccountLogic {
	return &CreateCryptoWalletAccountLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateCryptoWalletAccountLogic) CreateCryptoWalletAccount(req *types.CreateCryptoWalletAccountReq) (resp *types.RespBase, err error) {
	return logicutil.Proxy[types.RespBase](l.ctx, req, l.svcCtx.PaymentCli.CreateCryptoWalletAccount)
}
