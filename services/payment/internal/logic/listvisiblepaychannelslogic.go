package logic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListVisiblePayChannelsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListVisiblePayChannelsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListVisiblePayChannelsLogic {
	return &ListVisiblePayChannelsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取当前金额下可见通道
func (l *ListVisiblePayChannelsLogic) ListVisiblePayChannels(in *payment.ListVisiblePayChannelsReq) (*payment.ListVisiblePayChannelsResp, error) {
	// todo: add your logic here and delete this line

	return &payment.ListVisiblePayChannelsResp{}, nil
}
