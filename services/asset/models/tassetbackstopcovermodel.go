package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TAssetBackstopCoverModel = (*customTAssetBackstopCoverModel)(nil)

type (
	// TAssetBackstopCoverModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTAssetBackstopCoverModel.
	TAssetBackstopCoverModel interface {
		tAssetBackstopCoverModel
	}

	customTAssetBackstopCoverModel struct {
		*defaultTAssetBackstopCoverModel
	}
)

// NewTAssetBackstopCoverModel returns a model for the database table.
func NewTAssetBackstopCoverModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TAssetBackstopCoverModel {
	return &customTAssetBackstopCoverModel{
		defaultTAssetBackstopCoverModel: newTAssetBackstopCoverModel(conn, c, opts...),
	}
}
