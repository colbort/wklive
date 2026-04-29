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

type UpdateCryptoWalletAccountLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateCryptoWalletAccountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCryptoWalletAccountLogic {
	return &UpdateCryptoWalletAccountLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateCryptoWalletAccountLogic) UpdateCryptoWalletAccount(req *types.UpdateCryptoWalletAccountReq) (resp *types.RespBase, err error) {
	return logicutil.Proxy[types.RespBase](l.ctx, req, l.svcCtx.PaymentCli.UpdateCryptoWalletAccount)
}
