package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionExerciseModel = (*customTOptionExerciseModel)(nil)

type (
	// TOptionExerciseModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionExerciseModel.
	TOptionExerciseModel interface {
		tOptionExerciseModel
	}

	customTOptionExerciseModel struct {
		*defaultTOptionExerciseModel
	}
)

// NewTOptionExerciseModel returns a model for the database table.
func NewTOptionExerciseModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionExerciseModel {
	return &customTOptionExerciseModel{
		defaultTOptionExerciseModel: newTOptionExerciseModel(conn, c, opts...),
	}
}
