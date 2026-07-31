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

type ReviewTradingCalendarLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReviewTradingCalendarLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReviewTradingCalendarLogic {
	return &ReviewTradingCalendarLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReviewTradingCalendarLogic) ReviewTradingCalendar(req *types.ReviewTradingCalendarReq) (resp *types.GetTradingCalendarResp, err error) {
	return logicutil.Proxy[types.GetTradingCalendarResp](
		l.ctx, req, l.svcCtx.OptionCli.ReviewTradingCalendar,
	)
}
