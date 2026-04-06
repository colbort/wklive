package models

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type UserIdentityModel interface {
	tUserIdentityModel
	FindPage(ctx context.Context, tenantId int64, cursor int64, limit int64) ([]*TUserIdentity, int64, error)
	FindByEmail(ctx context.Context, tenantId int64, email string) (*TUserIdentity, error)
	FindByPhone(ctx context.Context, tenantId int64, phone string) (*TUserIdentity, error)
}

func (m *defaultTUserIdentityModel) FindPage(ctx context.Context, tenantId int64, cursor int64, limit int64) ([]*TUserIdentity, int64, error) {
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
			tUserIdentityRows, m.table, where,
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
			tUserIdentityRows, m.table, where,
		)
		listArgs = append(listArgs, cursor, limit)
	}

	var list []*TUserIdentity
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (m *defaultTUserIdentityModel) FindByEmail(ctx context.Context, tenantId int64, email string) (*TUserIdentity, error) {
	var resp TUserIdentity

	query := fmt.Sprintf(`
		SELECT %s 
		FROM %s 
		WHERE tenant_id = ? 
		AND email = ? 
		LIMIT 1
	`, tUserIdentityRows, m.table)

	err := m.QueryRowNoCacheCtx(ctx, &resp, query, tenantId, email)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &resp, nil
}

func (m *defaultTUserIdentityModel) FindByPhone(ctx context.Context, tenantId int64, phone string) (*TUserIdentity, error) {
	var resp TUserIdentity

	query := fmt.Sprintf(`
		SELECT %s 
		FROM %s 
		WHERE tenant_id = ? 
		AND phone = ? 
		LIMIT 1
	`, tUserIdentityRows, m.table)

	err := m.QueryRowNoCacheCtx(ctx, &resp, query, tenantId, phone)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &resp, nil
}
