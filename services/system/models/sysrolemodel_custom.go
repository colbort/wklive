package models

import (
	"context"
	"fmt"
)

type RoleModel interface {
	sysRoleModel
	FindPage(ctx context.Context, keyword string, status, cursor, pageSize int64) ([]*SysRole, int64, error)
}

func (m *defaultSysRoleModel) FindPage(
	ctx context.Context,
	keyword string,
	status int64,
	cursor, pageSize int64,
) ([]*SysRole, int64, error) {

	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
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
	if status >= 0 {
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
		listArgs = append(listArgs, pageSize)
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
		listArgs = append(listArgs, cursor, pageSize)
	}

	var list []*SysRole
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
