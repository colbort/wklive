package models

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	g "github.com/zeromicro/go-zero/core/stores/sqlx"
)

type RoleMenuModel interface {
	sysRoleMenuModel
	FindMenuIdsByRoleIds(ctx context.Context, roleIds []int64) ([]int64, error)
	DeleteByRoleId(ctx context.Context, roleId int64) error
	InsertBatch(ctx context.Context, data []*SysRoleMenu) error
	TransactCtx(ctx context.Context, fn func(context.Context, g.Session) error) error
	ListByRoleId(ctx context.Context, roleId int64) ([]*SysRoleMenu, error)
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

func (m *defaultSysRoleMenuModel) DeleteByRoleId(ctx context.Context, roleId int64) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE role_id = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, roleId)
	return err
}

func (m *defaultSysRoleMenuModel) InsertBatch(ctx context.Context, data []*SysRoleMenu) error {
	if len(data) == 0 {
		return nil
	}

	valueStrings := make([]string, 0, len(data))
	valueArgs := make([]interface{}, 0, len(data)*2)

	for _, d := range data {
		valueStrings = append(valueStrings, "(?, ?)")
		valueArgs = append(valueArgs, d.RoleId, d.MenuId)
	}

	stmt := fmt.Sprintf(
		"INSERT INTO %s (role_id, menu_id) VALUES %s",
		m.table,
		strings.Join(valueStrings, ","),
	)

	_, err := m.conn.ExecCtx(ctx, stmt, valueArgs...)
	return err
}

func (m *defaultSysRoleMenuModel) TransactCtx(ctx context.Context, fn func(context.Context, g.Session) error) error {
	return m.conn.TransactCtx(ctx, fn)
}

func (m *defaultSysRoleMenuModel) ListByRoleId(ctx context.Context, roleId int64) ([]*SysRoleMenu, error) {
	var list []*SysRoleMenu
	query := fmt.Sprintf("SELECT %s FROM %s WHERE role_id = ?", sysRoleMenuRows, m.table)
	err := m.conn.QueryRowsCtx(ctx, &list, query, roleId)
	return list, err
}
