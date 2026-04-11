package logic

import (
	"context"
	"database/sql"
	"errors"
	"wklive/common/conv"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"
	"wklive/services/payment/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateTenantPayChannelLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateTenantPayChannelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTenantPayChannelLogic {
	return &UpdateTenantPayChannelLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新租户支付通道
func (l *UpdateTenantPayChannelLogic) UpdateTenantPayChannel(in *payment.UpdateTenantPayChannelReq) (*payment.AdminCommonResp, error) {
	var (
		errLogic = "UpdateTenantPayChannel"
	)

	// 查询通道是否存在
	channel, err := l.svcCtx.TenantPayChannelModel.FindOne(l.ctx, in.Id)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		l.Logger.Errorf("%s error: %s", errLogic, err.Error())
		return nil, err
	}

	if channel == nil {
		return &payment.AdminCommonResp{
			Base: helper.GetErrResp(404, i18n.Translate(i18n.PaymentChannelNotFound, l.ctx)),
		}, nil
	}

	now := utils.NowMillis()
	if in.ChannelName != "" {
		channel.ChannelName = in.ChannelName
	}
	if in.DisplayName != "" {
		channel.DisplayName = sql.NullString{String: in.DisplayName, Valid: true}
	}
	if in.Icon != "" {
		channel.Icon = sql.NullString{String: in.Icon, Valid: true}
	}
	if in.Currency != "" {
		channel.Currency = in.Currency
	}
	if in.Sort != 0 {
		channel.Sort = in.Sort
	}
	if in.Visible != 0 {
		channel.Visible = in.Visible
	}
	if in.Status != 0 {
		channel.Status = int64(in.Status)
	}
	if in.SingleMinAmount != 0 {
		channel.SingleMinAmount = in.SingleMinAmount
	}
	if in.SingleMaxAmount != 0 {
		channel.SingleMaxAmount = in.SingleMaxAmount
	}
	if in.DailyMaxAmount != 0 {
		channel.DailyMaxAmount = in.DailyMaxAmount
	}
	if in.DailyMaxCount != 0 {
		channel.DailyMaxCount = in.DailyMaxCount
	}
	if in.FeeType != 0 {
		channel.FeeType = int64(in.FeeType)
	}
	if in.FeeRate != "" {
		feeRate, err := conv.ParseFloatField(in.FeeRate)
		if err != nil {
			return nil, err
		}
		channel.FeeRate = feeRate
	}
	if in.FeeFixedAmount != 0 {
		channel.FeeFixedAmount = in.FeeFixedAmount
	}
	if in.ExtConfig != "" {
		channel.ExtConfig = sql.NullString{String: in.ExtConfig, Valid: true}
	}
	if in.Remark != "" {
		channel.Remark = sql.NullString{String: in.Remark, Valid: true}
	}
	channel.UpdateTimes = now

	err = l.svcCtx.TenantPayChannelModel.Update(l.ctx, channel)
	if err != nil {
		l.Logger.Errorf("%s error: %s", errLogic, err.Error())
		return nil, err
	}

	l.Logger.Infof("Update tenant pay channel success: %d", in.Id)

	return &payment.AdminCommonResp{
		Base: helper.OkResp(),
	}, nil
}
