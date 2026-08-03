// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package staking

import (
	"context"

	"wklive/admin-api/internal/logicutil"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type OperationRetryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOperationRetryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OperationRetryLogic {
	return &OperationRetryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OperationRetryLogic) OperationRetry(req *types.OperationRetryReq) (resp *types.OperationRetryResp, err error) {
	return logicutil.Proxy[types.OperationRetryResp](l.ctx, req, l.svcCtx.StakingCli.OperationRetry)
}
