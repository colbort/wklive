package models

import (
	"context"
	"fmt"
	"strings"
	"wklive/common/sqlutil"
)

type VerificationCodeRecordModel interface {
	sysVerificationCodeRecordModel
	FindPage(ctx context.Context, tenantId, channel int64, target string, scene, status, cursor, limit int64) ([]*SysVerificationCodeRecord, int64, error)
}

func (m *defaultSysVerificationCodeRecordModel) FindPage(
	ctx context.Context,
	tenantId, channel int64,
	target string,
	scene, status, cursor, limit int64,
) ([]*SysVerificationCodeRecord, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", tenantId)
	builder.EqInt64("channel", channel)
	builder.LikeString("target", "%"+strings.TrimSpace(target)+"%")
	builder.EqInt64("scene", scene)
	builder.EqInt64("status", status)

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
