package models

import (
	"context"
	"fmt"

	"wklive/common/sqlutil"
)

type PayPlatformPageFilter struct {
	Keyword      string
	PlatformCode string
	PlatformType int64
	Enabled      int64
}

type PayPlatformModel interface {
	tPayPlatformModel
	FindPage(ctx context.Context, filter PayPlatformPageFilter, cursor int64, limit int64) ([]*TPayPlatform, int64, error)
}

func (m *defaultTPayPlatformModel) FindPage(ctx context.Context, filter PayPlatformPageFilter, cursor int64, limit int64) ([]*TPayPlatform, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	if filter.Keyword != "" {
		builder.And("(platform_code LIKE ? OR platform_name LIKE ?)", "%"+filter.Keyword+"%", "%"+filter.Keyword+"%")
	}
	builder.EqString("platform_code", filter.PlatformCode)
	builder.EqInt64("platform_type", filter.PlatformType)
	builder.EqInt64("enabled", filter.Enabled)

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
