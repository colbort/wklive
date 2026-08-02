// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package liquidity

import (
	"context"

	"wklive/liquidity-admin-api/internal/svc"
	"wklive/liquidity-admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ResolveRiskEventLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewResolveRiskEventLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResolveRiskEventLogic {
	return &ResolveRiskEventLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ResolveRiskEventLogic) ResolveRiskEvent(req *types.ResolveRiskEventReq) (resp *types.RespBase, err error) {
	return resolveRiskEvent(l.ctx, l.svcCtx, req)
}
