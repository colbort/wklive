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

type ListCorporateActionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListCorporateActionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListCorporateActionsLogic {
	return &ListCorporateActionsLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// 分页查询公司行动及合约迁移进度
func (l *ListCorporateActionsLogic) ListCorporateActions(
	in *option.ListCorporateActionsReq,
) (*option.ListCorporateActionsResp, error) {
	tenantId, allowed, forbidden, err := utils.ResolveAdminTenantReadScopeFromMd(l.ctx, in.TenantId)
	if err != nil {
		return nil, err
	}
	if forbidden || !allowed {
		return &option.ListCorporateActionsResp{
			Base: helper.ErrResp(i18n.PermissionDenied, i18n.Translate(i18n.PermissionDenied, l.ctx)),
		}, nil
	}
	cursor, limit := pageutil.Input(in.Page)
	items, total, err := l.svcCtx.OptionCorporateActionModel.FindPage(l.ctx, models.OptionCorporateActionPageFilter{
		TenantId: tenantId, UnderlyingSymbol: in.UnderlyingSymbol,
		ActionType: int64(in.ActionType), Status: int64(in.Status),
	}, cursor, limit)
	if err != nil {
		return nil, err
	}
	result := make([]*option.OptionCorporateAction, 0, len(items))
	for _, item := range items {
		mappings, findErr := l.svcCtx.OptionCorporateActionContractModel.FindByAction(l.ctx, item.TenantId, item.Id)
		if findErr != nil {
			return nil, findErr
		}
		result = append(result, helpers.ToCorporateActionProto(item, mappings))
	}
	lastID := int64(0)
	if len(items) > 0 {
		lastID = items[len(items)-1].Id
	}
	return &option.ListCorporateActionsResp{
		Base: pageutil.Base(cursor, limit, len(items), total, lastID), Data: result, Total: total,
	}, nil
}
