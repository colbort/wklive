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

type CreateTradingCalendarLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateTradingCalendarLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTradingCalendarLogic {
	return &CreateTradingCalendarLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateTradingCalendarLogic) CreateTradingCalendar(req *types.CreateTradingCalendarReq) (resp *types.GetTradingCalendarResp, err error) {
	return logicutil.Proxy[types.GetTradingCalendarResp](
		l.ctx, req, l.svcCtx.OptionCli.CreateTradingCalendar,
	)
}
