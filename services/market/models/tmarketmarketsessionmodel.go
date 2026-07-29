package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TMarketMarketSessionModel = (*customTMarketMarketSessionModel)(nil)

type (
	// TMarketMarketSessionModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTMarketMarketSessionModel.
	TMarketMarketSessionModel interface {
		tMarketMarketSessionModel
	}

	customTMarketMarketSessionModel struct {
		*defaultTMarketMarketSessionModel
	}
)

// NewTMarketMarketSessionModel returns a model for the database table.
func NewTMarketMarketSessionModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TMarketMarketSessionModel {
	return &customTMarketMarketSessionModel{
		defaultTMarketMarketSessionModel: newTMarketMarketSessionModel(conn, c, opts...),
	}
}
