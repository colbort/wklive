package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionSettlementPriceModel = (*customTOptionSettlementPriceModel)(nil)

type (
	// TOptionSettlementPriceModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionSettlementPriceModel.
	TOptionSettlementPriceModel interface {
		tOptionSettlementPriceModel
	}

	customTOptionSettlementPriceModel struct {
		*defaultTOptionSettlementPriceModel
	}
)

// NewTOptionSettlementPriceModel returns a model for the database table.
func NewTOptionSettlementPriceModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionSettlementPriceModel {
	return &customTOptionSettlementPriceModel{
		defaultTOptionSettlementPriceModel: newTOptionSettlementPriceModel(conn, c, opts...),
	}
}
