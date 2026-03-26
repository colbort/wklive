package logic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CheckRechargeChannelsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCheckRechargeChannelsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckRechargeChannelsLogic {
	return &CheckRechargeChannelsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 检查某金额可用的通道（预检查）
func (l *CheckRechargeChannelsLogic) CheckRechargeChannels(in *payment.CheckRechargeChannelsReq) (*payment.CheckRechargeChannelsResp, error) {
	// todo: add your logic here and delete this line

	return &payment.CheckRechargeChannelsResp{}, nil
}
