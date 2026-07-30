package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TItickMarketSessionModel = (*customTMarketMarketSessionModel)(nil)

type (
	// TItickMarketSessionModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTMarketMarketSessionModel.
	TItickMarketSessionModel interface {
		tItickMarketSessionModel
	}

	customTMarketMarketSessionModel struct {
		*defaultTMarketMarketSessionModel
	}
)

// NewTMarketMarketSessionModel returns a model for the database table.
func NewTMarketMarketSessionModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TItickMarketSessionModel {
	return &customTMarketMarketSessionModel{
		defaultTMarketMarketSessionModel: newTMarketMarketSessionModel(conn, c, opts...),
	}
}
