// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package trade

import (
	"context"

	"wklive/admin-api/internal/logicutil"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RetryAccountLiquidationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRetryAccountLiquidationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RetryAccountLiquidationLogic {
	return &RetryAccountLiquidationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RetryAccountLiquidationLogic) RetryAccountLiquidation(req *types.RetryAccountLiquidationReq) (resp *types.CommonResp, err error) {
	return logicutil.Proxy[types.CommonResp](l.ctx, req, l.svcCtx.TradeCli.RetryAccountLiquidation)
}
