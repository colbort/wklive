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

type ListInsuranceInventoryExitsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListInsuranceInventoryExitsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListInsuranceInventoryExitsLogic {
	return &ListInsuranceInventoryExitsLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *ListInsuranceInventoryExitsLogic) ListInsuranceInventoryExits(
	in *option.ListInsuranceInventoryExitsReq,
) (*option.ListInsuranceInventoryExitsResp, error) {
	_, allowed, forbidden, err := utils.ResolveAdminTenantWriteScopeFromMd(l.ctx, in.TenantId)
	if err != nil {
		return nil, err
	}
	if in.TenantId <= 0 || forbidden || !allowed {
		return &option.ListInsuranceInventoryExitsResp{
			Base: helper.ErrResp(i18n.PermissionDenied, i18n.Translate(i18n.PermissionDenied, l.ctx)),
		}, nil
	}
	cursor, limit := pageutil.Input(in.Page)
	items, total, err := l.svcCtx.OptionInsuranceInventoryExitModel.FindPage(
		l.ctx, models.OptionInsuranceInventoryExitPageFilter{
			TenantId: in.TenantId, PositionId: in.PositionId,
			ContractId: in.ContractId, Status: int64(in.Status),
		}, cursor, limit,
	)
	if err != nil {
		return nil, err
	}
	data := make([]*option.OptionInsuranceInventoryExit, 0, len(items))
	lastID := int64(0)
	for _, item := range items {
		exit := helpers.ToInsuranceInventoryExitProto(item)
		if item.OrderId > 0 {
			order, findErr := l.svcCtx.OptionOrderModel.FindOne(l.ctx, item.OrderId)
			if findErr != nil {
				return nil, findErr
			}
			if order.TenantId != item.TenantId {
				return nil, models.ErrNotFound
			}
			helpers.ApplyInsuranceInventoryExitOrder(exit, order)
		}
		data = append(data, exit)
		lastID = item.Id
	}
	return &option.ListInsuranceInventoryExitsResp{
		Base: pageutil.Base(cursor, limit, len(items), total, lastID), Data: data, Total: total,
	}, nil
}
