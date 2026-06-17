package models

import (
	"context"
	"fmt"
	"wklive/common/sqlutil"
)

type VerificationCodeRecordPageFilter struct {
	TenantId int64
	Channel  int64
	Target   string
	Scene    int64
	Status   int64
}

type VerificationCodeRecordModel interface {
	sysVerificationCodeRecordModel
	FindPage(ctx context.Context, filter VerificationCodeRecordPageFilter, cursor int64, limit int64) ([]*SysVerificationCodeRecord, int64, error)
}

func (m *defaultSysVerificationCodeRecordModel) FindPage(
	ctx context.Context,
	filter VerificationCodeRecordPageFilter,
	cursor int64,
	limit int64,
) ([]*SysVerificationCodeRecord, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqInt64("channel", filter.Channel)
	builder.LikeString("target", filter.Target)
	builder.EqInt64("scene", filter.Scene)
	builder.EqInt64("status", filter.Status)

	where := builder.Where()
	args := builder.Args()

	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSQL, args...); err != nil {
		return nil, 0, err
	}

	listArgs := append([]any{}, args...)
	var listSQL string
	if cursor <= 0 {
		listSQL = fmt.Sprintf(
			`SELECT %s
			FROM %s
			WHERE %s
			ORDER BY id DESC
			LIMIT ?`,
			sysVerificationCodeRecordRows, m.table, where,
		)
		listArgs = append(listArgs, limit)
	} else {
		listSQL = fmt.Sprintf(
			`SELECT %s
			FROM %s
			WHERE %s AND id < ?
			ORDER BY id DESC
			LIMIT ?`,
			sysVerificationCodeRecordRows, m.table, where,
		)
		listArgs = append(listArgs, cursor, limit)
	}

	var list []*SysVerificationCodeRecord
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSQL, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
