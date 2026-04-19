package models

import (
	"context"
	"database/sql"
	"fmt"
	"wklive/common/sqlutil"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ItickSyncTaskModel interface {
	tItickSyncTaskModel
	FindPage(ctx context.Context, cursor, limit int64) ([]*TItickSyncTask, int64, error)
	UpdateStatusByTaskNo(ctx context.Context, taskNo string, status int64, message string, updatedAt int64) error
}

func (m *defaultTItickSyncTaskModel) FindPage(ctx context.Context, cursor, limit int64) ([]*TItickSyncTask, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()

	where := builder.Where()
	args := builder.Args()

	// ---- total ----
	var total int64
	countSql := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSql, args...); err != nil {
		return nil, 0, err
	}

	// ---- list ----
	listArgs := append([]any{}, args...)
	var listSql string

	if cursor <= 0 {
		// 第一页
		listSql = fmt.Sprintf(
			`SELECT %s
			FROM %s
			WHERE %s
			ORDER BY id DESC
			LIMIT ?`,
			tItickSyncTaskRows, m.table, where,
		)
		listArgs = append(listArgs, limit)
	} else {
		// 后续页
		listSql = fmt.Sprintf(
			`SELECT %s
			FROM %s
			WHERE %s AND id < ?
			ORDER BY id DESC
			LIMIT ?`,
			tItickSyncTaskRows, m.table, where,
		)
		listArgs = append(listArgs, cursor, limit)
	}

	var list []*TItickSyncTask
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (m *defaultTItickSyncTaskModel) UpdateStatusByTaskNo(ctx context.Context, taskNo string, status int64, message string, updatedAt int64) error {
	query := fmt.Sprintf("update %s set status = ?, message = ?, update_times = ? where task_no = ?", m.table)
	_, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		return conn.ExecCtx(ctx, query, status, message, updatedAt, taskNo)
	})
	return err
}
