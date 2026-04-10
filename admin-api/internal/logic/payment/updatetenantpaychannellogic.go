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

type UpdateTenantPayChannelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateTenantPayChannelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTenantPayChannelLogic {
	return &UpdateTenantPayChannelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateTenantPayChannelLogic) UpdateTenantPayChannel(req *types.UpdateTenantPayChannelReq) (resp *types.RespBase, err error) {
	result, err := l.svcCtx.PaymentCli.UpdateTenantPayChannel(l.ctx, &payment.UpdateTenantPayChannelReq{
		Id:              req.Id,
		TenantId:        req.TenantId,
		ChannelName:     req.ChannelName,
		DisplayName:     req.DisplayName,
		Icon:            req.Icon,
		Currency:        req.Currency,
		Sort:            req.Sort,
		Visible:         req.Visible,
		Status:          payment.CommonStatus(req.Status),
		SingleMinAmount: req.SingleMinAmount,
		SingleMaxAmount: req.SingleMaxAmount,
		DailyMaxAmount:  req.DailyMaxAmount,
		DailyMaxCount:   req.DailyMaxCount,
		FeeType:         payment.FeeType(req.FeeType),
		FeeRate:         req.FeeRate,
		FeeFixedAmount:  req.FeeFixedAmount,
		ExtConfig:       req.ExtConfig,
		Remark:          req.Remark,
	})
	if err != nil {
		return nil, err
	}

	return &types.RespBase{
		Code: result.Base.Code,
		Msg:  result.Base.Msg,
	}, nil
}
