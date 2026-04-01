package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TRechargeNotifyLogModel = (*customTRechargeNotifyLogModel)(nil)

type (
	// TRechargeNotifyLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTRechargeNotifyLogModel.
	TRechargeNotifyLogModel interface {
		tRechargeNotifyLogModel
	}

	customTRechargeNotifyLogModel struct {
		*defaultTRechargeNotifyLogModel
	}
)

// NewTRechargeNotifyLogModel returns a model for the database table.
func NewTRechargeNotifyLogModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TRechargeNotifyLogModel {
	return &customTRechargeNotifyLogModel{
		defaultTRechargeNotifyLogModel: newTRechargeNotifyLogModel(conn, c, opts...),
	}
}
