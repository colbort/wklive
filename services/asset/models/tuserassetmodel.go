package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TUserAssetModel = (*customTUserAssetModel)(nil)

type (
	// TUserAssetModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTUserAssetModel.
	TUserAssetModel interface {
		tUserAssetModel
	}

	customTUserAssetModel struct {
		*defaultTUserAssetModel
	}
)

// NewTUserAssetModel returns a model for the database table.
func NewTUserAssetModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TUserAssetModel {
	return &customTUserAssetModel{
		defaultTUserAssetModel: newTUserAssetModel(conn, c, opts...),
	}
}
