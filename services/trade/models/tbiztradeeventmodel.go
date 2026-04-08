package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TBizTradeEventModel = (*customTBizTradeEventModel)(nil)

type (
	// TBizTradeEventModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTBizTradeEventModel.
	TBizTradeEventModel interface {
		tBizTradeEventModel
	}

	customTBizTradeEventModel struct {
		*defaultTBizTradeEventModel
	}
)

// NewTBizTradeEventModel returns a model for the database table.
func NewTBizTradeEventModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TBizTradeEventModel {
	return &customTBizTradeEventModel{
		defaultTBizTradeEventModel: newTBizTradeEventModel(conn, c, opts...),
	}
}
