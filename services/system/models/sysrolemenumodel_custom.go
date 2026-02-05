package models

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type RoleMenuModel interface {
	sysRoleMenuModel
	FindMenuIdsByRoleIds(ctx context.Context, roleIds []int64) ([]int64, error)
}

func (m *defaultSysRoleMenuModel) FindMenuIdsByRoleIds(ctx context.Context, roleIds []int64) ([]int64, error) {
	var ids []int64
	query := "select menu_id from " + m.table + " where role_id in (?)"
	query, args, err := sqlx.In(query, roleIds)
	if err != nil {
		return nil, err
	}
	err = m.conn.QueryRowsCtx(ctx, &ids, query, args...)
	return ids, err
}
