package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TContractPositionHistoryModel = (*customTContractPositionHistoryModel)(nil)

type (
	// TContractPositionHistoryModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTContractPositionHistoryModel.
	TContractPositionHistoryModel interface {
		tContractPositionHistoryModel
	}

	customTContractPositionHistoryModel struct {
		*defaultTContractPositionHistoryModel
	}
)

// NewTContractPositionHistoryModel returns a model for the database table.
func NewTContractPositionHistoryModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TContractPositionHistoryModel {
	return &customTContractPositionHistoryModel{
		defaultTContractPositionHistoryModel: newTContractPositionHistoryModel(conn, c, opts...),
	}
}
