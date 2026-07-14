package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TItickMarketHolidayModel = (*customTItickMarketHolidayModel)(nil)

type (
	// TItickMarketHolidayModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTItickMarketHolidayModel.
	TItickMarketHolidayModel interface {
		tItickMarketHolidayModel
	}

	customTItickMarketHolidayModel struct {
		*defaultTItickMarketHolidayModel
	}
)

// NewTItickMarketHolidayModel returns a model for the database table.
func NewTItickMarketHolidayModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TItickMarketHolidayModel {
	return &customTItickMarketHolidayModel{
		defaultTItickMarketHolidayModel: newTItickMarketHolidayModel(conn, c, opts...),
	}
}
