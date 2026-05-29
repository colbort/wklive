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

type GetSymbolLeverageConfigListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetSymbolLeverageConfigListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSymbolLeverageConfigListLogic {
	return &GetSymbolLeverageConfigListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSymbolLeverageConfigListLogic) GetSymbolLeverageConfigList(req *types.GetSymbolLeverageConfigListReq) (resp *types.GetSymbolLeverageConfigListResp, err error) {
	return logicutil.Proxy[types.GetSymbolLeverageConfigListResp](l.ctx, req, l.svcCtx.TradeCli.GetSymbolLeverageConfigList)
}
