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

type ListUserTradeControlsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListUserTradeControlsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUserTradeControlsLogic {
	return &ListUserTradeControlsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListUserTradeControlsLogic) ListUserTradeControls(req *types.ListUserTradeControlsReq) (resp *types.ListUserTradeControlsResp, err error) {
	return logicutil.Proxy[types.ListUserTradeControlsResp](l.ctx, req, l.svcCtx.TradeCli.ListUserTradeControls)
}
