package logic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListTenantPayPlatformsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListTenantPayPlatformsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTenantPayPlatformsLogic {
	return &ListTenantPayPlatformsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 租户开通平台列表
func (l *ListTenantPayPlatformsLogic) ListTenantPayPlatforms(in *payment.ListTenantPayPlatformsReq) (*payment.ListTenantPayPlatformsResp, error) {
	// todo: add your logic here and delete this line

	return &payment.ListTenantPayPlatformsResp{}, nil
}
