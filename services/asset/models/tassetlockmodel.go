package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TAssetLockModel = (*customTAssetLockModel)(nil)

type (
	// TAssetLockModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTAssetLockModel.
	TAssetLockModel interface {
		tAssetLockModel
	}

	customTAssetLockModel struct {
		*defaultTAssetLockModel
	}
)

// NewTAssetLockModel returns a model for the database table.
func NewTAssetLockModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TAssetLockModel {
	return &customTAssetLockModel{
		defaultTAssetLockModel: newTAssetLockModel(conn, c, opts...),
	}
}
