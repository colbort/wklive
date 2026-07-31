// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package option

import (
	"context"

	"wklive/admin-api/internal/logicutil"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListTradingControlEventsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListTradingControlEventsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTradingControlEventsLogic {
	return &ListTradingControlEventsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListTradingControlEventsLogic) ListTradingControlEvents(req *types.ListTradingControlEventsReq) (resp *types.ListTradingControlEventsResp, err error) {
	return logicutil.Proxy[types.ListTradingControlEventsResp](
		l.ctx, req, l.svcCtx.OptionCli.ListTradingControlEvents,
	)
}
