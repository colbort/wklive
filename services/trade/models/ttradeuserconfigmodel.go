package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TTradeUserConfigModel = (*customTTradeUserConfigModel)(nil)

type (
	// TTradeUserConfigModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTTradeUserConfigModel.
	TTradeUserConfigModel interface {
		tTradeUserConfigModel
	}

	customTTradeUserConfigModel struct {
		*defaultTTradeUserConfigModel
	}
)

// NewTTradeUserConfigModel returns a model for the database table.
func NewTTradeUserConfigModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TTradeUserConfigModel {
	return &customTTradeUserConfigModel{
		defaultTTradeUserConfigModel: newTTradeUserConfigModel(conn, c, opts...),
	}
}
