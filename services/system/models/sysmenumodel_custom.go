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
	FindPage(ctx context.Context, keyword string, menuType, status, visible, cursor, limit int64) ([]*SysMenu, int64, error)
}

func (m *defaultSysMenuModel) FindByIds(ctx context.Context, ids []int64) ([]*SysMenu, error) {
	var menus []*SysMenu
	query := "select " + sysMenuRows + " from " + m.table + " where id in (?) order by `sort` asc"
	query, args, err := sqlx.In(query, ids)
	if err != nil {
		return nil, err
	}
	err = m.QueryRowsNoCacheCtx(ctx, &menus, query, args...)
	return menus, err
}

func (m *defaultSysMenuModel) FindIdsByIds(ctx context.Context, ids []int64) ([]int64, error) {
	var existIds []int64
	query := "select id from " + m.table + " where id in (?)"
	query, args, err := sqlx.In(query, ids)
	if err != nil {
		return nil, err
	}
	err = m.QueryRowsNoCacheCtx(ctx, &existIds, query, args...)
	return existIds, err
}

func (m *defaultSysMenuModel) ListAll(ctx context.Context) ([]*SysMenu, error) {
	var menus []*SysMenu
	query := "select " + sysMenuRows + " from " + m.table + " order by `sort` asc"
	err := m.QueryRowsNoCacheCtx(ctx, &menus, query)
	return menus, err
}

func (m *defaultSysMenuModel) FindPage(
	ctx context.Context,
	keyword string,
	menuType, status, visible int64,
	cursor, limit int64,
) ([]*SysMenu, int64, error) {

	if limit <= 0 {
		limit = 100
	}
	if limit > 10000 {
		limit = 10000
	}

	where := "1=1"
	args := make([]any, 0, 6)

	if keyword != "" {
		like := "%" + keyword + "%"
		where += " AND (name LIKE ? OR code LIKE ?)"
		args = append(args, like, like)
	}

	// 假设 0 表示全部，不筛选
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
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSql, args...); err != nil {
		return nil, 0, err
	}

	// list
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
			sysMenuRows, m.table, where,
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
			sysMenuRows, m.table, where,
		)
		listArgs = append(listArgs, cursor, limit)
	}

	var list []*SysMenu
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
