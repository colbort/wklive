package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ SysChatMerchantModel = (*customSysChatMerchantModel)(nil)

type (
	// SysChatMerchantModel is an interface to be customized, add more methods here,
	// and implement the added methods in customSysChatMerchantModel.
	SysChatMerchantModel interface {
		sysChatMerchantModel
	}

	customSysChatMerchantModel struct {
		*defaultSysChatMerchantModel
	}
)

// NewSysChatMerchantModel returns a model for the database table.
func NewSysChatMerchantModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) SysChatMerchantModel {
	return &customSysChatMerchantModel{
		defaultSysChatMerchantModel: newSysChatMerchantModel(conn, c, opts...),
	}
}
