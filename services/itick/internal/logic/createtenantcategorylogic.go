package logic

import (
	"context"

	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateTenantCategoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateTenantCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTenantCategoryLogic {
	return &CreateTenantCategoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 租户产品类型
func (l *CreateTenantCategoryLogic) CreateTenantCategory(in *itick.CreateTenantCategoryReq) (*itick.AdminCommonResp, error) {
	// todo: add your logic here and delete this line

	return &itick.AdminCommonResp{}, nil
}
