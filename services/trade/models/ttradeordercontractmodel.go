package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TTradeOrderContractModel = (*customTTradeOrderContractModel)(nil)

type (
	// TTradeOrderContractModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTTradeOrderContractModel.
	TTradeOrderContractModel interface {
		tTradeOrderContractModel
	}

	customTTradeOrderContractModel struct {
		*defaultTTradeOrderContractModel
	}
)

// NewTTradeOrderContractModel returns a model for the database table.
func NewTTradeOrderContractModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TTradeOrderContractModel {
	return &customTTradeOrderContractModel{
		defaultTTradeOrderContractModel: newTTradeOrderContractModel(conn, c, opts...),
	}
}
