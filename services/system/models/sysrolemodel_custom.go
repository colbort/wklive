package models

import (
	"context"
	"fmt"
)

type RoleModel interface {
	sysRoleModel
	FindPage(ctx context.Context, keyword string, status int32, page, pageSize int64) ([]*SysRole, int64, error)
}

func (m *defaultSysRoleModel) FindPage(ctx context.Context, keyword string, status int32, page, pageSize int64) ([]*SysRole, int64, error) {

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	where := "1=1"
	args := make([]any, 0, 4)

	if keyword != "" {
		like := "%" + keyword + "%"
		where += " AND (name LIKE ? OR code LIKE ?)"
		args = append(args, like, like)
	}

	if status == 0 || status == 1 {
		where += " AND status = ?"
		args = append(args, status)
	}

	// total
	var total int64
	countSql := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.conn.QueryRowCtx(ctx, &total, countSql, args...); err != nil {
		return nil, 0, err
	}

	// list
	offset := (page - 1) * pageSize
	listSql := fmt.Sprintf(`
SELECT %s
FROM %s
WHERE %s
ORDER BY id DESC
LIMIT ? OFFSET ?`, sysRoleRows, m.table, where)

	listArgs := append(args, pageSize, offset)

	var list []*SysRole
	if err := m.conn.QueryRowsCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
