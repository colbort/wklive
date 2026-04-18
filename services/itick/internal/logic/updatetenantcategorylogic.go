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

type UpdateTenantCategoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateTenantCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTenantCategoryLogic {
	return &UpdateTenantCategoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新租户产品类型仅允许更新状态、排序和备注，关联的产品类型不允许修改
func (l *UpdateTenantCategoryLogic) UpdateTenantCategory(in *itick.UpdateTenantCategoryReq) (*itick.AdminCommonResp, error) {
	item, err := l.svcCtx.ItickTenantCategoryModel.FindOne(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}
	if item == nil || item.TenantId != in.TenantId {
		return &itick.AdminCommonResp{
			Base: helper.GetErrResp(1, i18n.Translate(i18n.BusinessDataNotFound, l.ctx)),
		}, nil
	}

	item.Enabled = in.Enabled
	item.AppVisible = in.AppVisible
	item.Sort = in.Sort
	item.Remark = in.Remark
	item.UpdateTimes = cutils.NowMillis()

	if err := l.svcCtx.ItickTenantCategoryModel.Update(l.ctx, item); err != nil {
		return nil, err
	}

	return &itick.AdminCommonResp{Base: helper.OkResp()}, nil
}
