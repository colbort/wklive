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

type GetTradeEventListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTradeEventListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTradeEventListLogic {
	return &GetTradeEventListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTradeEventListLogic) GetTradeEventList(req *types.GetTradeEventListReq) (resp *types.GetTradeEventListResp, err error) {
	return logicutil.Proxy[types.GetTradeEventListResp](l.ctx, req, l.svcCtx.TradeCli.GetTradeEventList)
}
