package logic

import (
	"context"
	"database/sql"

	"wklive/common/helper"
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
func (l *CreateTenantPayChannelRuleLogic) CreateTenantPayChannelRule(in *payment.CreateTenantPayChannelRuleReq) (*payment.AdminCommonResp, error) {
	var (
		errLogic = "CreateTenantPayChannelRule"
	)

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
		TenantId:             in.TenantId,
		ChannelId:            in.ChannelId,
		RuleName:             in.RuleName,
		Priority:             in.Priority,
		Enabled:              enableToModel(in.Enabled, int64(common.Enable_ENABLE_ENABLED)),
		SingleAmountMin:      in.SingleAmountMin,
		SingleAmountMax:      in.SingleAmountMax,
		UserTotalRechargeMin: in.UserTotalRechargeMin,
		UserTotalRechargeMax: in.UserTotalRechargeMax,
		MemberLevelMin:       in.MemberLevelMin,
		MemberLevelMax:       in.MemberLevelMax,
		KycLevelMin:          in.KycLevelMin,
		KycLevelMax:          in.KycLevelMax,
		AllowNewUser:         allowNewUser,
		AllowOldUser:         allowOldUser,
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
