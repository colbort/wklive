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

type ListCryptoRechargeAddressesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListCryptoRechargeAddressesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListCryptoRechargeAddressesLogic {
	return &ListCryptoRechargeAddressesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListCryptoRechargeAddressesLogic) ListCryptoRechargeAddresses(req *types.ListCryptoRechargeAddressesReq) (resp *types.ListCryptoRechargeAddressesResp, err error) {
	return logicutil.Proxy[types.ListCryptoRechargeAddressesResp](l.ctx, req, l.svcCtx.PaymentCli.ListCryptoRechargeAddresses)
}
