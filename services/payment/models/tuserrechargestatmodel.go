package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TUserRechargeStatModel = (*customTUserRechargeStatModel)(nil)

type (
	// TUserRechargeStatModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTUserRechargeStatModel.
	TUserRechargeStatModel interface {
		tUserRechargeStatModel
	}

	customTUserRechargeStatModel struct {
		*defaultTUserRechargeStatModel
	}
)

// NewTUserRechargeStatModel returns a model for the database table.
func NewTUserRechargeStatModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TUserRechargeStatModel {
	return &customTUserRechargeStatModel{
		defaultTUserRechargeStatModel: newTUserRechargeStatModel(conn, c, opts...),
	}
}
