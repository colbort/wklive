package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TItickMarketSessionModel = (*customTItickMarketSessionModel)(nil)

type (
	// TItickMarketSessionModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTItickMarketSessionModel.
	TItickMarketSessionModel interface {
		tItickMarketSessionModel
	}

	customTItickMarketSessionModel struct {
		*defaultTItickMarketSessionModel
	}
)

// NewTItickMarketSessionModel returns a model for the database table.
func NewTItickMarketSessionModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TItickMarketSessionModel {
	return &customTItickMarketSessionModel{
		defaultTItickMarketSessionModel: newTItickMarketSessionModel(conn, c, opts...),
	}
}
