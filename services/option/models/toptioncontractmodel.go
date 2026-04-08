package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionContractModel = (*customTOptionContractModel)(nil)

type (
	// TOptionContractModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionContractModel.
	TOptionContractModel interface {
		tOptionContractModel
	}

	customTOptionContractModel struct {
		*defaultTOptionContractModel
	}
)

// NewTOptionContractModel returns a model for the database table.
func NewTOptionContractModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionContractModel {
	return &customTOptionContractModel{
		defaultTOptionContractModel: newTOptionContractModel(conn, c, opts...),
	}
}
