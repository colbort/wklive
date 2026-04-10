package logic

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"wklive/common/helper"
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
func (l *UpdateTenantPayChannelRuleLogic) UpdateTenantPayChannelRule(in *payment.UpdateTenantPayChannelRuleReq) (*payment.AdminCommonResp, error) {
	var (
		errLogic = "UpdateTenantPayChannelRule"
	)

	// 查询规则是否存在
	rule, err := l.svcCtx.TenantPayChannelRuleModel.FindOne(l.ctx, in.Id)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		l.Logger.Errorf("%s error: %s", errLogic, err.Error())
		return nil, err
	}

	if rule == nil {
		return &payment.AdminCommonResp{
			Base: helper.GetErrResp(404, "规则不存在"),
		}, nil
	}

	now := time.Now().UnixMilli()
	if in.RuleName != "" {
		rule.RuleName = in.RuleName
	}
	if in.Priority != 0 {
		rule.Priority = in.Priority
	}
	if in.Status != 0 {
		rule.Status = int64(in.Status)
	}
	if in.SingleAmountMin != 0 {
		rule.SingleAmountMin = in.SingleAmountMin
	}
	if in.SingleAmountMax != 0 {
		rule.SingleAmountMax = in.SingleAmountMax
	}
	if in.UserTotalRechargeMin != 0 {
		rule.UserTotalRechargeMin = in.UserTotalRechargeMin
	}
	if in.UserTotalRechargeMax != 0 {
		rule.UserTotalRechargeMax = in.UserTotalRechargeMax
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
	if in.AllowNewUser != 0 {
		rule.AllowNewUser = in.AllowNewUser
	}
	if in.AllowOldUser != 0 {
		rule.AllowOldUser = in.AllowOldUser
	}
	if in.AllowTags != "" {
		rule.AllowTags = sql.NullString{String: in.AllowTags, Valid: true}
	}
	if in.DenyTags != "" {
		rule.DenyTags = sql.NullString{String: in.DenyTags, Valid: true}
	}
	if in.Remark != "" {
		rule.Remark = sql.NullString{String: in.Remark, Valid: true}
	}
	rule.UpdateTimes = now

	err = l.svcCtx.TenantPayChannelRuleModel.Update(l.ctx, rule)
	if err != nil {
		l.Logger.Errorf("%s error: %s", errLogic, err.Error())
		return nil, err
	}

	l.Logger.Infof("Update tenant pay channel rule success: %d", in.Id)

	return &payment.AdminCommonResp{
		Base: helper.OkResp(),
	}, nil
}
