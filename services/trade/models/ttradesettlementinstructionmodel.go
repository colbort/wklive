package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TTradeSettlementInstructionModel = (*customTTradeSettlementInstructionModel)(nil)

type (
	// TTradeSettlementInstructionModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTTradeSettlementInstructionModel.
	TTradeSettlementInstructionModel interface {
		tTradeSettlementInstructionModel
	}

	customTTradeSettlementInstructionModel struct {
		*defaultTTradeSettlementInstructionModel
	}
)

// NewTTradeSettlementInstructionModel returns a model for the database table.
func NewTTradeSettlementInstructionModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TTradeSettlementInstructionModel {
	return &customTTradeSettlementInstructionModel{
		defaultTTradeSettlementInstructionModel: newTTradeSettlementInstructionModel(conn, c, opts...),
	}
}
