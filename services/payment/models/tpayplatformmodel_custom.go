package models

import (
	"context"
	"fmt"

	"wklive/common/sqlutil"
)

type PayPlatformModel interface {
	tPayPlatformModel
	FindPage(ctx context.Context, keyword string, platformCode string, platformType int64, status int64, cursor int64, limit int64) ([]*TPayPlatform, int64, error)
}

func (m *defaultTPayPlatformModel) FindPage(ctx context.Context, keyword string, platformCode string, platformType int64, status int64, cursor int64, limit int64) ([]*TPayPlatform, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	if keyword != "" {
		builder.And("(platform_code LIKE ? OR platform_name LIKE ?)", "%"+keyword+"%", "%"+keyword+"%")
	}
	builder.EqString("platform_code", platformCode)
	builder.EqInt64("platform_type", platformType)
	builder.EqInt64("status", status)

	where := builder.Where()
	args := builder.Args()

	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSQL, args...); err != nil {
		return nil, 0, err
	}

	listArgs := append([]any{}, args...)
	listSQL := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tPayPlatformRows, m.table, where)
	if cursor > 0 {
		listSQL += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	listSQL += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)

	var list []*TPayPlatform
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSQL, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
