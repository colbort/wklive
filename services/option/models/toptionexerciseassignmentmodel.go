package models

import (
	"context"
	"fmt"

	"wklive/proto/option"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionExerciseAssignmentModel = (*customTOptionExerciseAssignmentModel)(nil)

type (
	// TOptionExerciseAssignmentModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionExerciseAssignmentModel.
	TOptionExerciseAssignmentModel interface {
		tOptionExerciseAssignmentModel
		FindByExercise(ctx context.Context, tenantId, exerciseId int64) ([]*TOptionExerciseAssignment, error)
		ResetForRetry(ctx context.Context, tenantId, exerciseId, now int64) error
		SetPendingStatus(ctx context.Context, tenantId, exerciseId, now int64, status option.ExerciseAssignmentStatus) error
	}

	customTOptionExerciseAssignmentModel struct {
		*defaultTOptionExerciseAssignmentModel
	}
)

func (m *customTOptionExerciseAssignmentModel) FindByExercise(ctx context.Context, tenantId, exerciseId int64) ([]*TOptionExerciseAssignment, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE tenant_id = ? AND exercise_id = ? ORDER BY id",
		tOptionExerciseAssignmentRows, m.table)
	var list []*TOptionExerciseAssignment
	err := m.QueryRowsNoCacheCtx(ctx, &list, query, tenantId, exerciseId)
	return list, err
}

func (m *customTOptionExerciseAssignmentModel) ResetForRetry(
	ctx context.Context,
	tenantId, exerciseId, now int64,
) error {
	_, err := m.ExecNoCacheCtx(ctx, `UPDATE t_option_exercise_assignment
SET status = ?, update_times = ?
WHERE tenant_id = ? AND exercise_id = ? AND status IN (?, ?)`,
		int64(option.ExerciseAssignmentStatus_EXERCISE_ASSIGNMENT_STATUS_PENDING), now,
		tenantId, exerciseId,
		int64(option.ExerciseAssignmentStatus_EXERCISE_ASSIGNMENT_STATUS_FAILED),
		int64(option.ExerciseAssignmentStatus_EXERCISE_ASSIGNMENT_STATUS_MANUAL_REVIEW),
	)
	return err
}

func (m *customTOptionExerciseAssignmentModel) SetPendingStatus(
	ctx context.Context,
	tenantId, exerciseId, now int64,
	status option.ExerciseAssignmentStatus,
) error {
	_, err := m.ExecNoCacheCtx(ctx, `UPDATE t_option_exercise_assignment
SET status = ?, update_times = ?
WHERE tenant_id = ? AND exercise_id = ? AND status <> ?`,
		int64(status), now, tenantId, exerciseId,
		int64(option.ExerciseAssignmentStatus_EXERCISE_ASSIGNMENT_STATUS_DONE),
	)
	return err
}

// NewTOptionExerciseAssignmentModel returns a model for the database table.
func NewTOptionExerciseAssignmentModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionExerciseAssignmentModel {
	return &customTOptionExerciseAssignmentModel{
		defaultTOptionExerciseAssignmentModel: newTOptionExerciseAssignmentModel(conn, c, opts...),
	}
}
