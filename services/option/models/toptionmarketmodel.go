package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionMarketModel = (*customTOptionMarketModel)(nil)

type (
	// TOptionMarketModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionMarketModel.
	TOptionMarketModel interface {
		tOptionMarketModel
	}

	customTOptionMarketModel struct {
		*defaultTOptionMarketModel
	}
)

// NewTOptionMarketModel returns a model for the database table.
func NewTOptionMarketModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionMarketModel {
	return &customTOptionMarketModel{
		defaultTOptionMarketModel: newTOptionMarketModel(conn, c, opts...),
	}
}
