package payment

import (
	"context"

	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"

	"wklive/app-api/internal/logicutil"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMyCryptoRechargeAddressLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMyCryptoRechargeAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMyCryptoRechargeAddressLogic {
	return &GetMyCryptoRechargeAddressLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMyCryptoRechargeAddressLogic) GetMyCryptoRechargeAddress(req *types.GetMyCryptoRechargeAddressReq) (*types.GetMyCryptoRechargeAddressResp, error) {
	return logicutil.Proxy[types.GetMyCryptoRechargeAddressResp](l.ctx, req, l.svcCtx.PaymentCli.GetMyCryptoRechargeAddress)
}
