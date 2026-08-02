// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package liquidity

import (
	"context"

	"wklive/liquidity-admin-api/internal/svc"
	"wklive/liquidity-admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CancelHedgeTaskLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCancelHedgeTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelHedgeTaskLogic {
	return &CancelHedgeTaskLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CancelHedgeTaskLogic) CancelHedgeTask(req *types.HedgeTaskActionReq) (resp *types.RespBase, err error) {
	return cancelHedgeTask(l.ctx, l.svcCtx, req)
}
