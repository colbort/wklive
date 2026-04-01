package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TWithdrawOrderModel = (*customTWithdrawOrderModel)(nil)

type (
	// TWithdrawOrderModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTWithdrawOrderModel.
	TWithdrawOrderModel interface {
		tWithdrawOrderModel
	}

	customTWithdrawOrderModel struct {
		*defaultTWithdrawOrderModel
	}
)

// NewTWithdrawOrderModel returns a model for the database table.
func NewTWithdrawOrderModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TWithdrawOrderModel {
	return &customTWithdrawOrderModel{
		defaultTWithdrawOrderModel: newTWithdrawOrderModel(conn, c, opts...),
	}
}
