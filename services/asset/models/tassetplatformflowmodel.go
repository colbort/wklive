package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TAssetPlatformFlowModel = (*customTAssetPlatformFlowModel)(nil)

type (
	// TAssetPlatformFlowModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTAssetPlatformFlowModel.
	TAssetPlatformFlowModel interface {
		tAssetPlatformFlowModel
	}

	customTAssetPlatformFlowModel struct {
		*defaultTAssetPlatformFlowModel
	}
)

// NewTAssetPlatformFlowModel returns a model for the database table.
func NewTAssetPlatformFlowModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TAssetPlatformFlowModel {
	return &customTAssetPlatformFlowModel{
		defaultTAssetPlatformFlowModel: newTAssetPlatformFlowModel(conn, c, opts...),
	}
}
