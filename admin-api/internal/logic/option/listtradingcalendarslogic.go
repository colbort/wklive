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

type ListTradingCalendarsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListTradingCalendarsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTradingCalendarsLogic {
	return &ListTradingCalendarsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListTradingCalendarsLogic) ListTradingCalendars(req *types.ListTradingCalendarsReq) (resp *types.ListTradingCalendarsResp, err error) {
	return logicutil.Proxy[types.ListTradingCalendarsResp](
		l.ctx, req, l.svcCtx.OptionCli.ListTradingCalendars,
	)
}
