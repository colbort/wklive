// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package liquidity

import (
	"context"

	"wklive/liquidity-admin-api/internal/svc"
	"wklive/liquidity-admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ProviderUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewProviderUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProviderUpdateLogic {
	return &ProviderUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ProviderUpdateLogic) ProviderUpdate(req *types.UpdateProviderReq) (resp *types.ProviderDetailResp, err error) {
	return providerUpdate(l.ctx, l.svcCtx, req)
}
