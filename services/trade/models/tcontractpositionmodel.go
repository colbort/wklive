package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TContractPositionModel = (*customTContractPositionModel)(nil)

type (
	// TContractPositionModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTContractPositionModel.
	TContractPositionModel interface {
		tContractPositionModel
	}

	customTContractPositionModel struct {
		*defaultTContractPositionModel
	}
)

// NewTContractPositionModel returns a model for the database table.
func NewTContractPositionModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TContractPositionModel {
	return &customTContractPositionModel{
		defaultTContractPositionModel: newTContractPositionModel(conn, c, opts...),
	}
}
