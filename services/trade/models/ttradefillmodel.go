package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TTradeFillModel = (*customTTradeFillModel)(nil)

type (
	// TTradeFillModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTTradeFillModel.
	TTradeFillModel interface {
		tTradeFillModel
	}

	customTTradeFillModel struct {
		*defaultTTradeFillModel
	}
)

// NewTTradeFillModel returns a model for the database table.
func NewTTradeFillModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TTradeFillModel {
	return &customTTradeFillModel{
		defaultTTradeFillModel: newTTradeFillModel(conn, c, opts...),
	}
}
