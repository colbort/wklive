package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TUserBankModel = (*customTUserBankModel)(nil)

type (
	// TUserBankModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTUserBankModel.
	TUserBankModel interface {
		tUserBankModel
	}

	customTUserBankModel struct {
		*defaultTUserBankModel
	}
)

// NewTUserBankModel returns a model for the database table.
func NewTUserBankModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TUserBankModel {
	return &customTUserBankModel{
		defaultTUserBankModel: newTUserBankModel(conn, c, opts...),
	}
}
