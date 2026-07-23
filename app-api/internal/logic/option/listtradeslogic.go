// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package option

import (
	"context"

	"wklive/app-api/internal/logicutil"
	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListTradesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListTradesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTradesLogic {
	return &ListTradesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListTradesLogic) ListTrades(req *types.ListTradesReq) (resp *types.ListTradesResp, err error) {
	return logicutil.Proxy[types.ListTradesResp](l.ctx, req, l.svcCtx.OptionCli.ListTrades)
}
