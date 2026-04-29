package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TCryptoRechargeAddressModel = (*customTCryptoRechargeAddressModel)(nil)

type (
	// TCryptoRechargeAddressModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTCryptoRechargeAddressModel.
	TCryptoRechargeAddressModel interface {
		tCryptoRechargeAddressModel
	}

	customTCryptoRechargeAddressModel struct {
		*defaultTCryptoRechargeAddressModel
	}
)

// NewTCryptoRechargeAddressModel returns a model for the database table.
func NewTCryptoRechargeAddressModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TCryptoRechargeAddressModel {
	return &customTCryptoRechargeAddressModel{
		defaultTCryptoRechargeAddressModel: newTCryptoRechargeAddressModel(conn, c, opts...),
	}
}
