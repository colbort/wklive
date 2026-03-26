package logic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAvailableRechargeChannelsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListAvailableRechargeChannelsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAvailableRechargeChannelsLogic {
	return &ListAvailableRechargeChannelsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取当前登录用户在指定充值金额下可用的充值通道
func (l *ListAvailableRechargeChannelsLogic) ListAvailableRechargeChannels(in *payment.ListAvailableRechargeChannelsReq) (*payment.ListAvailableRechargeChannelsResp, error) {
	// todo: add your logic here and delete this line

	return &payment.ListAvailableRechargeChannelsResp{}, nil
}
