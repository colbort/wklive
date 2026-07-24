package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TLiquidityProviderModel = (*customTLiquidityProviderModel)(nil)

type (
	// TLiquidityProviderModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTLiquidityProviderModel.
	TLiquidityProviderModel interface {
		tLiquidityProviderModel
	}

	customTLiquidityProviderModel struct {
		*defaultTLiquidityProviderModel
	}
)

// NewTLiquidityProviderModel returns a model for the database table.
func NewTLiquidityProviderModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TLiquidityProviderModel {
	return &customTLiquidityProviderModel{
		defaultTLiquidityProviderModel: newTLiquidityProviderModel(conn, c, opts...),
	}
}
