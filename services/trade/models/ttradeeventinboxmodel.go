package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TTradeEventInboxModel = (*customTTradeEventInboxModel)(nil)

type (
	// TTradeEventInboxModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTTradeEventInboxModel.
	TTradeEventInboxModel interface {
		tTradeEventInboxModel
	}

	customTTradeEventInboxModel struct {
		*defaultTTradeEventInboxModel
	}
)

// NewTTradeEventInboxModel returns a model for the database table.
func NewTTradeEventInboxModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TTradeEventInboxModel {
	return &customTTradeEventInboxModel{
		defaultTTradeEventInboxModel: newTTradeEventInboxModel(conn, c, opts...),
	}
}
