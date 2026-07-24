package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TLiquidityEventOutboxModel = (*customTLiquidityEventOutboxModel)(nil)

type (
	// TLiquidityEventOutboxModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTLiquidityEventOutboxModel.
	TLiquidityEventOutboxModel interface {
		tLiquidityEventOutboxModel
	}

	customTLiquidityEventOutboxModel struct {
		*defaultTLiquidityEventOutboxModel
	}
)

// NewTLiquidityEventOutboxModel returns a model for the database table.
func NewTLiquidityEventOutboxModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TLiquidityEventOutboxModel {
	return &customTLiquidityEventOutboxModel{
		defaultTLiquidityEventOutboxModel: newTLiquidityEventOutboxModel(conn, c, opts...),
	}
}
