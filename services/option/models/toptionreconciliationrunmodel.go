package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionReconciliationRunModel = (*customTOptionReconciliationRunModel)(nil)

type (
	// TOptionReconciliationRunModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionReconciliationRunModel.
	TOptionReconciliationRunModel interface {
		tOptionReconciliationRunModel
	}

	customTOptionReconciliationRunModel struct {
		*defaultTOptionReconciliationRunModel
	}
)

// NewTOptionReconciliationRunModel returns a model for the database table.
func NewTOptionReconciliationRunModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionReconciliationRunModel {
	return &customTOptionReconciliationRunModel{
		defaultTOptionReconciliationRunModel: newTOptionReconciliationRunModel(conn, c, opts...),
	}
}
