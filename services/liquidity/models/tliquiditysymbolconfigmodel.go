package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TLiquiditySymbolConfigModel = (*customTLiquiditySymbolConfigModel)(nil)

type (
	// TLiquiditySymbolConfigModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTLiquiditySymbolConfigModel.
	TLiquiditySymbolConfigModel interface {
		tLiquiditySymbolConfigModel
	}

	customTLiquiditySymbolConfigModel struct {
		*defaultTLiquiditySymbolConfigModel
	}
)

// NewTLiquiditySymbolConfigModel returns a model for the database table.
func NewTLiquiditySymbolConfigModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TLiquiditySymbolConfigModel {
	return &customTLiquiditySymbolConfigModel{
		defaultTLiquiditySymbolConfigModel: newTLiquiditySymbolConfigModel(conn, c, opts...),
	}
}
