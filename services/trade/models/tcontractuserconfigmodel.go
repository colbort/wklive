package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TContractUserConfigModel = (*customTContractUserConfigModel)(nil)

type (
	// TContractUserConfigModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTContractUserConfigModel.
	TContractUserConfigModel interface {
		tContractUserConfigModel
	}

	customTContractUserConfigModel struct {
		*defaultTContractUserConfigModel
	}
)

// NewTContractUserConfigModel returns a model for the database table.
func NewTContractUserConfigModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TContractUserConfigModel {
	return &customTContractUserConfigModel{
		defaultTContractUserConfigModel: newTContractUserConfigModel(conn, c, opts...),
	}
}
