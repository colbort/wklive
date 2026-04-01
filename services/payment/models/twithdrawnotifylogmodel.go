package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TWithdrawNotifyLogModel = (*customTWithdrawNotifyLogModel)(nil)

type (
	// TWithdrawNotifyLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTWithdrawNotifyLogModel.
	TWithdrawNotifyLogModel interface {
		tWithdrawNotifyLogModel
	}

	customTWithdrawNotifyLogModel struct {
		*defaultTWithdrawNotifyLogModel
	}
)

// NewTWithdrawNotifyLogModel returns a model for the database table.
func NewTWithdrawNotifyLogModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TWithdrawNotifyLogModel {
	return &customTWithdrawNotifyLogModel{
		defaultTWithdrawNotifyLogModel: newTWithdrawNotifyLogModel(conn, c, opts...),
	}
}
