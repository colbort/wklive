package logic

import (
	"context"
	"errors"
	"github.com/zeromicro/go-zero/core/logx"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"
	"wklive/services/payment/models"
)

type GetTenantPayChannelRuleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTenantPayChannelRuleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTenantPayChannelRuleLogic {
	return &GetTenantPayChannelRuleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取通道规则详情
func (l *GetTenantPayChannelRuleLogic) GetTenantPayChannelRule(in *payment.GetTenantPayChannelRuleReq) (*payment.GetTenantPayChannelRuleResp, error) {
	var (
		errLogic = "GetTenantPayChannelRule"
	)

	rule, err := l.svcCtx.TenantPayChannelRuleModel.FindOne(l.ctx, in.Id)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		l.Logger.Errorf("%s error: %s", errLogic, err.Error())
		return nil, err
	}

	if rule == nil {
		return &payment.GetTenantPayChannelRuleResp{
			Base: helper.GetErrResp(404, i18n.Translate(i18n.NotifyChannelRuleNotFound, l.ctx)),
		}, nil
	}

	return &payment.GetTenantPayChannelRuleResp{
		Base: helper.OkResp(),
		Data: &payment.TenantPayChannelRule{
			Id:                   rule.Id,
			TenantId:             rule.TenantId,
			ChannelId:            rule.ChannelId,
			RuleName:             rule.RuleName,
			Priority:             rule.Priority,
			Status:               payment.CommonStatus(rule.Status),
			SingleAmountMin:      rule.SingleAmountMin,
			SingleAmountMax:      rule.SingleAmountMax,
			UserTotalRechargeMin: rule.UserTotalRechargeMin,
			UserTotalRechargeMax: rule.UserTotalRechargeMax,
			MemberLevelMin:       rule.MemberLevelMin,
			MemberLevelMax:       rule.MemberLevelMax,
			KycLevelMin:          rule.KycLevelMin,
			KycLevelMax:          rule.KycLevelMax,
			AllowNewUser:         rule.AllowNewUser,
			AllowOldUser:         rule.AllowOldUser,
			AllowTags:            rule.AllowTags.String,
			DenyTags:             rule.DenyTags.String,
			Remark:               rule.Remark.String,
			CreateTimes:          rule.CreateTimes,
			UpdateTimes:          rule.UpdateTimes,
		},
	}, nil
}
