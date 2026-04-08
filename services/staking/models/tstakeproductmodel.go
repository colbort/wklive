package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TStakeProductModel = (*customTStakeProductModel)(nil)

type (
	// TStakeProductModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTStakeProductModel.
	TStakeProductModel interface {
		tStakeProductModel
	}

	customTStakeProductModel struct {
		*defaultTStakeProductModel
	}
)

// NewTStakeProductModel returns a model for the database table.
func NewTStakeProductModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TStakeProductModel {
	return &customTStakeProductModel{
		defaultTStakeProductModel: newTStakeProductModel(conn, c, opts...),
	}
}
