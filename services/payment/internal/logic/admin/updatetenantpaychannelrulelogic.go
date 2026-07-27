package adminlogic

import (
	"context"
	"database/sql"
	"errors"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"
	"wklive/services/payment/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateTenantPayChannelRuleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateTenantPayChannelRuleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTenantPayChannelRuleLogic {
	return &UpdateTenantPayChannelRuleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新通道规则
func (l *UpdateTenantPayChannelRuleLogic) UpdateTenantPayChannelRule(in *payment.UpdateTenantPayChannelRuleReq) (*payment.CommonResp, error) {
	var (
		errLogic = "UpdateTenantPayChannelRule"
	)
	if in.Id <= 0 {
		return paymentErrorResp(l.ctx, i18n.PaymentRequiredParamsMissing), nil
	}

	// 查询规则是否存在
	rule, err := l.svcCtx.TenantPayChannelRuleModel.FindOne(l.ctx, in.Id)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		l.Logger.Errorf("%s error: %s", errLogic, err.Error())
		return nil, err
	}

	if errors.Is(err, models.ErrNotFound) || rule == nil {
		return &payment.CommonResp{
			Base: helper.ErrResp(i18n.ChannelRuleNotFound, i18n.Translate(i18n.ChannelRuleNotFound, l.ctx)),
		}, nil
	}

	allowTenantUpdate, resp, err := applyAdminTenantUpdateScope(l.ctx, rule.TenantId, i18n.ChannelRuleNotFound)
	if err != nil {
		return nil, err
	}
	if resp != nil {
		return resp, nil
	}
	if allowTenantUpdate && in.TenantId > 0 {
		rule.TenantId = in.TenantId
	}

	now := utils.NowMillis()
	if in.RuleName != "" {
		rule.RuleName = in.RuleName
	}
	if in.Priority != 0 {
		rule.Priority = in.Priority
	}
	if in.Enabled != 0 {
		rule.Enabled = int64(in.Enabled)
	}
	if in.SingleAmountMin != "" {
		value, err := parseNonNegativeAmount(in.SingleAmountMin)
		if err != nil {
			return paymentErrorResp(l.ctx, i18n.InvalidPaymentDecimal), nil
		}
		rule.SingleAmountMin = value
	}
	if in.SingleAmountMax != "" {
		value, err := parseNonNegativeAmount(in.SingleAmountMax)
		if err != nil {
			return paymentErrorResp(l.ctx, i18n.InvalidPaymentDecimal), nil
		}
		rule.SingleAmountMax = value
	}
	if in.UserTotalRechargeMin != "" {
		value, err := parseNonNegativeAmount(in.UserTotalRechargeMin)
		if err != nil {
			return paymentErrorResp(l.ctx, i18n.InvalidPaymentDecimal), nil
		}
		rule.UserTotalRechargeMin = value
	}
	if in.UserTotalRechargeMax != "" {
		value, err := parseNonNegativeAmount(in.UserTotalRechargeMax)
		if err != nil {
			return paymentErrorResp(l.ctx, i18n.InvalidPaymentDecimal), nil
		}
		rule.UserTotalRechargeMax = value
	}
	if in.MemberLevelMin != 0 {
		rule.MemberLevelMin = in.MemberLevelMin
	}
	if in.MemberLevelMax != 0 {
		rule.MemberLevelMax = in.MemberLevelMax
	}
	if in.KycLevelMin != 0 {
		rule.KycLevelMin = in.KycLevelMin
	}
	if in.KycLevelMax != 0 {
		rule.KycLevelMax = in.KycLevelMax
	}
	if common.YesNo(in.AllowNewUser) != common.YesNo_YES_NO_UNKNOWN {
		rule.AllowNewUser = int64(in.AllowNewUser)
	}
	if common.YesNo(in.AllowOldUser) != common.YesNo_YES_NO_UNKNOWN {
		rule.AllowOldUser = int64(in.AllowOldUser)
	}
	if in.AllowTags != "" {
		if !validJSONArray(in.AllowTags) {
			return paymentErrorResp(l.ctx, i18n.InvalidPaymentJSON), nil
		}
		rule.AllowTags = sql.NullString{String: in.AllowTags, Valid: true}
	}
	if in.DenyTags != "" {
		if !validJSONArray(in.DenyTags) {
			return paymentErrorResp(l.ctx, i18n.InvalidPaymentJSON), nil
		}
		rule.DenyTags = sql.NullString{String: in.DenyTags, Valid: true}
	}
	if in.Remark != "" {
		rule.Remark = sql.NullString{String: in.Remark, Valid: true}
	}
	if !requiredStrings(rule.RuleName) {
		return paymentErrorResp(l.ctx, i18n.PaymentRequiredParamsMissing), nil
	}
	if !validDecimalRange(rule.SingleAmountMin, rule.SingleAmountMax) ||
		!validDecimalRange(rule.UserTotalRechargeMin, rule.UserTotalRechargeMax) ||
		!validNonNegativeRange(rule.MemberLevelMin, rule.MemberLevelMax) ||
		!validNonNegativeRange(rule.KycLevelMin, rule.KycLevelMax) {
		return paymentErrorResp(l.ctx, i18n.InvalidPaymentAmountRange), nil
	}
	rule.UpdateTimes = now

	err = l.svcCtx.TenantPayChannelRuleModel.Update(l.ctx, rule)
	if err != nil {
		l.Logger.Errorf("%s error: %s", errLogic, err.Error())
		return nil, err
	}

	l.Logger.Infof("Update tenant pay channel rule success: %d", in.Id)

	return &payment.CommonResp{
		Base: helper.OkResp(),
	}, nil
}
