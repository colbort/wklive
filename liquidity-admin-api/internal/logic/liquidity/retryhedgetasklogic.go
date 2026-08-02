// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package liquidity

import (
	"context"

	"wklive/liquidity-admin-api/internal/svc"
	"wklive/liquidity-admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RetryHedgeTaskLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRetryHedgeTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RetryHedgeTaskLogic {
	return &RetryHedgeTaskLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RetryHedgeTaskLogic) RetryHedgeTask(req *types.HedgeTaskActionReq) (resp *types.RespBase, err error) {
	return retryHedgeTask(l.ctx, l.svcCtx, req)
}
