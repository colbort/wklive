package models

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type MenuModel interface {
	sysMenuModel
	FindByIds(ctx context.Context, ids []int64) ([]*SysMenu, error)
	FindIdsByIds(ctx context.Context, ids []int64) ([]int64, error)
	ListAll(ctx context.Context) ([]*SysMenu, error)
	FindPage(ctx context.Context, keyword string, menuType, status, visible, page, pageSize int64) ([]*SysMenu, int64, error)
}

func (m *defaultSysMenuModel) FindByIds(ctx context.Context, ids []int64) ([]*SysMenu, error) {
	var menus []*SysMenu
	query := "select " + sysMenuRows + " from " + m.table + " where id in (?) order by `sort` asc"
	query, args, err := sqlx.In(query, ids)
	if err != nil {
		return nil, err
	}
	err = m.conn.QueryRowsCtx(ctx, &menus, query, args...)
	return menus, err
}

func (m *defaultSysMenuModel) FindIdsByIds(ctx context.Context, ids []int64) ([]int64, error) {
	var existIds []int64
	query := "select id from " + m.table + " where id in (?)"
	query, args, err := sqlx.In(query, ids)
	if err != nil {
		return nil, err
	}
	err = m.conn.QueryRowsCtx(ctx, &existIds, query, args...)
	return existIds, err
}

func (m *defaultSysMenuModel) ListAll(ctx context.Context) ([]*SysMenu, error) {
	var menus []*SysMenu
	query := "select " + sysMenuRows + " from " + m.table + " order by `sort` asc"
	err := m.conn.QueryRowsCtx(ctx, &menus, query)
	return menus, err
}

func (m *defaultSysMenuModel) FindPage(ctx context.Context, keyword string, menuType, status, visible, page, pageSize int64) ([]*SysMenu, int64, error) {

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 10000 {
		pageSize = 10000
	}

	where := "1=1"
	args := make([]any, 0, 4)

	if keyword != "" {
		like := "%" + keyword + "%"
		where += " AND (name LIKE ? OR code LIKE ?)"
		args = append(args, like, like)
	}

	if menuType != 0 {
		where += " AND menu_type = ?"
		args = append(args, menuType)
	}

	if status != 0 {
		where += " AND status = ?"
		args = append(args, status)
	}

	if visible != 0 {
		where += " AND visible = ?"
		args = append(args, visible)
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
LIMIT ? OFFSET ?`, sysMenuRows, m.table, where)

	listArgs := append(args, pageSize, offset)

	var list []*SysMenu
	if err := m.conn.QueryRowsCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
