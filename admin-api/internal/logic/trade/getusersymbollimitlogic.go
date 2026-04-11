// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package trade

import (
	"context"

	"wklive/admin-api/internal/logicutil"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserSymbolLimitLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserSymbolLimitLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserSymbolLimitLogic {
	return &GetUserSymbolLimitLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserSymbolLimitLogic) GetUserSymbolLimit(req *types.GetUserSymbolLimitReq) (resp *types.GetUserSymbolLimitResp, err error) {
	return logicutil.Proxy[types.GetUserSymbolLimitResp](l.ctx, req, l.svcCtx.TradeCli.GetUserSymbolLimit)
}
