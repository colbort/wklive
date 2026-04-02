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

type CreateTenantPayChannelRuleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateTenantPayChannelRuleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTenantPayChannelRuleLogic {
	return &CreateTenantPayChannelRuleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateTenantPayChannelRuleLogic) CreateTenantPayChannelRule(req *types.CreateTenantPayChannelRuleReq) (resp *types.RespBase, err error) {
	result, err := l.svcCtx.PaymentCli.CreateTenantPayChannelRule(l.ctx, &payment.CreateTenantPayChannelRuleReq{
		TenantId:             req.TenantId,
		ChannelId:            req.ChannelId,
		RuleName:             req.RuleName,
		Priority:             int32(req.Priority),
		Status:               payment.CommonStatus(req.Status),
		SingleAmountMin:      req.SingleAmountMin,
		SingleAmountMax:      req.SingleAmountMax,
		UserTotalRechargeMin: req.UserTotalRechargeMin,
		UserTotalRechargeMax: req.UserTotalRechargeMax,
		MemberLevelMin:       int32(req.MemberLevelMin),
		MemberLevelMax:       int32(req.MemberLevelMax),
		KycLevelMin:          int32(req.KycLevelMin),
		KycLevelMax:          int32(req.KycLevelMax),
		AllowNewUser:         req.AllowNewUser,
		AllowOldUser:         req.AllowOldUser,
		AllowTags:            req.AllowTags,
		DenyTags:             req.DenyTags,
		Remark:               req.Remark,
	})
	if err != nil {
		return nil, err
	}

	resp = &types.RespBase{
		Code: result.Base.Code,
		Msg:  result.Base.Msg,
	}
	return resp, nil
}
