package logic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListTenantPayAccountsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListTenantPayAccountsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTenantPayAccountsLogic {
	return &ListTenantPayAccountsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 租户支付账号列表
func (l *ListTenantPayAccountsLogic) ListTenantPayAccounts(in *payment.ListTenantPayAccountsReq) (*payment.ListTenantPayAccountsResp, error) {
	// todo: add your logic here and delete this line

	return &payment.ListTenantPayAccountsResp{}, nil
}
