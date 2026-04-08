package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionPositionModel = (*customTOptionPositionModel)(nil)

type (
	// TOptionPositionModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionPositionModel.
	TOptionPositionModel interface {
		tOptionPositionModel
	}

	customTOptionPositionModel struct {
		*defaultTOptionPositionModel
	}
)

// NewTOptionPositionModel returns a model for the database table.
func NewTOptionPositionModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionPositionModel {
	return &customTOptionPositionModel{
		defaultTOptionPositionModel: newTOptionPositionModel(conn, c, opts...),
	}
}
