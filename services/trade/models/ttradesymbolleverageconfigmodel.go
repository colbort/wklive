package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TTradeSymbolLeverageConfigModel = (*customTTradeSymbolLeverageConfigModel)(nil)

type (
	TTradeSymbolLeverageConfigModel interface {
		tTradeSymbolLeverageConfigModel
	}

	customTTradeSymbolLeverageConfigModel struct {
		*defaultTTradeSymbolLeverageConfigModel
	}
)

func NewTTradeSymbolLeverageConfigModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TTradeSymbolLeverageConfigModel {
	return &customTTradeSymbolLeverageConfigModel{
		defaultTTradeSymbolLeverageConfigModel: newTTradeSymbolLeverageConfigModel(conn, c, opts...),
	}
}
