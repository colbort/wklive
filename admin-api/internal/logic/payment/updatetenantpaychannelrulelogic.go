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

type UpdateTenantPayChannelRuleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateTenantPayChannelRuleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTenantPayChannelRuleLogic {
	return &UpdateTenantPayChannelRuleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateTenantPayChannelRuleLogic) UpdateTenantPayChannelRule(req *types.UpdateTenantPayChannelRuleReq) (resp *types.RespBase, err error) {
	result, err := l.svcCtx.PaymentCli.UpdateTenantPayChannelRule(l.ctx, &payment.UpdateTenantPayChannelRuleReq{
		Id:                   req.Id,
		TenantId:             req.TenantId,
		RuleName:             req.RuleName,
		Priority:             req.Priority,
		Status:               payment.CommonStatus(req.Status),
		SingleAmountMin:      req.SingleAmountMin,
		SingleAmountMax:      req.SingleAmountMax,
		UserTotalRechargeMin: req.UserTotalRechargeMin,
		UserTotalRechargeMax: req.UserTotalRechargeMax,
		MemberLevelMin:       req.MemberLevelMin,
		MemberLevelMax:       req.MemberLevelMax,
		KycLevelMin:          req.KycLevelMin,
		KycLevelMax:          req.KycLevelMax,
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
