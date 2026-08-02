// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package liquidity

import (
	"context"

	"wklive/liquidity-admin-api/internal/svc"
	"wklive/liquidity-admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ResolveReconcileDifferenceLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewResolveReconcileDifferenceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResolveReconcileDifferenceLogic {
	return &ResolveReconcileDifferenceLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ResolveReconcileDifferenceLogic) ResolveReconcileDifference(req *types.ResolveReconcileDifferenceReq) (resp *types.RespBase, err error) {
	return resolveReconcileDifference(l.ctx, l.svcCtx, req)
}
