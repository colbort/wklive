package logic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTenantPayChannelLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTenantPayChannelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTenantPayChannelLogic {
	return &GetTenantPayChannelLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取租户支付通道详情
func (l *GetTenantPayChannelLogic) GetTenantPayChannel(in *payment.GetTenantPayChannelReq) (*payment.GetTenantPayChannelResp, error) {
	// todo: add your logic here and delete this line

	return &payment.GetTenantPayChannelResp{}, nil
}
