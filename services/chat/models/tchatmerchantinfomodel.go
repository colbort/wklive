package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TChatMerchantInfoModel = (*customTChatMerchantInfoModel)(nil)

type (
	// TChatMerchantInfoModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTChatMerchantInfoModel.
	TChatMerchantInfoModel interface {
		tChatMerchantInfoModel
	}

	customTChatMerchantInfoModel struct {
		*defaultTChatMerchantInfoModel
	}
)

// NewTChatMerchantInfoModel returns a model for the database table.
func NewTChatMerchantInfoModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TChatMerchantInfoModel {
	return &customTChatMerchantInfoModel{
		defaultTChatMerchantInfoModel: newTChatMerchantInfoModel(conn, c, opts...),
	}
}
