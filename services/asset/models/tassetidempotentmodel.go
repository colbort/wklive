package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TAssetIdempotentModel = (*customTAssetIdempotentModel)(nil)

type (
	// TAssetIdempotentModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTAssetIdempotentModel.
	TAssetIdempotentModel interface {
		tAssetIdempotentModel
	}

	customTAssetIdempotentModel struct {
		*defaultTAssetIdempotentModel
	}
)

// NewTAssetIdempotentModel returns a model for the database table.
func NewTAssetIdempotentModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TAssetIdempotentModel {
	return &customTAssetIdempotentModel{
		defaultTAssetIdempotentModel: newTAssetIdempotentModel(conn, c, opts...),
	}
}
