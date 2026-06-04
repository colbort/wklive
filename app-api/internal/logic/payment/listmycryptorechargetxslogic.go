package payment

import (
	"context"

	"wklive/app-api/internal/logicutil"

	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMyCryptoRechargeTxsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListMyCryptoRechargeTxsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMyCryptoRechargeTxsLogic {
	return &ListMyCryptoRechargeTxsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListMyCryptoRechargeTxsLogic) ListMyCryptoRechargeTxs(req *types.ListMyCryptoRechargeTxsReq) (*types.ListMyCryptoRechargeTxsResp, error) {
	return logicutil.Proxy[types.ListMyCryptoRechargeTxsResp](l.ctx, req, l.svcCtx.PaymentCli.ListMyCryptoRechargeTxs)
}
