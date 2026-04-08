package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TStakeRewardLogModel = (*customTStakeRewardLogModel)(nil)

type (
	// TStakeRewardLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTStakeRewardLogModel.
	TStakeRewardLogModel interface {
		tStakeRewardLogModel
	}

	customTStakeRewardLogModel struct {
		*defaultTStakeRewardLogModel
	}
)

// NewTStakeRewardLogModel returns a model for the database table.
func NewTStakeRewardLogModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TStakeRewardLogModel {
	return &customTStakeRewardLogModel{
		defaultTStakeRewardLogModel: newTStakeRewardLogModel(conn, c, opts...),
	}
}
