package adminlogic

import (
	"context"
	"errors"

	"wklive/common/pageutil"
	"wklive/common/utils"
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
	if in.TenantId <= 0 {
		if tenantId, err := utils.GetTenantIdFromMd(l.ctx); err == nil {
			in.TenantId = tenantId
		}
	}
	rules, total, err := l.svcCtx.TenantPayChannelRuleModel.FindPage(l.ctx, in.ChannelId, in.Page.Cursor, in.Page.Limit)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	lastID := int64(0)
	if len(rules) > 0 {
		lastID = rules[len(rules)-1].Id
	}

	data := make([]*payment.TenantPayChannelRule, 0, len(rules))
	channelIDs := make([]int64, 0, len(rules))
	seenChannelIDs := make(map[int64]struct{}, len(rules))
	for _, rule := range rules {
		if _, exists := seenChannelIDs[rule.ChannelId]; !exists {
			seenChannelIDs[rule.ChannelId] = struct{}{}
			channelIDs = append(channelIDs, rule.ChannelId)
		}
	}
	channels, err := l.svcCtx.TenantPayChannelModel.FindByIDs(l.ctx, channelIDs)
	if err != nil {
		return nil, err
	}
	channelNames := make(map[int64]string, len(channels))
	for _, channel := range channels {
		channelNames[channel.Id] = channel.ChannelName
	}
	for _, r := range rules {
		item := toTenantPayChannelRuleProto(r)
		item.ChannelName = channelNames[r.ChannelId]
		data = append(data, item)
	}

	return &payment.ListTenantPayChannelRulesResp{
		Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, len(rules), total, lastID),
		Data: data,
	}, nil
}
