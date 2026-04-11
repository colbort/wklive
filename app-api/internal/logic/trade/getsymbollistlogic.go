// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package trade

import (
	"context"

	"wklive/app-api/internal/logicutil"
	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSymbolListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetSymbolListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSymbolListLogic {
	return &GetSymbolListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSymbolListLogic) GetSymbolList(req *types.GetSymbolListReq) (resp *types.GetSymbolListResp, err error) {
	return logicutil.Proxy[types.GetSymbolListResp](l.ctx, req, l.svcCtx.TradeCli.GetSymbolList)
}
