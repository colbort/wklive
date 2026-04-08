package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TContractMarginAccountModel = (*customTContractMarginAccountModel)(nil)

type (
	// TContractMarginAccountModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTContractMarginAccountModel.
	TContractMarginAccountModel interface {
		tContractMarginAccountModel
	}

	customTContractMarginAccountModel struct {
		*defaultTContractMarginAccountModel
	}
)

// NewTContractMarginAccountModel returns a model for the database table.
func NewTContractMarginAccountModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TContractMarginAccountModel {
	return &customTContractMarginAccountModel{
		defaultTContractMarginAccountModel: newTContractMarginAccountModel(conn, c, opts...),
	}
}
