package models

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ItickSyncTaskModel interface {
	tItickSyncTaskModel
	FindPage(ctx context.Context, cursor, limit int64) ([]*TItickSyncTask, int64, error)
	UpdateStatusByTaskNo(ctx context.Context, taskNo string, status int64, message string, updatedAt int64) error
}

func (m *defaultTItickSyncTaskModel) FindPage(ctx context.Context, cursor, limit int64) ([]*TItickSyncTask, int64, error) {
	query := fmt.Sprintf("select %s from %s where id > ? order by id limit ?", tItickSyncTaskRows, m.table)
	var resp []*TItickSyncTask
	err := m.QueryRowsNoCacheCtx(ctx, &resp, query, cursor, limit)
	if err != nil {
		return nil, 0, err
	}
	var nextCursor int64
	if len(resp) > 0 {
		nextCursor = resp[len(resp)-1].Id
	}
	return resp, nextCursor, nil
}

func (m *defaultTItickSyncTaskModel) UpdateStatusByTaskNo(ctx context.Context, taskNo string, status int64, message string, updatedAt int64) error {
	query := fmt.Sprintf("update %s set status = ?, message = ?, updated_at = ? where task_no = ?", m.table)
	_, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		return conn.ExecCtx(ctx, query, status, message, updatedAt, taskNo)
	})
	return err
}
