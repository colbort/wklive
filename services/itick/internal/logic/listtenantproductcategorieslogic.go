package logic

import (
	"context"

	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListTenantProductCategoriesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListTenantProductCategoriesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTenantProductCategoriesLogic {
	return &ListTenantProductCategoriesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 租户产品类型列表
func (l *ListTenantProductCategoriesLogic) ListTenantProductCategories(in *itick.ListTenantProductCategoriesReq) (*itick.ListTenantProductCategoriesResp, error) {
	// todo: add your logic here and delete this line

	return &itick.ListTenantProductCategoriesResp{}, nil
}
