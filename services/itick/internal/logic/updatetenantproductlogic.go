package logic

import (
	"context"

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
	// todo: add your logic here and delete this line

	return &itick.AdminCommonResp{}, nil
}
