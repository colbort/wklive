package logic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTenantPayPlatformLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTenantPayPlatformLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTenantPayPlatformLogic {
	return &GetTenantPayPlatformLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取租户开通平台详情
func (l *GetTenantPayPlatformLogic) GetTenantPayPlatform(in *payment.GetTenantPayPlatformReq) (*payment.GetTenantPayPlatformResp, error) {
	// todo: add your logic here and delete this line

	return &payment.GetTenantPayPlatformResp{}, nil
}
