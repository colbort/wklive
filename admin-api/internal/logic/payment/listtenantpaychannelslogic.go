// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/proto/payment"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListTenantPayChannelsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListTenantPayChannelsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTenantPayChannelsLogic {
	return &ListTenantPayChannelsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListTenantPayChannelsLogic) ListTenantPayChannels(req *types.ListTenantPayChannelsReq) (resp *types.ListTenantPayChannelsResp, err error) {
	result, err := l.svcCtx.PaymentCli.ListTenantPayChannels(l.ctx, &payment.ListTenantPayChannelsReq{
		Page: &payment.PageReq{
			Cursor: req.Cursor,
			Limit:  req.Limit,
		},
		TenantId:   req.TenantId,
		PlatformId: req.PlatformId,
		ProductId:  req.ProductId,
		AccountId:  req.AccountId,
		Keyword:    req.Keyword,
		Status:     payment.CommonStatus(req.Status),
		Visible:    req.Visible,
	})
	if err != nil {
		return nil, err
	}

	data := make([]types.TenantPayChannel, len(result.Data))
	for i, item := range result.Data {
		data[i] = types.TenantPayChannel{
			Id:              item.Id,
			TenantId:        item.TenantId,
			PlatformId:      item.PlatformId,
			ProductId:       item.ProductId,
			AccountId:       item.AccountId,
			ChannelCode:     item.ChannelCode,
			ChannelName:     item.ChannelName,
			DisplayName:     item.DisplayName,
			Icon:            item.Icon,
			Currency:        item.Currency,
			Sort:            int64(item.Sort),
			Visible:         item.Visible,
			Status:          int64(item.Status),
			SingleMinAmount: item.SingleMinAmount,
			SingleMaxAmount: item.SingleMaxAmount,
			DailyMaxAmount:  item.DailyMaxAmount,
			DailyMaxCount:   int64(item.DailyMaxCount),
			FeeType:         int64(item.FeeType),
			FeeRate:         item.FeeRate,
			FeeFixedAmount:  item.FeeFixedAmount,
			ExtConfig:       item.ExtConfig,
			Remark:          item.Remark,
			CreateTime:      item.CreateTime,
			UpdateTime:      item.UpdateTime,
		}
	}

	resp = &types.ListTenantPayChannelsResp{
		RespBase: types.RespBase{
			Code:       result.Base.Code,
			Msg:        result.Base.Msg,
			Total:      result.Base.Total,
			HasNext:    result.Base.HasNext,
			HasPrev:    result.Base.HasPrev,
			NextCursor: result.Base.NextCursor,
			PrevCursor: result.Base.PrevCursor,
		},
		Data: data,
	}
	return resp, nil
}
