package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionInboxModel = (*customTOptionInboxModel)(nil)

type (
	// TOptionInboxModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionInboxModel.
	TOptionInboxModel interface {
		tOptionInboxModel
	}

	customTOptionInboxModel struct {
		*defaultTOptionInboxModel
	}
)

// NewTOptionInboxModel returns a model for the database table.
func NewTOptionInboxModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionInboxModel {
	return &customTOptionInboxModel{
		defaultTOptionInboxModel: newTOptionInboxModel(conn, c, opts...),
	}
}
