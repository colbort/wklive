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

type ListLiquidationsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListLiquidationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListLiquidationsLogic {
	return &ListLiquidationsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分页查询强平记录
func (l *ListLiquidationsLogic) ListLiquidations(in *option.ListLiquidationsReq) (*option.ListLiquidationsResp, error) {
	tenantId, allowed, forbidden, err := utils.ResolveAdminTenantReadScopeFromMd(l.ctx, in.TenantId)
	if err != nil {
		return nil, err
	}
	if forbidden || !allowed {
		return &option.ListLiquidationsResp{
			Base: helper.ErrResp(i18n.PermissionDenied, i18n.Translate(i18n.PermissionDenied, l.ctx)),
		}, nil
	}
	cursor, limit := pageutil.Input(in.Page)
	items, total, err := l.svcCtx.OptionLiquidationModel.FindPage(l.ctx, models.OptionLiquidationPageFilter{
		TenantId: tenantId, UserId: in.UserId, AccountId: in.AccountId,
		ContractId: in.ContractId, PositionId: in.PositionId, Status: int64(in.Status),
	}, cursor, limit)
	if err != nil {
		return nil, err
	}
	data := make([]*option.OptionLiquidation, 0, len(items))
	lastID := int64(0)
	for _, item := range items {
		lastID = item.Id
		data = append(data, helpers.ToLiquidationProto(item))
	}
	return &option.ListLiquidationsResp{
		Base: pageutil.Base(cursor, limit, len(items), total, lastID), Data: data,
	}, nil
}
