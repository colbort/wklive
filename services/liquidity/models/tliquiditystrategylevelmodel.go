package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TLiquidityStrategyLevelModel = (*customTLiquidityStrategyLevelModel)(nil)

type (
	// TLiquidityStrategyLevelModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTLiquidityStrategyLevelModel.
	TLiquidityStrategyLevelModel interface {
		tLiquidityStrategyLevelModel
	}

	customTLiquidityStrategyLevelModel struct {
		*defaultTLiquidityStrategyLevelModel
	}
)

// NewTLiquidityStrategyLevelModel returns a model for the database table.
func NewTLiquidityStrategyLevelModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TLiquidityStrategyLevelModel {
	return &customTLiquidityStrategyLevelModel{
		defaultTLiquidityStrategyLevelModel: newTLiquidityStrategyLevelModel(conn, c, opts...),
	}
}
