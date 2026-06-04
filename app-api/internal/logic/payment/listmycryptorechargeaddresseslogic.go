package payment

import (
	"context"

	"wklive/app-api/internal/logicutil"

	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMyCryptoRechargeAddressesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListMyCryptoRechargeAddressesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMyCryptoRechargeAddressesLogic {
	return &ListMyCryptoRechargeAddressesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListMyCryptoRechargeAddressesLogic) ListMyCryptoRechargeAddresses(req *types.ListMyCryptoRechargeAddressesReq) (*types.ListMyCryptoRechargeAddressesResp, error) {
	return logicutil.Proxy[types.ListMyCryptoRechargeAddressesResp](l.ctx, req, l.svcCtx.PaymentCli.ListMyCryptoRechargeAddresses)
}
