package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TCryptoWalletAccountModel = (*customTCryptoWalletAccountModel)(nil)

type (
	// TCryptoWalletAccountModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTCryptoWalletAccountModel.
	TCryptoWalletAccountModel interface {
		tCryptoWalletAccountModel
	}

	customTCryptoWalletAccountModel struct {
		*defaultTCryptoWalletAccountModel
	}
)

// NewTCryptoWalletAccountModel returns a model for the database table.
func NewTCryptoWalletAccountModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TCryptoWalletAccountModel {
	return &customTCryptoWalletAccountModel{
		defaultTCryptoWalletAccountModel: newTCryptoWalletAccountModel(conn, c, opts...),
	}
}
