// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/proto/common"
	"wklive/proto/payment"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListTenantPayChannelRulesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListTenantPayChannelRulesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTenantPayChannelRulesLogic {
	return &ListTenantPayChannelRulesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListTenantPayChannelRulesLogic) ListTenantPayChannelRules(req *types.ListTenantPayChannelRulesReq) (resp *types.ListTenantPayChannelRulesResp, err error) {
	result, err := l.svcCtx.PaymentCli.ListTenantPayChannelRules(l.ctx, &payment.ListTenantPayChannelRulesReq{
		Page: &common.PageReq{
			Cursor: req.Cursor,
			Limit:  req.Limit,
		},
		TenantId:  req.TenantId,
		ChannelId: req.ChannelId,
		Status:    payment.CommonStatus(req.Status),
	})
	if err != nil {
		return nil, err
	}

	data := make([]types.TenantPayChannelRule, len(result.Data))
	for i, item := range result.Data {
		data[i] = types.TenantPayChannelRule{
			Id:                   item.Id,
			TenantId:             item.TenantId,
			ChannelId:            item.ChannelId,
			RuleName:             item.RuleName,
			Priority:             int64(item.Priority),
			Status:               int64(item.Status),
			SingleAmountMin:      item.SingleAmountMin,
			SingleAmountMax:      item.SingleAmountMax,
			UserTotalRechargeMin: item.UserTotalRechargeMin,
			UserTotalRechargeMax: item.UserTotalRechargeMax,
			MemberLevelMin:       int64(item.MemberLevelMin),
			MemberLevelMax:       int64(item.MemberLevelMax),
			KycLevelMin:          int64(item.KycLevelMin),
			KycLevelMax:          int64(item.KycLevelMax),
			AllowNewUser:         item.AllowNewUser,
			AllowOldUser:         item.AllowOldUser,
			AllowTags:            item.AllowTags,
			DenyTags:             item.DenyTags,
			Remark:               item.Remark,
			CreateTimes:           item.CreateTimes,
			UpdateTimes:           item.UpdateTimes,
		}
	}

	resp = &types.ListTenantPayChannelRulesResp{
		RespBase: types.RespBase{
			Code:       result.Base.Code,
			Msg:        result.Base.Msg,
			Total:      result.Base.Total,
			HasNext:    result.Base.HasNext,
			HasPrev:    result.Base.HasPrev,
			NextCursor: result.Base.NextCursor,
			PrevCursor: result.Base.PrevCursor,
		},
		Data: data,
	}
	return resp, nil
}
