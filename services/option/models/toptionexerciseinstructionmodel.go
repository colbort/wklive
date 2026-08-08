package models

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionExerciseInstructionModel = (*customTOptionExerciseInstructionModel)(nil)

type (
	// TOptionExerciseInstructionModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionExerciseInstructionModel.
	TOptionExerciseInstructionModel interface {
		tOptionExerciseInstructionModel
		FindLatestByPosition(ctx context.Context, tenantId, positionId int64) (*TOptionExerciseInstruction, error)
		FindLatestByPositionForUpdate(ctx context.Context, tenantId, positionId int64) (*TOptionExerciseInstruction, error)
	}

	customTOptionExerciseInstructionModel struct {
		*defaultTOptionExerciseInstructionModel
	}
)

func (m *customTOptionExerciseInstructionModel) FindLatestByPosition(
	ctx context.Context,
	tenantId, positionId int64,
) (*TOptionExerciseInstruction, error) {
	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE tenant_id = ? AND position_id = ? ORDER BY version DESC LIMIT 1",
		tOptionExerciseInstructionRows, m.table,
	)
	var item TOptionExerciseInstruction
	if err := m.QueryRowNoCacheCtx(ctx, &item, query, tenantId, positionId); err != nil {
		return nil, err
	}
	return &item, nil
}

func (m *customTOptionExerciseInstructionModel) FindLatestByPositionForUpdate(
	ctx context.Context,
	tenantId, positionId int64,
) (*TOptionExerciseInstruction, error) {
	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE tenant_id = ? AND position_id = ? ORDER BY version DESC LIMIT 1 FOR UPDATE",
		tOptionExerciseInstructionRows, m.table,
	)
	var item TOptionExerciseInstruction
	if err := m.QueryRowNoCacheCtx(ctx, &item, query, tenantId, positionId); err != nil {
		return nil, err
	}
	return &item, nil
}

// NewTOptionExerciseInstructionModel returns a model for the database table.
func NewTOptionExerciseInstructionModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionExerciseInstructionModel {
	return &customTOptionExerciseInstructionModel{
		defaultTOptionExerciseInstructionModel: newTOptionExerciseInstructionModel(conn, c, opts...),
	}
}
