package logic

import (
	"context"

	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListTenantCategoriesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListTenantCategoriesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTenantCategoriesLogic {
	return &ListTenantCategoriesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 租户产品类型列表
func (l *ListTenantCategoriesLogic) ListTenantCategories(in *itick.ListTenantCategoriesReq) (*itick.ListTenantCategoriesResp, error) {
	// todo: add your logic here and delete this line

	return &itick.ListTenantCategoriesResp{}, nil
}
