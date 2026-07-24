// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package liquidity

import (
	"context"

	"wklive/liquidity-admin-api/internal/svc"
	"wklive/liquidity-admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SymbolConfigListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSymbolConfigListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SymbolConfigListLogic {
	return &SymbolConfigListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SymbolConfigListLogic) SymbolConfigList(req *types.PageQuery) (resp *types.SymbolConfigListResp, err error) {
	req.Limit = listLimit(req.Limit)
	return symbolConfigList(l.ctx, l.svcCtx, req)
}
