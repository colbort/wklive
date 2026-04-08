package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionAccountModel = (*customTOptionAccountModel)(nil)

type (
	// TOptionAccountModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionAccountModel.
	TOptionAccountModel interface {
		tOptionAccountModel
	}

	customTOptionAccountModel struct {
		*defaultTOptionAccountModel
	}
)

// NewTOptionAccountModel returns a model for the database table.
func NewTOptionAccountModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionAccountModel {
	return &customTOptionAccountModel{
		defaultTOptionAccountModel: newTOptionAccountModel(conn, c, opts...),
	}
}
