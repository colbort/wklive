package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TTenantPayChannelModel = (*customTTenantPayChannelModel)(nil)

type (
	// TTenantPayChannelModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTTenantPayChannelModel.
	TTenantPayChannelModel interface {
		tTenantPayChannelModel
	}

	customTTenantPayChannelModel struct {
		*defaultTTenantPayChannelModel
	}
)

// NewTTenantPayChannelModel returns a model for the database table.
func NewTTenantPayChannelModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TTenantPayChannelModel {
	return &customTTenantPayChannelModel{
		defaultTTenantPayChannelModel: newTTenantPayChannelModel(conn, c, opts...),
	}
}
