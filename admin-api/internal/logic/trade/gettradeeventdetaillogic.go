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

type GetTradeEventDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTradeEventDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTradeEventDetailLogic {
	return &GetTradeEventDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTradeEventDetailLogic) GetTradeEventDetail(req *types.GetTradeEventDetailReq) (resp *types.GetTradeEventDetailResp, err error) {
	return logicutil.Proxy[types.GetTradeEventDetailResp](l.ctx, req, l.svcCtx.TradeCli.GetTradeEventDetail)
}
