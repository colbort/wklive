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

type GetUserTradeConfigLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserTradeConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserTradeConfigLogic {
	return &GetUserTradeConfigLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserTradeConfigLogic) GetUserTradeConfig(req *types.GetUserTradeConfigReq) (resp *types.GetUserTradeConfigResp, err error) {
	return logicutil.Proxy[types.GetUserTradeConfigResp](l.ctx, req, l.svcCtx.TradeCli.GetUserTradeConfig)
}
