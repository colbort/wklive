package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TAssetFlowModel = (*customTAssetFlowModel)(nil)

type (
	// TAssetFlowModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTAssetFlowModel.
	TAssetFlowModel interface {
		tAssetFlowModel
	}

	customTAssetFlowModel struct {
		*defaultTAssetFlowModel
	}
)

// NewTAssetFlowModel returns a model for the database table.
func NewTAssetFlowModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TAssetFlowModel {
	return &customTAssetFlowModel{
		defaultTAssetFlowModel: newTAssetFlowModel(conn, c, opts...),
	}
}
