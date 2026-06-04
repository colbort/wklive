package payment

import (
	"context"

	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"

	"wklive/app-api/internal/logicutil"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMyCryptoRechargeTxLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMyCryptoRechargeTxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMyCryptoRechargeTxLogic {
	return &GetMyCryptoRechargeTxLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMyCryptoRechargeTxLogic) GetMyCryptoRechargeTx(req *types.GetMyCryptoRechargeTxReq) (*types.GetMyCryptoRechargeTxResp, error) {
	return logicutil.Proxy[types.GetMyCryptoRechargeTxResp](l.ctx, req, l.svcCtx.PaymentCli.GetMyCryptoRechargeTx)
}
