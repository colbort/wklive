package adminlogic

import (
	"context"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/pageutil"
	"wklive/common/utils"
	"wklive/proto/option"
	"wklive/services/option/internal/logic/helpers"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListTradingCalendarsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListTradingCalendarsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTradingCalendarsLogic {
	return &ListTradingCalendarsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分页查询交易日历版本
func (l *ListTradingCalendarsLogic) ListTradingCalendars(in *option.ListTradingCalendarsReq) (*option.ListTradingCalendarsResp, error) {
	tenantId, allowed, forbidden, err := utils.ResolveAdminTenantReadScopeFromMd(l.ctx, in.TenantId)
	if err != nil {
		return nil, err
	}
	if forbidden || !allowed {
		return &option.ListTradingCalendarsResp{
			Base: helper.ErrResp(i18n.PermissionDenied, i18n.Translate(i18n.PermissionDenied, l.ctx)),
		}, nil
	}
	code := ""
	if in.CalendarCode != "" {
		var valid bool
		code, valid = helpers.NormalizeTradingCalendarCode(in.CalendarCode)
		if !valid {
			return &option.ListTradingCalendarsResp{
				Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx)),
			}, nil
		}
	}
	cursor, limit := pageutil.Input(in.Page)
	items, total, err := l.svcCtx.OptionTradingCalendarModel.FindPage(
		l.ctx, models.OptionTradingCalendarPageFilter{
			TenantId: tenantId, CalendarCode: code, Status: int64(in.Status),
		}, cursor, limit,
	)
	if err != nil {
		return nil, err
	}
	data := make([]*option.OptionTradingCalendar, 0, len(items))
	lastID := int64(0)
	for _, item := range items {
		lastID = item.Id
		sessions, findErr := l.svcCtx.OptionTradingCalendarSessionModel.FindByCalendar(
			l.ctx, item.TenantId, item.Id,
		)
		if findErr != nil {
			return nil, findErr
		}
		exceptions, findErr := l.svcCtx.OptionTradingCalendarExceptionModel.FindByCalendar(
			l.ctx, item.TenantId, item.Id,
		)
		if findErr != nil {
			return nil, findErr
		}
		data = append(data, helpers.ToTradingCalendarProto(item, sessions, exceptions))
	}
	return &option.ListTradingCalendarsResp{
		Base: pageutil.Base(cursor, limit, len(items), total, lastID),
		Data: data, Total: total,
	}, nil
}
