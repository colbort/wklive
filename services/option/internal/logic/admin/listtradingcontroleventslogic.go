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

type ListTradingControlEventsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListTradingControlEventsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTradingControlEventsLogic {
	return &ListTradingControlEventsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分页查询交易控制审计事件
func (l *ListTradingControlEventsLogic) ListTradingControlEvents(in *option.ListTradingControlEventsReq) (*option.ListTradingControlEventsResp, error) {
	tenantId, allowed, forbidden, err := utils.ResolveAdminTenantReadScopeFromMd(l.ctx, in.TenantId)
	if err != nil {
		return nil, err
	}
	if forbidden || !allowed {
		return &option.ListTradingControlEventsResp{
			Base: helper.ErrResp(i18n.PermissionDenied, i18n.Translate(i18n.PermissionDenied, l.ctx)),
		}, nil
	}
	cursor, limit := pageutil.Input(in.Page)
	items, total, err := l.svcCtx.OptionTradingControlEventModel.FindPage(
		l.ctx, models.OptionTradingControlEventPageFilter{
			TenantId: tenantId, UserId: in.UserId, ContractId: in.ContractId,
			EventType: in.EventType, Reason: in.Reason,
		}, cursor, limit,
	)
	if err != nil {
		return nil, err
	}
	data := make([]*option.OptionTradingControlEvent, 0, len(items))
	lastID := int64(0)
	for _, item := range items {
		lastID = item.Id
		data = append(data, helpers.ToTradingControlEventProto(item))
	}
	return &option.ListTradingControlEventsResp{
		Base: pageutil.Base(cursor, limit, len(items), total, lastID), Data: data,
	}, nil
}
