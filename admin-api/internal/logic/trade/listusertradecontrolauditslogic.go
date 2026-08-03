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

type ListUserTradeControlAuditsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListUserTradeControlAuditsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUserTradeControlAuditsLogic {
	return &ListUserTradeControlAuditsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListUserTradeControlAuditsLogic) ListUserTradeControlAudits(req *types.ListUserTradeControlAuditsReq) (resp *types.ListUserTradeControlAuditsResp, err error) {
	return logicutil.Proxy[types.ListUserTradeControlAuditsResp](l.ctx, req, l.svcCtx.TradeCli.ListUserTradeControlAudits)
}
