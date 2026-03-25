package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TUserIdentityModel = (*customTUserIdentityModel)(nil)

type (
	// TUserIdentityModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTUserIdentityModel.
	TUserIdentityModel interface {
		tUserIdentityModel
	}

	customTUserIdentityModel struct {
		*defaultTUserIdentityModel
	}
)

// NewTUserIdentityModel returns a model for the database table.
func NewTUserIdentityModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TUserIdentityModel {
	return &customTUserIdentityModel{
		defaultTUserIdentityModel: newTUserIdentityModel(conn, c, opts...),
	}
}
