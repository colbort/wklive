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

type GetTenantPayChannelRuleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTenantPayChannelRuleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTenantPayChannelRuleLogic {
	return &GetTenantPayChannelRuleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTenantPayChannelRuleLogic) GetTenantPayChannelRule(req *types.GetTenantPayChannelRuleReq) (resp *types.GetTenantPayChannelRuleResp, err error) {
	result, err := l.svcCtx.PaymentCli.GetTenantPayChannelRule(l.ctx, &payment.GetTenantPayChannelRuleReq{
		Id:       req.Id,
		TenantId: req.TenantId,
	})
	if err != nil {
		return nil, err
	}

	resp = &types.GetTenantPayChannelRuleResp{
		RespBase: types.RespBase{
			Code: result.Base.Code,
			Msg:  result.Base.Msg,
		},
		Data: types.TenantPayChannelRule{
			Id:                   result.Data.Id,
			TenantId:             result.Data.TenantId,
			ChannelId:            result.Data.ChannelId,
			RuleName:             result.Data.RuleName,
			Priority:             int64(result.Data.Priority),
			Status:               int64(result.Data.Status),
			SingleAmountMin:      result.Data.SingleAmountMin,
			SingleAmountMax:      result.Data.SingleAmountMax,
			UserTotalRechargeMin: result.Data.UserTotalRechargeMin,
			UserTotalRechargeMax: result.Data.UserTotalRechargeMax,
			MemberLevelMin:       int64(result.Data.MemberLevelMin),
			MemberLevelMax:       int64(result.Data.MemberLevelMax),
			KycLevelMin:          int64(result.Data.KycLevelMin),
			KycLevelMax:          int64(result.Data.KycLevelMax),
			AllowNewUser:         result.Data.AllowNewUser,
			AllowOldUser:         result.Data.AllowOldUser,
			AllowTags:            result.Data.AllowTags,
			DenyTags:             result.Data.DenyTags,
			Remark:               result.Data.Remark,
			CreateTime:           result.Data.CreateTime,
			UpdateTime:           result.Data.UpdateTime,
		},
	}
	return resp, nil
}
