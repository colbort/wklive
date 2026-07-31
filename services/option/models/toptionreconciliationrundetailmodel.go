package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionReconciliationRunDetailModel = (*customTOptionReconciliationRunDetailModel)(nil)

type (
	// TOptionReconciliationRunDetailModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionReconciliationRunDetailModel.
	TOptionReconciliationRunDetailModel interface {
		tOptionReconciliationRunDetailModel
	}

	customTOptionReconciliationRunDetailModel struct {
		*defaultTOptionReconciliationRunDetailModel
	}
)

// NewTOptionReconciliationRunDetailModel returns a model for the database table.
func NewTOptionReconciliationRunDetailModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionReconciliationRunDetailModel {
	return &customTOptionReconciliationRunDetailModel{
		defaultTOptionReconciliationRunDetailModel: newTOptionReconciliationRunDetailModel(conn, c, opts...),
	}
}
