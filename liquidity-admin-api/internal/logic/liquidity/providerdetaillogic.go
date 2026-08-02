// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package liquidity

import (
	"context"

	"wklive/liquidity-admin-api/internal/svc"
	"wklive/liquidity-admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ProviderDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewProviderDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProviderDetailLogic {
	return &ProviderDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ProviderDetailLogic) ProviderDetail(req *types.ProviderDetailReq) (resp *types.ProviderDetailResp, err error) {
	return providerDetail(l.ctx, l.svcCtx, req)
}
