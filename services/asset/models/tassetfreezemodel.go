package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TAssetFreezeModel = (*customTAssetFreezeModel)(nil)

type (
	// TAssetFreezeModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTAssetFreezeModel.
	TAssetFreezeModel interface {
		tAssetFreezeModel
	}

	customTAssetFreezeModel struct {
		*defaultTAssetFreezeModel
	}
)

// NewTAssetFreezeModel returns a model for the database table.
func NewTAssetFreezeModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TAssetFreezeModel {
	return &customTAssetFreezeModel{
		defaultTAssetFreezeModel: newTAssetFreezeModel(conn, c, opts...),
	}
}
