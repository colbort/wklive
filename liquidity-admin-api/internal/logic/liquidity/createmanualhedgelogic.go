// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package liquidity

import (
	"context"

	"wklive/liquidity-admin-api/internal/svc"
	"wklive/liquidity-admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateManualHedgeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateManualHedgeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateManualHedgeLogic {
	return &CreateManualHedgeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateManualHedgeLogic) CreateManualHedge(req *types.CreateManualHedgeReq) (resp *types.HedgeTaskResp, err error) {
	return createManualHedge(l.ctx, l.svcCtx, req)
}
