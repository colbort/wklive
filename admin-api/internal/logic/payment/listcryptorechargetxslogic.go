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

type ListCryptoRechargeTxsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListCryptoRechargeTxsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListCryptoRechargeTxsLogic {
	return &ListCryptoRechargeTxsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListCryptoRechargeTxsLogic) ListCryptoRechargeTxs(req *types.ListCryptoRechargeTxsReq) (resp *types.ListCryptoRechargeTxsResp, err error) {
	return logicutil.Proxy[types.ListCryptoRechargeTxsResp](l.ctx, req, l.svcCtx.PaymentCli.ListCryptoRechargeTxs)
}
