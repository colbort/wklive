package logic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTenantPayAccountLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTenantPayAccountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTenantPayAccountLogic {
	return &GetTenantPayAccountLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取租户支付账号详情
func (l *GetTenantPayAccountLogic) GetTenantPayAccount(in *payment.GetTenantPayAccountReq) (*payment.GetTenantPayAccountResp, error) {
	// todo: add your logic here and delete this line

	return &payment.GetTenantPayAccountResp{}, nil
}
