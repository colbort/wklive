package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TRiskOrderCheckLogModel = (*customTRiskOrderCheckLogModel)(nil)

type (
	// TRiskOrderCheckLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTRiskOrderCheckLogModel.
	TRiskOrderCheckLogModel interface {
		tRiskOrderCheckLogModel
	}

	customTRiskOrderCheckLogModel struct {
		*defaultTRiskOrderCheckLogModel
	}
)

// NewTRiskOrderCheckLogModel returns a model for the database table.
func NewTRiskOrderCheckLogModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TRiskOrderCheckLogModel {
	return &customTRiskOrderCheckLogModel{
		defaultTRiskOrderCheckLogModel: newTRiskOrderCheckLogModel(conn, c, opts...),
	}
}
