package adminlogic

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
func (l *UpdateTenantPayChannelLogic) UpdateTenantPayChannel(in *payment.UpdateTenantPayChannelReq) (*payment.CommonResp, error) {
	var (
		errLogic = "UpdateTenantPayChannel"
	)
	if in.Id <= 0 {
		return paymentErrorResp(l.ctx, i18n.PaymentRequiredParamsMissing), nil
	}

	// 查询通道是否存在
	channel, err := l.svcCtx.TenantPayChannelModel.FindOne(l.ctx, in.Id)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		l.Logger.Errorf("%s error: %s", errLogic, err.Error())
		return nil, err
	}

	if errors.Is(err, models.ErrNotFound) || channel == nil {
		return &payment.CommonResp{
			Base: helper.ErrResp(i18n.PaymentChannelNotFound, i18n.Translate(i18n.PaymentChannelNotFound, l.ctx)),
		}, nil
	}

	allowTenantUpdate, resp, err := applyAdminTenantUpdateScope(l.ctx, channel.TenantId, i18n.PaymentChannelNotFound)
	if err != nil {
		return nil, err
	}
	if resp != nil {
		return resp, nil
	}
	if allowTenantUpdate && in.TenantId > 0 {
		channel.TenantId = in.TenantId
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
		channel.Visible = switchToModel(in.Visible, channel.Visible)
	}
	if in.Enabled != 0 {
		channel.Enabled = int64(in.Enabled)
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
		if _, ok := payment.FeeType_name[int32(in.FeeType)]; !ok {
			return paymentErrorResp(l.ctx, i18n.ParamError), nil
		}
		channel.FeeType = int64(in.FeeType)
	}
	if in.FeeRate != "" {
		feeRate, err := conv.ParseDecimalField(in.FeeRate)
		if err != nil {
			return paymentErrorResp(l.ctx, i18n.InvalidPaymentDecimal), nil
		}
		if feeRate.IsNegative() {
			return paymentErrorResp(l.ctx, i18n.InvalidPaymentAmountRange), nil
		}
		channel.FeeRate = feeRate
	}
	if in.FeeFixedAmount != 0 {
		channel.FeeFixedAmount = in.FeeFixedAmount
	}
	if in.ExtConfig != "" {
		extConfig, valid := nullableJSON(in.ExtConfig)
		if !valid {
			return &payment.CommonResp{
				Base: helper.ErrResp(i18n.InvalidPaymentJSON, i18n.Translate(i18n.InvalidPaymentJSON, l.ctx)),
			}, nil
		}
		channel.ExtConfig = extConfig
	}
	if in.Remark != "" {
		channel.Remark = sql.NullString{String: in.Remark, Valid: true}
	}
	if !requiredStrings(channel.ChannelName, channel.Currency) {
		return paymentErrorResp(l.ctx, i18n.PaymentRequiredParamsMissing), nil
	}
	if !validNonNegativeRange(channel.SingleMinAmount, channel.SingleMaxAmount) ||
		channel.DailyMaxAmount < 0 || channel.DailyMaxCount < 0 || channel.FeeFixedAmount < 0 {
		return paymentErrorResp(l.ctx, i18n.InvalidPaymentAmountRange), nil
	}
	channel.UpdateTimes = now

	err = l.svcCtx.TenantPayChannelModel.Update(l.ctx, channel)
	if err != nil {
		if isDuplicateEntry(err) {
			return paymentErrorResp(l.ctx, i18n.TenantPayChannelCodeAlreadyExists), nil
		}
		l.Logger.Errorf("%s error: %s", errLogic, err.Error())
		return nil, err
	}

	l.Logger.Infof("Update tenant pay channel success: %d", in.Id)

	return &payment.CommonResp{
		Base: helper.OkResp(),
	}, nil
}
