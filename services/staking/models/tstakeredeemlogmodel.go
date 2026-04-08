package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TStakeRedeemLogModel = (*customTStakeRedeemLogModel)(nil)

type (
	// TStakeRedeemLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTStakeRedeemLogModel.
	TStakeRedeemLogModel interface {
		tStakeRedeemLogModel
	}

	customTStakeRedeemLogModel struct {
		*defaultTStakeRedeemLogModel
	}
)

// NewTStakeRedeemLogModel returns a model for the database table.
func NewTStakeRedeemLogModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TStakeRedeemLogModel {
	return &customTStakeRedeemLogModel{
		defaultTStakeRedeemLogModel: newTStakeRedeemLogModel(conn, c, opts...),
	}
}
