package logic

import (
	"context"

	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateTenantProductCategoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateTenantProductCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTenantProductCategoryLogic {
	return &CreateTenantProductCategoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 租户产品类型
func (l *CreateTenantProductCategoryLogic) CreateTenantProductCategory(in *itick.CreateTenantProductCategoryReq) (*itick.AdminCommonResp, error) {
	// todo: add your logic here and delete this line

	return &itick.AdminCommonResp{}, nil
}
