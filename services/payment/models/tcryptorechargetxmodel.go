package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TCryptoRechargeTxModel = (*customTCryptoRechargeTxModel)(nil)

type (
	// TCryptoRechargeTxModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTCryptoRechargeTxModel.
	TCryptoRechargeTxModel interface {
		tCryptoRechargeTxModel
	}

	customTCryptoRechargeTxModel struct {
		*defaultTCryptoRechargeTxModel
	}
)

// NewTCryptoRechargeTxModel returns a model for the database table.
func NewTCryptoRechargeTxModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TCryptoRechargeTxModel {
	return &customTCryptoRechargeTxModel{
		defaultTCryptoRechargeTxModel: newTCryptoRechargeTxModel(conn, c, opts...),
	}
}
