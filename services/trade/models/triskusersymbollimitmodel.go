package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TRiskUserSymbolLimitModel = (*customTRiskUserSymbolLimitModel)(nil)

type (
	// TRiskUserSymbolLimitModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTRiskUserSymbolLimitModel.
	TRiskUserSymbolLimitModel interface {
		tRiskUserSymbolLimitModel
	}

	customTRiskUserSymbolLimitModel struct {
		*defaultTRiskUserSymbolLimitModel
	}
)

// NewTRiskUserSymbolLimitModel returns a model for the database table.
func NewTRiskUserSymbolLimitModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TRiskUserSymbolLimitModel {
	return &customTRiskUserSymbolLimitModel{
		defaultTRiskUserSymbolLimitModel: newTRiskUserSymbolLimitModel(conn, c, opts...),
	}
}
