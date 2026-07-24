package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TLiquidityEventInboxModel = (*customTLiquidityEventInboxModel)(nil)

type (
	// TLiquidityEventInboxModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTLiquidityEventInboxModel.
	TLiquidityEventInboxModel interface {
		tLiquidityEventInboxModel
	}

	customTLiquidityEventInboxModel struct {
		*defaultTLiquidityEventInboxModel
	}
)

// NewTLiquidityEventInboxModel returns a model for the database table.
func NewTLiquidityEventInboxModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TLiquidityEventInboxModel {
	return &customTLiquidityEventInboxModel{
		defaultTLiquidityEventInboxModel: newTLiquidityEventInboxModel(conn, c, opts...),
	}
}
