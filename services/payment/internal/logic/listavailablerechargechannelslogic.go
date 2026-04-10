package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"
	"wklive/services/payment/models"

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
	channels, _, err := l.svcCtx.TenantPayChannelModel.FindPage(l.ctx, in.TenantId, 0, 0, 0, "", 1, 0, 0)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	data := make([]*payment.VisiblePayChannel, 0)
	for _, ch := range channels {
		visibleChannel := &payment.VisiblePayChannel{
			ChannelId:   ch.Id,
			ChannelName: ch.ChannelName,
			// DisplayName: ch.DisplayName,
			// Icon:        ch.Icon,
			// Currency:    ch.Currency,
			// Sort:        ch.Sort,
			// FeeType:     payment.FeeType(ch.FeeType),
			// FeeRate:     ch.FeeRate,
			// FeeFixed:    ch.FeeFixedAmount,
			// MinAmount:   ch.SingleMinAmount,
			// MaxAmount:   ch.SingleMaxAmount,
		}
		data = append(data, visibleChannel)
	}

	return &payment.ListAvailableRechargeChannelsResp{
		Base: helper.OkResp(),
		List: data,
	}, nil
}
