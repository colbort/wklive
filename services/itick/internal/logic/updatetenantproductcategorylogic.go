package logic

import (
	"context"

	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateTenantProductCategoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateTenantProductCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTenantProductCategoryLogic {
	return &UpdateTenantProductCategoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新租户产品类型仅允许更新状态、排序和备注，关联的产品类型不允许修改
func (l *UpdateTenantProductCategoryLogic) UpdateTenantProductCategory(in *itick.UpdateTenantProductCategoryReq) (*itick.AdminCommonResp, error) {
	// todo: add your logic here and delete this line

	return &itick.AdminCommonResp{}, nil
}
