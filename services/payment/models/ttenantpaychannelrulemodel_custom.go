package models

import "context"

type TenantPayChannelRuleModel interface {
	tTenantPayChannelRuleModel
	FindPage(ctx context.Context, channelId int64, cursor int64, limit int64) ([]*TTenantPayChannelRule, int64, error)
}
