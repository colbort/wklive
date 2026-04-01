package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TRechargeOrderModel = (*customTRechargeOrderModel)(nil)

type (
	// TRechargeOrderModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTRechargeOrderModel.
	TRechargeOrderModel interface {
		tRechargeOrderModel
	}

	customTRechargeOrderModel struct {
		*defaultTRechargeOrderModel
	}
)

// NewTRechargeOrderModel returns a model for the database table.
func NewTRechargeOrderModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TRechargeOrderModel {
	return &customTRechargeOrderModel{
		defaultTRechargeOrderModel: newTRechargeOrderModel(conn, c, opts...),
	}
}
