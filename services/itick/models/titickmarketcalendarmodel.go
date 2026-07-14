package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TItickMarketCalendarModel = (*customTItickMarketCalendarModel)(nil)

type (
	// TItickMarketCalendarModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTItickMarketCalendarModel.
	TItickMarketCalendarModel interface {
		tItickMarketCalendarModel
	}

	customTItickMarketCalendarModel struct {
		*defaultTItickMarketCalendarModel
	}
)

// NewTItickMarketCalendarModel returns a model for the database table.
func NewTItickMarketCalendarModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TItickMarketCalendarModel {
	return &customTItickMarketCalendarModel{
		defaultTItickMarketCalendarModel: newTItickMarketCalendarModel(conn, c, opts...),
	}
}
