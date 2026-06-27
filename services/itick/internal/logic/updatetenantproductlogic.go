package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/common/i18n"
	cutils "wklive/common/utils"
	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateTenantProductLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateTenantProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTenantProductLogic {
	return &UpdateTenantProductLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新租户产品仅允许更新状态、排序和备注，关联的产品不允许修改
func (l *UpdateTenantProductLogic) UpdateTenantProduct(in *itick.UpdateTenantProductReq) (*itick.AdminCommonResp, error) {
	item, err := l.svcCtx.ItickTenantProductModel.FindOne(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return &itick.AdminCommonResp{
			Base: helper.ErrResp(i18n.BusinessDataNotFound, i18n.Translate(i18n.BusinessDataNotFound, l.ctx)),
		}, nil
	}
	_, allowed, forbidden, err := cutils.ResolveAdminTenantWriteScopeFromMd(l.ctx, item.TenantId)
	if err != nil {
		return nil, i18n.StatusError(l.ctx, i18n.UserNotFound)
	}
	if forbidden {
		return &itick.AdminCommonResp{
			Base: helper.ErrResp(i18n.PermissionDenied, i18n.Translate(i18n.PermissionDenied, l.ctx)),
		}, nil
	}
	if !allowed {
		return &itick.AdminCommonResp{
			Base: helper.ErrResp(i18n.BusinessDataNotFound, i18n.Translate(i18n.BusinessDataNotFound, l.ctx)),
		}, nil
	}

	if in.Enabled != 0 {
		item.Enabled = int64(in.Enabled)
	}
	if in.AppVisible != 0 {
		item.AppVisible = int64(in.AppVisible)
	}
	if in.Sort != 0 {
		item.Sort = in.Sort
	}
	if in.Remark != "" {
		item.Remark = in.Remark
	}
	item.UpdateTimes = cutils.NowMillis()

	if err := l.svcCtx.ItickTenantProductModel.Update(l.ctx, item); err != nil {
		return nil, err
	}

	return &itick.AdminCommonResp{Base: helper.OkResp()}, nil
}
