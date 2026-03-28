package logic

import (
	"context"

	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTenantProductCategoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTenantProductCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTenantProductCategoryLogic {
	return &GetTenantProductCategoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取租户产品类型详情
func (l *GetTenantProductCategoryLogic) GetTenantProductCategory(in *itick.GetTenantProductCategoryReq) (*itick.GetTenantProductCategoryResp, error) {
	// todo: add your logic here and delete this line

	return &itick.GetTenantProductCategoryResp{}, nil
}
