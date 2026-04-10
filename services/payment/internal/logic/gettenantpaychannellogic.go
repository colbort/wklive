package logic

import (
	"context"
	"errors"
	"fmt"

	"wklive/common/helper"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"
	"wklive/services/payment/models"

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
	var (
		errLogic = "GetTenantPayChannel"
	)

	channel, err := l.svcCtx.TenantPayChannelModel.FindOne(l.ctx, in.Id)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		l.Logger.Errorf("%s error: %s", errLogic, err.Error())
		return nil, err
	}

	if channel == nil {
		return &payment.GetTenantPayChannelResp{
			Base: helper.GetErrResp(404, "通道不存在"),
		}, nil
	}

	return &payment.GetTenantPayChannelResp{
		Base: helper.OkResp(),
		Data: &payment.TenantPayChannel{
			Id:              channel.Id,
			TenantId:        channel.TenantId,
			PlatformId:      channel.PlatformId,
			ProductId:       channel.ProductId,
			AccountId:       channel.AccountId,
			ChannelCode:     channel.ChannelCode,
			ChannelName:     channel.ChannelName,
			DisplayName:     channel.DisplayName.String,
			Icon:            channel.Icon.String,
			Currency:        channel.Currency,
			Sort:            channel.Sort,
			Visible:         channel.Visible,
			Status:          payment.CommonStatus(channel.Status),
			SingleMinAmount: channel.SingleMinAmount,
			SingleMaxAmount: channel.SingleMaxAmount,
			DailyMaxAmount:  channel.DailyMaxAmount,
			DailyMaxCount:   channel.DailyMaxCount,
			FeeType:         payment.FeeType(channel.FeeType),
			FeeRate:         fmt.Sprintf("%f", channel.FeeRate),
			FeeFixedAmount:  channel.FeeFixedAmount,
			ExtConfig:       channel.ExtConfig.String,
			Remark:          channel.Remark.String,
			CreateTimes:     channel.CreateTimes,
			UpdateTimes:     channel.UpdateTimes,
		},
	}, nil
}
