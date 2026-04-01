package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TItickQuoteModel = (*customTItickQuoteModel)(nil)

type (
	// TItickQuoteModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTItickQuoteModel.
	TItickQuoteModel interface {
		tItickQuoteModel
	}

	customTItickQuoteModel struct {
		*defaultTItickQuoteModel
	}
)

// NewTItickQuoteModel returns a model for the database table.
func NewTItickQuoteModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TItickQuoteModel {
	return &customTItickQuoteModel{
		defaultTItickQuoteModel: newTItickQuoteModel(conn, c, opts...),
	}
}
