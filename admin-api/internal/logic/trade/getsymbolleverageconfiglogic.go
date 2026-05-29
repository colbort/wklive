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

type GetSymbolLeverageConfigLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetSymbolLeverageConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSymbolLeverageConfigLogic {
	return &GetSymbolLeverageConfigLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSymbolLeverageConfigLogic) GetSymbolLeverageConfig(req *types.GetSymbolLeverageConfigReq) (resp *types.GetSymbolLeverageConfigResp, err error) {
	return logicutil.Proxy[types.GetSymbolLeverageConfigResp](l.ctx, req, l.svcCtx.TradeCli.GetSymbolLeverageConfig)
}
