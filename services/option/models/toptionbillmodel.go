package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionBillModel = (*customTOptionBillModel)(nil)

type (
	// TOptionBillModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionBillModel.
	TOptionBillModel interface {
		tOptionBillModel
	}

	customTOptionBillModel struct {
		*defaultTOptionBillModel
	}
)

// NewTOptionBillModel returns a model for the database table.
func NewTOptionBillModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionBillModel {
	return &customTOptionBillModel{
		defaultTOptionBillModel: newTOptionBillModel(conn, c, opts...),
	}
}
