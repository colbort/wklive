package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TAssetCoinConfigModel = (*customTAssetCoinConfigModel)(nil)

type (
	// TAssetCoinConfigModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTAssetCoinConfigModel.
	TAssetCoinConfigModel interface {
		tAssetCoinConfigModel
	}

	customTAssetCoinConfigModel struct {
		*defaultTAssetCoinConfigModel
	}
)

// NewTAssetCoinConfigModel returns a model for the database table.
func NewTAssetCoinConfigModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TAssetCoinConfigModel {
	return &customTAssetCoinConfigModel{
		defaultTAssetCoinConfigModel: newTAssetCoinConfigModel(conn, c, opts...),
	}
}
