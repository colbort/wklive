package adminlogic

import (
	"context"
	"database/sql"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"
	"wklive/services/payment/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateTenantPayChannelRuleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateTenantPayChannelRuleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTenantPayChannelRuleLogic {
	return &CreateTenantPayChannelRuleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建通道规则
func (l *CreateTenantPayChannelRuleLogic) CreateTenantPayChannelRule(in *payment.CreateTenantPayChannelRuleReq) (*payment.CommonResp, error) {
	var (
		errLogic = "CreateTenantPayChannelRule"
	)
	tenantID, resp, err := resolveAdminTenantCreateScope(l.ctx, in.TenantId)
	if err != nil {
		return nil, err
	}
	if resp != nil {
		return resp, nil
	}
	if in.ChannelId <= 0 || !requiredStrings(in.RuleName) {
		return paymentErrorResp(l.ctx, i18n.PaymentRequiredParamsMissing), nil
	}
	singleAmountMin, singleMinErr := parseNonNegativeAmount(in.SingleAmountMin)
	singleAmountMax, singleMaxErr := parseNonNegativeAmount(in.SingleAmountMax)
	userTotalMin, userMinErr := parseNonNegativeAmount(in.UserTotalRechargeMin)
	userTotalMax, userMaxErr := parseNonNegativeAmount(in.UserTotalRechargeMax)
	if singleMinErr != nil || singleMaxErr != nil || userMinErr != nil || userMaxErr != nil ||
		!validDecimalRange(singleAmountMin, singleAmountMax) ||
		!validDecimalRange(userTotalMin, userTotalMax) ||
		!validNonNegativeRange(in.MemberLevelMin, in.MemberLevelMax) ||
		!validNonNegativeRange(in.KycLevelMin, in.KycLevelMax) {
		return paymentErrorResp(l.ctx, i18n.InvalidPaymentAmountRange), nil
	}
	if !validJSONArray(in.AllowTags) || !validJSONArray(in.DenyTags) {
		return paymentErrorResp(l.ctx, i18n.InvalidPaymentJSON), nil
	}
	if relationResp, err := validateChannelTenant(l.ctx, l.svcCtx, in.ChannelId, tenantID); err != nil {
		return nil, err
	} else if relationResp != nil {
		return relationResp, nil
	}

	now := utils.NowMillis()
	allowNewUser := int64(in.AllowNewUser)
	if common.YesNo(in.AllowNewUser) == common.YesNo_YES_NO_UNKNOWN {
		allowNewUser = int64(common.YesNo_YES_NO_YES)
	}
	allowOldUser := int64(in.AllowOldUser)
	if common.YesNo(in.AllowOldUser) == common.YesNo_YES_NO_UNKNOWN {
		allowOldUser = int64(common.YesNo_YES_NO_YES)
	}
	rule := &models.TTenantPayChannelRule{
		TenantId:             tenantID,
		ChannelId:            in.ChannelId,
		RuleName:             in.RuleName,
		Priority:             in.Priority,
		Enabled:              enableToModel(in.Enabled, int64(common.Enable_ENABLE_ENABLED)),
		SingleAmountMin:      singleAmountMin,
		SingleAmountMax:      singleAmountMax,
		UserTotalRechargeMin: userTotalMin,
		UserTotalRechargeMax: userTotalMax,
		MemberLevelMin:       in.MemberLevelMin,
		MemberLevelMax:       in.MemberLevelMax,
		KycLevelMin:          in.KycLevelMin,
		KycLevelMax:          in.KycLevelMax,
		AllowNewUser:         allowNewUser,
		AllowOldUser:         allowOldUser,
		AllowTags:            nullableString(in.AllowTags),
		DenyTags:             nullableString(in.DenyTags),
		Remark:               sql.NullString{String: in.Remark, Valid: true},
		CreateTimes:          now,
		UpdateTimes:          now,
	}

	_, err = l.svcCtx.TenantPayChannelRuleModel.Insert(l.ctx, rule)
	if err != nil {
		l.Logger.Errorf("%s error: %s", errLogic, err.Error())
		return nil, err
	}

	l.Logger.Infof("Create tenant pay channel rule success: %s", in.RuleName)

	return &payment.CommonResp{
		Base: helper.OkResp(),
	}, nil
}
