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

type UpdateCryptoRechargeAddressLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateCryptoRechargeAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCryptoRechargeAddressLogic {
	return &UpdateCryptoRechargeAddressLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateCryptoRechargeAddressLogic) UpdateCryptoRechargeAddress(req *types.UpdateCryptoRechargeAddressReq) (resp *types.RespBase, err error) {
	return logicutil.Proxy[types.RespBase](l.ctx, req, l.svcCtx.PaymentCli.UpdateCryptoRechargeAddress)
}
