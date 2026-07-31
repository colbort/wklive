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

	"github.com/zeromicro/go-zero/core/logx"
)

type ListCorporateActionPositionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListCorporateActionPositionsLogic(
	ctx context.Context, svcCtx *svc.ServiceContext,
) *ListCorporateActionPositionsLogic {
	return &ListCorporateActionPositionsLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// 分页查询逐持仓迁移审计结果
func (l *ListCorporateActionPositionsLogic) ListCorporateActionPositions(
	in *option.ListCorporateActionPositionsReq,
) (*option.ListCorporateActionPositionsResp, error) {
	_, allowed, forbidden, err := utils.ResolveAdminTenantWriteScopeFromMd(l.ctx, in.TenantId)
	if err != nil {
		return nil, err
	}
	if forbidden || !allowed {
		return &option.ListCorporateActionPositionsResp{
			Base: helper.ErrResp(i18n.PermissionDenied, i18n.Translate(i18n.PermissionDenied, l.ctx)),
		}, nil
	}
	cursor, limit := pageutil.Input(in.Page)
	items, total, err := l.svcCtx.OptionCorporateActionPositionModel.FindPage(
		l.ctx, in.TenantId, in.ActionId, in.ActionContractId, int64(in.Status), cursor, limit,
	)
	if err != nil {
		return nil, err
	}
	result := make([]*option.OptionCorporateActionPosition, 0, len(items))
	for _, item := range items {
		result = append(result, helpers.ToCorporateActionPositionProto(item))
	}
	lastID := int64(0)
	if len(items) > 0 {
		lastID = items[len(items)-1].Id
	}
	return &option.ListCorporateActionPositionsResp{
		Base: pageutil.Base(cursor, limit, len(items), total, lastID), Data: result, Total: total,
	}, nil
}
