package logic

import (
	"context"
	"errors"

	"wklive/common/pageutil"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"
	"wklive/services/payment/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListTenantPayChannelRulesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListTenantPayChannelRulesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTenantPayChannelRulesLogic {
	return &ListTenantPayChannelRulesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 通道规则列表
func (l *ListTenantPayChannelRulesLogic) ListTenantPayChannelRules(in *payment.ListTenantPayChannelRulesReq) (*payment.ListTenantPayChannelRulesResp, error) {
	rules, total, err := l.svcCtx.TenantPayChannelRuleModel.FindPage(l.ctx, in.ChannelId, in.Page.Cursor, in.Page.Limit)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	lastID := int64(0)
	if len(rules) > 0 {
		lastID = rules[len(rules)-1].Id
	}

	data := make([]*payment.TenantPayChannelRule, 0, len(rules))
	for _, r := range rules {
		data = append(data, toTenantPayChannelRuleProto(r))
	}

	return &payment.ListTenantPayChannelRulesResp{
		Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, len(rules), total, lastID),
		Data: data,
	}, nil
}
