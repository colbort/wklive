package models

import (
	"context"
	"fmt"
	"strings"
)

type RoleModel interface {
	sysRoleModel
	FindPage(ctx context.Context, keyword string, status, cursor, limit int64) ([]*SysRole, int64, error)
	FindIdsByIds(ctx context.Context, ids []int64) ([]int64, error)
}

func (m *defaultSysRoleModel) FindPage(
	ctx context.Context,
	keyword string,
	status int64,
	cursor, limit int64,
) ([]*SysRole, int64, error) {

	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	// ---- WHERE 条件 ----
	where := "1=1"
	args := make([]any, 0, 4)

	if keyword != "" {
		like := "%" + keyword + "%"
		where += " AND (name LIKE ? OR code LIKE ?)"
		args = append(args, like, like)
	}

	// 假设 status < 0 表示全部
	if status > 0 {
		where += " AND status = ?"
		args = append(args, status)
	}

	// ---- total ----
	var total int64
	countSql := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSql, args...); err != nil {
		return nil, 0, err
	}

	// ---- list ----
	var listSql string
	listArgs := append([]any{}, args...)

	if cursor <= 0 {
		// 第一页
		listSql = fmt.Sprintf(
			`SELECT %s
			FROM %s
			WHERE %s
			ORDER BY id DESC
			LIMIT ?`,
			sysRoleRows, m.table, where,
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
			sysRoleRows, m.table, where,
		)
		listArgs = append(listArgs, cursor, limit)
	}

	var list []*SysRole
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (m *defaultSysRoleModel) FindIdsByIds(ctx context.Context, ids []int64) ([]int64, error) {
	if len(ids) == 0 {
		return []int64{}, nil
	}

	placeholders := make([]string, 0, len(ids))
	args := make([]any, 0, len(ids))
	for _, id := range ids {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}

	query := fmt.Sprintf(
		"SELECT id FROM %s WHERE id IN (%s)",
		m.table,
		strings.Join(placeholders, ","),
	)

	var existIds []int64
	err := m.QueryRowsNoCacheCtx(ctx, &existIds, query, args...)
	if err != nil {
		return nil, err
	}
	return existIds, nil
}
