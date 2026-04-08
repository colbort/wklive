package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TStakeOrderModel = (*customTStakeOrderModel)(nil)

type (
	// TStakeOrderModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTStakeOrderModel.
	TStakeOrderModel interface {
		tStakeOrderModel
	}

	customTStakeOrderModel struct {
		*defaultTStakeOrderModel
	}
)

// NewTStakeOrderModel returns a model for the database table.
func NewTStakeOrderModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TStakeOrderModel {
	return &customTStakeOrderModel{
		defaultTStakeOrderModel: newTStakeOrderModel(conn, c, opts...),
	}
}
