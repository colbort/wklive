package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TUserSecurityModel = (*customTUserSecurityModel)(nil)

type (
	// TUserSecurityModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTUserSecurityModel.
	TUserSecurityModel interface {
		tUserSecurityModel
	}

	customTUserSecurityModel struct {
		*defaultTUserSecurityModel
	}
)

// NewTUserSecurityModel returns a model for the database table.
func NewTUserSecurityModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TUserSecurityModel {
	return &customTUserSecurityModel{
		defaultTUserSecurityModel: newTUserSecurityModel(conn, c, opts...),
	}
}
