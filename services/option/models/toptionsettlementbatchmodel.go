package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionSettlementBatchModel = (*customTOptionSettlementBatchModel)(nil)

type (
	// TOptionSettlementBatchModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionSettlementBatchModel.
	TOptionSettlementBatchModel interface {
		tOptionSettlementBatchModel
	}

	customTOptionSettlementBatchModel struct {
		*defaultTOptionSettlementBatchModel
	}
)

// NewTOptionSettlementBatchModel returns a model for the database table.
func NewTOptionSettlementBatchModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionSettlementBatchModel {
	return &customTOptionSettlementBatchModel{
		defaultTOptionSettlementBatchModel: newTOptionSettlementBatchModel(conn, c, opts...),
	}
}
