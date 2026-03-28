package logic

import (
	"context"

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
	// todo: add your logic here and delete this line

	return &itick.AdminCommonResp{}, nil
}
