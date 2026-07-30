package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionMarginLotApplicationModel = (*customTOptionMarginLotApplicationModel)(nil)

type (
	// TOptionMarginLotApplicationModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionMarginLotApplicationModel.
	TOptionMarginLotApplicationModel interface {
		tOptionMarginLotApplicationModel
	}

	customTOptionMarginLotApplicationModel struct {
		*defaultTOptionMarginLotApplicationModel
	}
)

// NewTOptionMarginLotApplicationModel returns a model for the database table.
func NewTOptionMarginLotApplicationModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionMarginLotApplicationModel {
	return &customTOptionMarginLotApplicationModel{
		defaultTOptionMarginLotApplicationModel: newTOptionMarginLotApplicationModel(conn, c, opts...),
	}
}
