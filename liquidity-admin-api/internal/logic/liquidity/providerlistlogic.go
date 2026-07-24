// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package liquidity

import (
	"context"

	"wklive/liquidity-admin-api/internal/svc"
	"wklive/liquidity-admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ProviderListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewProviderListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProviderListLogic {
	return &ProviderListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ProviderListLogic) ProviderList(req *types.PageQuery) (resp *types.ProviderListResp, err error) {
	req.Limit = listLimit(req.Limit)
	return providerList(l.ctx, l.svcCtx, req)
}
