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

type GetCryptoRechargeTxLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetCryptoRechargeTxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCryptoRechargeTxLogic {
	return &GetCryptoRechargeTxLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCryptoRechargeTxLogic) GetCryptoRechargeTx(req *types.GetCryptoRechargeTxReq) (resp *types.GetCryptoRechargeTxResp, err error) {
	return logicutil.Proxy[types.GetCryptoRechargeTxResp](l.ctx, req, l.svcCtx.PaymentCli.GetCryptoRechargeTx)
}
