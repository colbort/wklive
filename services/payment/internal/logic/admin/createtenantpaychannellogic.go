package adminlogic

import (
	"context"
	"database/sql"
	"wklive/services/payment/internal/logic/helpers"

	"wklive/common/conv"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"
	"wklive/services/payment/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateTenantPayChannelLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateTenantPayChannelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTenantPayChannelLogic {
	return &CreateTenantPayChannelLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建租户支付通道
func (l *CreateTenantPayChannelLogic) CreateTenantPayChannel(in *payment.CreateTenantPayChannelReq) (*payment.CommonResp, error) {
	var (
		errLogic = "CreateTenantPayChannel"
	)
	tenantID, resp, err := resolveAdminTenantCreateScope(l.ctx, in.TenantId)
	if err != nil {
		return nil, err
	}
	if resp != nil {
		return resp, nil
	}
	if in.PlatformId <= 0 || in.ProductId <= 0 || in.AccountId <= 0 ||
		!requiredStrings(in.ChannelCode, in.ChannelName, in.Currency) || in.FeeType == 0 {
		return paymentErrorResp(l.ctx, i18n.PaymentRequiredParamsMissing), nil
	}
	if _, ok := payment.FeeType_name[int32(in.FeeType)]; !ok {
		return paymentErrorResp(l.ctx, i18n.ParamError), nil
	}
	singleMinAmount, minErr := parseNonNegativeAmount(in.SingleMinAmount)
	singleMaxAmount, maxErr := parseNonNegativeAmount(in.SingleMaxAmount)
	dailyMaxAmount, dailyErr := parseNonNegativeAmount(in.DailyMaxAmount)
	feeFixedAmount, fixedErr := parseNonNegativeAmount(in.FeeFixedAmount)
	if minErr != nil || maxErr != nil || dailyErr != nil || fixedErr != nil ||
		!validDecimalRange(singleMinAmount, singleMaxAmount) || in.DailyMaxCount < 0 {
		return paymentErrorResp(l.ctx, i18n.InvalidPaymentAmountRange), nil
	}
	if relationResp, err := validateProductPlatform(l.ctx, l.svcCtx, in.ProductId, in.PlatformId); err != nil {
		return nil, err
	} else if relationResp != nil {
		return relationResp, nil
	}
	if relationResp, err := validateAccountPlatform(l.ctx, l.svcCtx, in.AccountId, tenantID, in.PlatformId); err != nil {
		return nil, err
	} else if relationResp != nil {
		return relationResp, nil
	}

	feeRate, err := conv.ParseDecimalField(in.FeeRate)
	if err != nil {
		return paymentErrorResp(l.ctx, i18n.InvalidPaymentDecimal), nil
	}
	if feeRate.IsNegative() {
		return paymentErrorResp(l.ctx, i18n.InvalidPaymentAmountRange), nil
	}
	extConfig, valid := helpers.NullableJSON(in.ExtConfig)
	if !valid {
		return &payment.CommonResp{
			Base: helper.ErrResp(i18n.InvalidPaymentJSON, i18n.Translate(i18n.InvalidPaymentJSON, l.ctx)),
		}, nil
	}

	now := utils.NowMillis()
	channel := &models.TTenantPayChannel{
		TenantId:        tenantID,
		PlatformId:      in.PlatformId,
		ProductId:       in.ProductId,
		AccountId:       in.AccountId,
		ChannelCode:     in.ChannelCode,
		ChannelName:     in.ChannelName,
		DisplayName:     sql.NullString{String: in.DisplayName, Valid: true},
		Icon:            sql.NullString{String: in.Icon, Valid: true},
		Currency:        in.Currency,
		Sort:            in.Sort,
		Visible:         helpers.SwitchToModel(in.Visible, int64(common.Switch_SWITCH_OFF)),
		Enabled:         helpers.EnableToModel(in.Enabled, int64(common.Enable_ENABLE_ENABLED)),
		SingleMinAmount: singleMinAmount,
		SingleMaxAmount: singleMaxAmount,
		DailyMaxAmount:  dailyMaxAmount,
		DailyMaxCount:   in.DailyMaxCount,
		FeeType:         int64(in.FeeType),
		FeeRate:         feeRate,
		FeeFixedAmount:  feeFixedAmount,
		ExtConfig:       extConfig,
		Remark:          sql.NullString{String: in.Remark, Valid: true},
		CreateTimes:     now,
		UpdateTimes:     now,
	}

	_, err = l.svcCtx.TenantPayChannelModel.Insert(l.ctx, channel)
	if err != nil {
		if helpers.IsDuplicateEntry(err) {
			return paymentErrorResp(l.ctx, i18n.TenantPayChannelCodeAlreadyExists), nil
		}
		l.Logger.Errorf("%s error: %s", errLogic, err.Error())
		return nil, err
	}

	l.Logger.Infof("Create tenant pay channel success: %s", in.ChannelCode)

	return &payment.CommonResp{
		Base: helper.OkResp(),
	}, nil
}
