package logic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateTenantPayPlatformLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateTenantPayPlatformLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTenantPayPlatformLogic {
	return &UpdateTenantPayPlatformLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新租户开通平台
func (l *UpdateTenantPayPlatformLogic) UpdateTenantPayPlatform(in *payment.UpdateTenantPayPlatformReq) (*payment.AdminCommonResp, error) {
	// todo: add your logic here and delete this line

	return &payment.AdminCommonResp{}, nil
}
