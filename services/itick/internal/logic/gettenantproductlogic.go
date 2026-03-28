package logic

import (
	"context"

	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTenantProductLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTenantProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTenantProductLogic {
	return &GetTenantProductLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取租户产品详情
func (l *GetTenantProductLogic) GetTenantProduct(in *itick.GetTenantProductReq) (*itick.GetTenantProductResp, error) {
	// todo: add your logic here and delete this line

	return &itick.GetTenantProductResp{}, nil
}
