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

type GetUserLeverageConfigLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserLeverageConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserLeverageConfigLogic {
	return &GetUserLeverageConfigLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserLeverageConfigLogic) GetUserLeverageConfig(req *types.GetUserLeverageConfigReq) (resp *types.GetUserLeverageConfigResp, err error) {
	return logicutil.Proxy[types.GetUserLeverageConfigResp](l.ctx, req, l.svcCtx.TradeCli.GetUserLeverageConfig)
}
