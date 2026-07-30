package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionSettlementDetailModel = (*customTOptionSettlementDetailModel)(nil)

type (
	// TOptionSettlementDetailModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionSettlementDetailModel.
	TOptionSettlementDetailModel interface {
		tOptionSettlementDetailModel
	}

	customTOptionSettlementDetailModel struct {
		*defaultTOptionSettlementDetailModel
	}
)

// NewTOptionSettlementDetailModel returns a model for the database table.
func NewTOptionSettlementDetailModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionSettlementDetailModel {
	return &customTOptionSettlementDetailModel{
		defaultTOptionSettlementDetailModel: newTOptionSettlementDetailModel(conn, c, opts...),
	}
}
