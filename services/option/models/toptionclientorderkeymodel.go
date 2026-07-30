package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionClientOrderKeyModel = (*customTOptionClientOrderKeyModel)(nil)

type (
	// TOptionClientOrderKeyModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionClientOrderKeyModel.
	TOptionClientOrderKeyModel interface {
		tOptionClientOrderKeyModel
	}

	customTOptionClientOrderKeyModel struct {
		*defaultTOptionClientOrderKeyModel
	}
)

// NewTOptionClientOrderKeyModel returns a model for the database table.
func NewTOptionClientOrderKeyModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionClientOrderKeyModel {
	return &customTOptionClientOrderKeyModel{
		defaultTOptionClientOrderKeyModel: newTOptionClientOrderKeyModel(conn, c, opts...),
	}
}
