package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionSettlementModel = (*customTOptionSettlementModel)(nil)

type (
	// TOptionSettlementModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionSettlementModel.
	TOptionSettlementModel interface {
		tOptionSettlementModel
	}

	customTOptionSettlementModel struct {
		*defaultTOptionSettlementModel
	}
)

// NewTOptionSettlementModel returns a model for the database table.
func NewTOptionSettlementModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionSettlementModel {
	return &customTOptionSettlementModel{
		defaultTOptionSettlementModel: newTOptionSettlementModel(conn, c, opts...),
	}
}
