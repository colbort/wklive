package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TTradeSymbolContractModel = (*customTTradeSymbolContractModel)(nil)

type (
	// TTradeSymbolContractModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTTradeSymbolContractModel.
	TTradeSymbolContractModel interface {
		tTradeSymbolContractModel
	}

	customTTradeSymbolContractModel struct {
		*defaultTTradeSymbolContractModel
	}
)

// NewTTradeSymbolContractModel returns a model for the database table.
func NewTTradeSymbolContractModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TTradeSymbolContractModel {
	return &customTTradeSymbolContractModel{
		defaultTTradeSymbolContractModel: newTTradeSymbolContractModel(conn, c, opts...),
	}
}
