package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TTenantPayChannelRuleModel = (*customTTenantPayChannelRuleModel)(nil)

type (
	// TTenantPayChannelRuleModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTTenantPayChannelRuleModel.
	TTenantPayChannelRuleModel interface {
		tTenantPayChannelRuleModel
	}

	customTTenantPayChannelRuleModel struct {
		*defaultTTenantPayChannelRuleModel
	}
)

// NewTTenantPayChannelRuleModel returns a model for the database table.
func NewTTenantPayChannelRuleModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TTenantPayChannelRuleModel {
	return &customTTenantPayChannelRuleModel{
		defaultTTenantPayChannelRuleModel: newTTenantPayChannelRuleModel(conn, c, opts...),
	}
}
