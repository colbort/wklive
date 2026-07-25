package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TPayRequestLogModel = (*customTPayRequestLogModel)(nil)

type (
	// TPayRequestLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTPayRequestLogModel.
	TPayRequestLogModel interface {
		tPayRequestLogModel
	}

	customTPayRequestLogModel struct {
		*defaultTPayRequestLogModel
	}
)

// NewTPayRequestLogModel returns a model for the database table.
func NewTPayRequestLogModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TPayRequestLogModel {
	return &customTPayRequestLogModel{
		defaultTPayRequestLogModel: newTPayRequestLogModel(conn, c, opts...),
	}
}
