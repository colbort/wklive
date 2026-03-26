package logic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListTenantPayChannelRulesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListTenantPayChannelRulesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTenantPayChannelRulesLogic {
	return &ListTenantPayChannelRulesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 通道规则列表
func (l *ListTenantPayChannelRulesLogic) ListTenantPayChannelRules(in *payment.ListTenantPayChannelRulesReq) (*payment.ListTenantPayChannelRulesResp, error) {
	// todo: add your logic here and delete this line

	return &payment.ListTenantPayChannelRulesResp{}, nil
}
