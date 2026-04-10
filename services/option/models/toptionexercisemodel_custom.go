package models

import (
	"context"
	"fmt"

	"wklive/common/sqlutil"
)

type OptionExercisePageFilter struct {
	TenantId          int64
	Uid               int64
	AccountId         int64
	ContractId        int64
	ExerciseType      int64
	Status            int64
	ExerciseTimeStart int64
	ExerciseTimeEnd   int64
}

type OptionExerciseModel interface {
	tOptionExerciseModel
	FindPage(ctx context.Context, filter OptionExercisePageFilter, cursor int64, limit int64) ([]*TOptionExercise, int64, error)
}

func (m *defaultTOptionExerciseModel) FindPage(ctx context.Context, filter OptionExercisePageFilter, cursor int64, limit int64) ([]*TOptionExercise, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqInt64("uid", filter.Uid)
	builder.EqInt64("account_id", filter.AccountId)
	builder.EqInt64("contract_id", filter.ContractId)
	builder.EqInt64("exercise_type", filter.ExerciseType)
	builder.EqInt64("status", filter.Status)
	builder.GteInt64("exercise_time", filter.ExerciseTimeStart)
	builder.LteInt64("exercise_time", filter.ExerciseTimeEnd)

	where := builder.Where()
	args := builder.Args()

	var total int64
	countSql := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSql, args...); err != nil {
		return nil, 0, err
	}

	listArgs := append([]any{}, args...)
	listSql := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tOptionExerciseRows, m.table, where)
	if cursor > 0 {
		listSql += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	listSql += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)

	var list []*TOptionExercise
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
