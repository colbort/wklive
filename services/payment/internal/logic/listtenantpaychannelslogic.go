package logic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListTenantPayChannelsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListTenantPayChannelsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTenantPayChannelsLogic {
	return &ListTenantPayChannelsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 租户支付通道列表
func (l *ListTenantPayChannelsLogic) ListTenantPayChannels(in *payment.ListTenantPayChannelsReq) (*payment.ListTenantPayChannelsResp, error) {
	// todo: add your logic here and delete this line

	return &payment.ListTenantPayChannelsResp{}, nil
}
