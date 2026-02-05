package models

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type MenuModel interface {
	sysMenuModel
	FindByIds(ctx context.Context, ids []int64) ([]*SysMenu, error)
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
