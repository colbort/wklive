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

type GetLeverageConfigLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetLeverageConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLeverageConfigLogic {
	return &GetLeverageConfigLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetLeverageConfigLogic) GetLeverageConfig(req *types.GetLeverageConfigReq) (resp *types.GetLeverageConfigResp, err error) {
	return logicutil.Proxy[types.GetLeverageConfigResp](l.ctx, req, l.svcCtx.TradeCli.GetLeverageConfig)
}
