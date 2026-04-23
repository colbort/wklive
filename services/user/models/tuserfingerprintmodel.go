package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TUserFingerprintModel = (*customTUserFingerprintModel)(nil)

type (
	// TUserFingerprintModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTUserFingerprintModel.
	TUserFingerprintModel interface {
		tUserFingerprintModel
	}

	customTUserFingerprintModel struct {
		*defaultTUserFingerprintModel
	}
)

// NewTUserFingerprintModel returns a model for the database table.
func NewTUserFingerprintModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TUserFingerprintModel {
	return &customTUserFingerprintModel{
		defaultTUserFingerprintModel: newTUserFingerprintModel(conn, c, opts...),
	}
}
