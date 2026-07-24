package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TLiquidityRiskEventModel = (*customTLiquidityRiskEventModel)(nil)

type (
	// TLiquidityRiskEventModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTLiquidityRiskEventModel.
	TLiquidityRiskEventModel interface {
		tLiquidityRiskEventModel
	}

	customTLiquidityRiskEventModel struct {
		*defaultTLiquidityRiskEventModel
	}
)

// NewTLiquidityRiskEventModel returns a model for the database table.
func NewTLiquidityRiskEventModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TLiquidityRiskEventModel {
	return &customTLiquidityRiskEventModel{
		defaultTLiquidityRiskEventModel: newTLiquidityRiskEventModel(conn, c, opts...),
	}
}
