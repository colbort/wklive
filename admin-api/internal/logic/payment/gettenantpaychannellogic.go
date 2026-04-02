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

type GetTenantPayChannelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTenantPayChannelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTenantPayChannelLogic {
	return &GetTenantPayChannelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTenantPayChannelLogic) GetTenantPayChannel(req *types.GetTenantPayChannelReq) (resp *types.GetTenantPayChannelResp, err error) {
	result, err := l.svcCtx.PaymentCli.GetTenantPayChannel(l.ctx, &payment.GetTenantPayChannelReq{
		Id:       req.Id,
		TenantId: req.TenantId,
	})
	if err != nil {
		return nil, err
	}

	resp = &types.GetTenantPayChannelResp{
		RespBase: types.RespBase{
			Code: result.Base.Code,
			Msg:  result.Base.Msg,
		},
		Data: types.TenantPayChannel{
			Id:              result.Data.Id,
			TenantId:        result.Data.TenantId,
			PlatformId:      result.Data.PlatformId,
			ProductId:       result.Data.ProductId,
			AccountId:       result.Data.AccountId,
			ChannelCode:     result.Data.ChannelCode,
			ChannelName:     result.Data.ChannelName,
			DisplayName:     result.Data.DisplayName,
			Icon:            result.Data.Icon,
			Currency:        result.Data.Currency,
			Sort:            int64(result.Data.Sort),
			Visible:         result.Data.Visible,
			Status:          int64(result.Data.Status),
			SingleMinAmount: result.Data.SingleMinAmount,
			SingleMaxAmount: result.Data.SingleMaxAmount,
			DailyMaxAmount:  result.Data.DailyMaxAmount,
			DailyMaxCount:   int64(result.Data.DailyMaxCount),
			FeeType:         int64(result.Data.FeeType),
			FeeRate:         result.Data.FeeRate,
			FeeFixedAmount:  result.Data.FeeFixedAmount,
			ExtConfig:       result.Data.ExtConfig,
			Remark:          result.Data.Remark,
			CreateTime:      result.Data.CreateTime,
			UpdateTime:      result.Data.UpdateTime,
		},
	}
	return resp, nil
}
