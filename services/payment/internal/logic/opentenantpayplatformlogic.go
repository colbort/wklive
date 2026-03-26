package logic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type OpenTenantPayPlatformLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewOpenTenantPayPlatformLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OpenTenantPayPlatformLogic {
	return &OpenTenantPayPlatformLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 租户开通平台
func (l *OpenTenantPayPlatformLogic) OpenTenantPayPlatform(in *payment.OpenTenantPayPlatformReq) (*payment.AdminCommonResp, error) {
	// todo: add your logic here and delete this line

	return &payment.AdminCommonResp{}, nil
}
