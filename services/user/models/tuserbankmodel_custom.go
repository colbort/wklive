package models

import (
	"context"
	"fmt"
)

type UserBankModel interface {
	tUserBankModel
	FindPage(ctx context.Context, tenantId int64, userId int64, cursor int64, limit int64) ([]*TUserBank, int64, error)
}

func (m *defaultTUserBankModel) FindPage(ctx context.Context, tenantId int64, userId int64, cursor int64, limit int64) ([]*TUserBank, int64, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	where := "1=1"
	args := make([]any, 0, 2)

	if tenantId != 0 {
		where += " AND tenant_id = ?"
		args = append(args, tenantId)
	}

	if userId != 0 {
		where += " AND user_id = ?"
		args = append(args, userId)
	}

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
			tUserBankRows, m.table, where,
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
			tUserBankRows, m.table, where,
		)
		listArgs = append(listArgs, cursor, limit)
	}

	var list []*TUserBank
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
