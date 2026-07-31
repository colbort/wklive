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

type ListPhysicalDeliveryUnitsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListPhysicalDeliveryUnitsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPhysicalDeliveryUnitsLogic {
	return &ListPhysicalDeliveryUnitsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分页查询实物交割配对单元和失败状态
func (l *ListPhysicalDeliveryUnitsLogic) ListPhysicalDeliveryUnits(in *option.ListPhysicalDeliveryUnitsReq) (*option.ListPhysicalDeliveryUnitsResp, error) {
	_, allowed, forbidden, err := utils.ResolveAdminTenantWriteScopeFromMd(l.ctx, in.TenantId)
	if err != nil {
		return nil, err
	}
	if forbidden || !allowed {
		return &option.ListPhysicalDeliveryUnitsResp{
			Base: helper.ErrResp(i18n.PermissionDenied, i18n.Translate(i18n.PermissionDenied, l.ctx)),
		}, nil
	}
	cursor, limit := pageutil.Input(in.Page)
	items, total, err := l.svcCtx.OptionPhysicalDeliveryUnitModel.FindPage(
		l.ctx, models.OptionPhysicalDeliveryUnitPageFilter{
			TenantId: in.TenantId, ContractId: in.ContractId, BatchId: in.BatchId,
			UserId: in.UserId, Status: int64(in.Status),
		}, cursor, limit,
	)
	if err != nil {
		return nil, err
	}
	data := make([]*option.OptionPhysicalDeliveryUnit, 0, len(items))
	lastID := int64(0)
	for _, item := range items {
		lastID = item.Id
		instructions, findErr := l.svcCtx.OptionAssetInstructionModel.FindByDeliveryUnit(
			l.ctx, item.TenantId, item.Id,
		)
		if findErr != nil {
			return nil, findErr
		}
		data = append(data, helpers.ToPhysicalDeliveryUnitProto(item, instructions))
	}
	return &option.ListPhysicalDeliveryUnitsResp{
		Base: pageutil.Base(cursor, limit, len(items), total, lastID),
		Data: data, Total: total,
	}, nil
}
