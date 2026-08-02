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

type ListTradingHaltsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListTradingHaltsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTradingHaltsLogic {
	return &ListTradingHaltsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分页查询临时休市记录
func (l *ListTradingHaltsLogic) ListTradingHalts(in *option.ListTradingHaltsReq) (*option.ListTradingHaltsResp, error) {
	tenantId, allowed, forbidden, err := utils.ResolveAdminTenantReadScopeFromMd(l.ctx, in.TenantId)
	if err != nil {
		return nil, err
	}
	if forbidden || !allowed {
		return &option.ListTradingHaltsResp{
			Base: helper.ErrResp(i18n.PermissionDenied, i18n.Translate(i18n.PermissionDenied, l.ctx)),
		}, nil
	}
	cursor, limit := pageutil.Input(in.Page)
	items, total, err := l.svcCtx.OptionTradingHaltModel.FindPage(
		l.ctx, models.OptionTradingHaltPageFilter{
			TenantId: tenantId, ContractId: in.ContractId, Status: int64(in.Status),
		}, cursor, limit,
	)
	if err != nil {
		return nil, err
	}
	data := make([]*option.OptionTradingHalt, 0, len(items))
	lastID := int64(0)
	for _, item := range items {
		lastID = item.Id
		data = append(data, helpers.ToTradingHaltProto(item))
	}
	return &option.ListTradingHaltsResp{
		Base: pageutil.Base(cursor, limit, len(items), total, lastID),
		Data: data, Total: total,
	}, nil
}
