package logic

import (
	"context"

	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTenantCategoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTenantCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTenantCategoryLogic {
	return &GetTenantCategoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取租户产品类型详情
func (l *GetTenantCategoryLogic) GetTenantCategory(in *itick.GetTenantCategoryReq) (*itick.GetTenantCategoryResp, error) {
	// todo: add your logic here and delete this line

	return &itick.GetTenantCategoryResp{}, nil
}
