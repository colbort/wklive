package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TPayNotifyLogModel = (*customTPayNotifyLogModel)(nil)

type (
	// TPayNotifyLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTPayNotifyLogModel.
	TPayNotifyLogModel interface {
		tPayNotifyLogModel
	}

	customTPayNotifyLogModel struct {
		*defaultTPayNotifyLogModel
	}
)

// NewTPayNotifyLogModel returns a model for the database table.
func NewTPayNotifyLogModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TPayNotifyLogModel {
	return &customTPayNotifyLogModel{
		defaultTPayNotifyLogModel: newTPayNotifyLogModel(conn, c, opts...),
	}
}
