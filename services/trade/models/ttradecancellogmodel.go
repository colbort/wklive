package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TTradeCancelLogModel = (*customTTradeCancelLogModel)(nil)

type (
	// TTradeCancelLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTTradeCancelLogModel.
	TTradeCancelLogModel interface {
		tTradeCancelLogModel
	}

	customTTradeCancelLogModel struct {
		*defaultTTradeCancelLogModel
	}
)

// NewTTradeCancelLogModel returns a model for the database table.
func NewTTradeCancelLogModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TTradeCancelLogModel {
	return &customTTradeCancelLogModel{
		defaultTTradeCancelLogModel: newTTradeCancelLogModel(conn, c, opts...),
	}
}
