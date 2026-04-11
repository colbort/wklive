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

type GetUserTradeLimitLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserTradeLimitLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserTradeLimitLogic {
	return &GetUserTradeLimitLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserTradeLimitLogic) GetUserTradeLimit(req *types.GetUserTradeLimitReq) (resp *types.GetUserTradeLimitResp, err error) {
	return logicutil.Proxy[types.GetUserTradeLimitResp](l.ctx, req, l.svcCtx.TradeCli.GetUserTradeLimit)
}
