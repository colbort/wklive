package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TRiskUserTradeLimitModel = (*customTRiskUserTradeLimitModel)(nil)

type (
	// TRiskUserTradeLimitModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTRiskUserTradeLimitModel.
	TRiskUserTradeLimitModel interface {
		tRiskUserTradeLimitModel
	}

	customTRiskUserTradeLimitModel struct {
		*defaultTRiskUserTradeLimitModel
	}
)

// NewTRiskUserTradeLimitModel returns a model for the database table.
func NewTRiskUserTradeLimitModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TRiskUserTradeLimitModel {
	return &customTRiskUserTradeLimitModel{
		defaultTRiskUserTradeLimitModel: newTRiskUserTradeLimitModel(conn, c, opts...),
	}
}
