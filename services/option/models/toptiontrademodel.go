package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionTradeModel = (*customTOptionTradeModel)(nil)

type (
	// TOptionTradeModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionTradeModel.
	TOptionTradeModel interface {
		tOptionTradeModel
	}

	customTOptionTradeModel struct {
		*defaultTOptionTradeModel
	}
)

// NewTOptionTradeModel returns a model for the database table.
func NewTOptionTradeModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionTradeModel {
	return &customTOptionTradeModel{
		defaultTOptionTradeModel: newTOptionTradeModel(conn, c, opts...),
	}
}
