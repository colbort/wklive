package logic

import (
	"context"
	"database/sql"
	"time"

	"wklive/common/helper"
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
func (l *CreateTenantPayChannelRuleLogic) CreateTenantPayChannelRule(in *payment.CreateTenantPayChannelRuleReq) (*payment.AdminCommonResp, error) {
	var (
		errLogic = "CreateTenantPayChannelRule"
	)

	now := time.Now().UnixMilli()
	rule := &models.TTenantPayChannelRule{
		TenantId:             in.TenantId,
		ChannelId:            in.ChannelId,
		RuleName:             in.RuleName,
		Priority:             in.Priority,
		Status:               int64(in.Status),
		SingleAmountMin:      in.SingleAmountMin,
		SingleAmountMax:      in.SingleAmountMax,
		UserTotalRechargeMin: in.UserTotalRechargeMin,
		UserTotalRechargeMax: in.UserTotalRechargeMax,
		MemberLevelMin:       in.MemberLevelMin,
		MemberLevelMax:       in.MemberLevelMax,
		KycLevelMin:          in.KycLevelMin,
		KycLevelMax:          in.KycLevelMax,
		AllowNewUser:         in.AllowNewUser,
		AllowOldUser:         in.AllowOldUser,
		AllowTags:            sql.NullString{String: in.AllowTags, Valid: true},
		DenyTags:             sql.NullString{String: in.DenyTags, Valid: true},
		Remark:               sql.NullString{String: in.Remark, Valid: true},
		CreateTimes:          now,
		UpdateTimes:          now,
	}

	_, err := l.svcCtx.TenantPayChannelRuleModel.Insert(l.ctx, rule)
	if err != nil {
		l.Logger.Errorf("%s error: %s", errLogic, err.Error())
		return nil, err
	}

	l.Logger.Infof("Create tenant pay channel rule success: %s", in.RuleName)

	return &payment.AdminCommonResp{
		Base: helper.OkResp(),
	}, nil
}
