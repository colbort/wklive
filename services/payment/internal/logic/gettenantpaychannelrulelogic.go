package logic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTenantPayChannelRuleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTenantPayChannelRuleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTenantPayChannelRuleLogic {
	return &GetTenantPayChannelRuleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取通道规则详情
func (l *GetTenantPayChannelRuleLogic) GetTenantPayChannelRule(in *payment.GetTenantPayChannelRuleReq) (*payment.GetTenantPayChannelRuleResp, error) {
	// todo: add your logic here and delete this line

	return &payment.GetTenantPayChannelRuleResp{}, nil
}
