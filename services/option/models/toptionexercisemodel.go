package models

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"wklive/common/sqlutil"
	"wklive/proto/option"
)

var _ TOptionExerciseModel = (*customTOptionExerciseModel)(nil)

type (
	OptionExercisePageFilter struct {
		TenantId          int64
		UserId            int64
		AccountId         int64
		ContractId        int64
		PositionId        int64
		ExerciseType      int64
		Status            int64
		ExerciseTimeStart int64
		ExerciseTimeEnd   int64
	}

	// TOptionExerciseModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionExerciseModel.
	TOptionExerciseModel interface {
		tOptionExerciseModel
		FindPage(ctx context.Context, filter OptionExercisePageFilter, cursor int64, limit int64) ([]*TOptionExercise, int64, error)
		FindPending(ctx context.Context, tenantId int64, limit int64) ([]*TOptionExercise, error)
		HasPendingByContract(ctx context.Context, tenantId, contractId int64) (bool, error)
		FindOneForUpdate(ctx context.Context, id int64) (*TOptionExercise, error)
	}

	customTOptionExerciseModel struct {
		*defaultTOptionExerciseModel
	}
)

func (m *defaultTOptionExerciseModel) FindPending(ctx context.Context, tenantId int64, limit int64) ([]*TOptionExercise, error) {
	limit = sqlutil.NormalizeLimit(limit)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE status = ?", tOptionExerciseRows, m.table)
	args := []any{int64(option.ExerciseStatus_EXERCISE_STATUS_PENDING)}
	if tenantId > 0 {
		query += " AND tenant_id = ?"
		args = append(args, tenantId)
	}
	query += " ORDER BY id LIMIT ?"
	args = append(args, limit)
	var list []*TOptionExercise
	err := m.QueryRowsNoCacheCtx(ctx, &list, query, args...)
	return list, err
}

func (m *defaultTOptionExerciseModel) FindOneForUpdate(ctx context.Context, id int64) (*TOptionExercise, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE id = ? LIMIT 1 FOR UPDATE", tOptionExerciseRows, m.table)
	var item TOptionExercise
	err := m.QueryRowNoCacheCtx(ctx, &item, query, id)
	return &item, err
}

func (m *defaultTOptionExerciseModel) HasPendingByContract(ctx context.Context, tenantId, contractId int64) (bool, error) {
	query := fmt.Sprintf(
		"SELECT COUNT(1) FROM %s WHERE tenant_id = ? AND contract_id = ? AND status = ?",
		m.table,
	)
	var count int64
	err := m.QueryRowNoCacheCtx(
		ctx,
		&count,
		query,
		tenantId,
		contractId,
		int64(option.ExerciseStatus_EXERCISE_STATUS_PENDING),
	)
	return count > 0, err
}

// NewTOptionExerciseModel returns a model for the database table.
func NewTOptionExerciseModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionExerciseModel {
	return &customTOptionExerciseModel{
		defaultTOptionExerciseModel: newTOptionExerciseModel(conn, c, opts...),
	}
}

func (m *defaultTOptionExerciseModel) FindPage(ctx context.Context, filter OptionExercisePageFilter, cursor int64, limit int64) ([]*TOptionExercise, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqInt64("user_id", filter.UserId)
	builder.EqInt64("account_id", filter.AccountId)
	builder.EqInt64("contract_id", filter.ContractId)
	builder.EqInt64("position_id", filter.PositionId)
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
