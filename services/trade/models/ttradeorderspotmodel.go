package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TTradeOrderSpotModel = (*customTTradeOrderSpotModel)(nil)

type (
	// TTradeOrderSpotModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTTradeOrderSpotModel.
	TTradeOrderSpotModel interface {
		tTradeOrderSpotModel
	}

	customTTradeOrderSpotModel struct {
		*defaultTTradeOrderSpotModel
	}
)

// NewTTradeOrderSpotModel returns a model for the database table.
func NewTTradeOrderSpotModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TTradeOrderSpotModel {
	return &customTTradeOrderSpotModel{
		defaultTTradeOrderSpotModel: newTTradeOrderSpotModel(conn, c, opts...),
	}
}
