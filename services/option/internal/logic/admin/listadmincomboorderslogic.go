package adminlogic

import (
	"context"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/pageutil"
	"wklive/common/utils"
	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAdminComboOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListAdminComboOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAdminComboOrdersLogic {
	return &ListAdminComboOrdersLogic{
		ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx),
	}
}

// 分页查询组合父单，供运营按整组处理
func (l *ListAdminComboOrdersLogic) ListAdminComboOrders(
	in *option.ListAdminComboOrdersReq,
) (*option.ListAdminComboOrdersResp, error) {
	if in.TenantId <= 0 {
		return &option.ListAdminComboOrdersResp{
			Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx)),
		}, nil
	}
	_, allowed, forbidden, err := utils.ResolveAdminTenantWriteScopeFromMd(l.ctx, in.TenantId)
	if err != nil {
		return nil, err
	}
	if forbidden || !allowed {
		return &option.ListAdminComboOrdersResp{
			Base: helper.ErrResp(i18n.PermissionDenied, i18n.Translate(i18n.PermissionDenied, l.ctx)),
		}, nil
	}
	cursor, limit := pageutil.Input(in.Page)
	items, total, err := l.svcCtx.OptionComboOrderModel.FindPage(
		l.ctx,
		models.OptionComboOrderPageFilter{
			TenantId: in.TenantId, UserId: in.UserId, AccountId: in.AccountId,
			ComboNo: in.ComboNo, UnderlyingSymbol: in.UnderlyingSymbol,
			Status:          int64(in.Status),
			CreateTimeStart: pageutil.TimeRangeStart(in.CreateTimeRange),
			CreateTimeEnd:   pageutil.TimeRangeEnd(in.CreateTimeRange),
		},
		cursor, limit,
	)
	if err != nil {
		return nil, err
	}
	data := make([]*option.OptionComboOrderDetail, 0, len(items))
	lastID := int64(0)
	for _, item := range items {
		lastID = item.Id
		detail, buildErr := buildAdminComboSummary(l.ctx, l.svcCtx, item)
		if buildErr != nil {
			return nil, buildErr
		}
		data = append(data, detail)
	}
	return &option.ListAdminComboOrdersResp{
		Base: pageutil.Base(cursor, limit, len(items), total, lastID),
		Data: data, Total: total,
	}, nil
}
