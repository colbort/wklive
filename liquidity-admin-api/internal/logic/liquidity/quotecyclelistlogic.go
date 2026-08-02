// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package liquidity

import (
	"context"

	"wklive/liquidity-admin-api/internal/svc"
	"wklive/liquidity-admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type QuoteCycleListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQuoteCycleListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QuoteCycleListLogic {
	return &QuoteCycleListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QuoteCycleListLogic) QuoteCycleList(req *types.QuoteCycleQuery) (resp *types.QuoteCycleListResp, err error) {
	return quoteCycleList(l.ctx, l.svcCtx, req)
}
