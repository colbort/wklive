// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package option

import (
	"context"

	"wklive/admin-api/internal/logicutil"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RetryLiquidationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRetryLiquidationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RetryLiquidationLogic {
	return &RetryLiquidationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RetryLiquidationLogic) RetryLiquidation(req *types.OptionRetryLiquidationReq) (resp *types.OptionAdminCommonResp, err error) {
	return logicutil.Proxy[types.OptionAdminCommonResp](l.ctx, req, l.svcCtx.OptionCli.RetryLiquidation)
}
