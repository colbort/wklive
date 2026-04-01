// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package itick

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTenantProductLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTenantProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTenantProductLogic {
	return &GetTenantProductLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTenantProductLogic) GetTenantProduct(req *types.GetTenantProductReq) (resp *types.GetTenantProductResp, err error) {
	// todo: add your logic here and delete this line

	return
}
