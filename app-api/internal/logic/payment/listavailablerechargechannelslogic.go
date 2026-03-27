// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"
	"wklive/common/utils"
	"wklive/proto/payment"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAvailableRechargeChannelsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListAvailableRechargeChannelsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAvailableRechargeChannelsLogic {
	return &ListAvailableRechargeChannelsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListAvailableRechargeChannelsLogic) ListAvailableRechargeChannels(req *types.ListAvailableRechargeChannelsReq) (resp *types.ListAvailableRechargeChannelsResp, err error) {
	userId, err := utils.GetUidFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}
	result, err := l.svcCtx.PaymentCli.ListAvailableRechargeChannels(l.ctx, &payment.ListAvailableRechargeChannelsReq{
		UserId:         userId,
		TenantId:       req.TenantId,
		RechargeAmount: req.RechargeAmount,
		Currency:       req.Currency,
		ClientType:     payment.ClientType(req.ClientType),
	})
	if err != nil {
		return nil, err
	}
	data := make([]types.AvailableRechargeChannel, 0)
	for _, channel := range result.List {
		data = append(data, types.AvailableRechargeChannel{
			ChannelId:       channel.ChannelId,
			ChannelCode:     channel.ChannelCode,
			ChannelName:     channel.ChannelName,
			DisplayName:     channel.DisplayName,
			Icon:            channel.Icon,
			Currency:        channel.Currency,
			SingleMinAmount: channel.SingleMinAmount,
			SingleMaxAmount: channel.SingleMaxAmount,
			FeeType:         int64(channel.FeeType),
			FeeRate:         channel.FeeRate,
			FeeFixedAmount:  channel.FeeFixedAmount,
			PlatformId:      channel.PlatformId,
			ProductId:       channel.ProductId,
			AccountId:       channel.AccountId,
		})
	}
	return &types.ListAvailableRechargeChannelsResp{
		RespBase: types.RespBase{
			Code: result.Base.Code,
			Msg:  result.Base.Msg,
		},
		Data: data,
	}, nil
}
