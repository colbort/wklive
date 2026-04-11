package models

import (
	"context"
	"fmt"

	"wklive/common/sqlutil"
)

type BizTradeEventPageFilter struct {
	TenantId    int64
	EventType   string
	BizType     string
	BizId       string
	EventStatus int64
	TimeStart   int64
	TimeEnd     int64
}

type BizTradeEventModel interface {
	tBizTradeEventModel
	FindPage(ctx context.Context, filter BizTradeEventPageFilter, cursor int64, limit int64) ([]*TBizTradeEvent, int64, error)
}

func (m *defaultTBizTradeEventModel) FindPage(ctx context.Context, filter BizTradeEventPageFilter, cursor int64, limit int64) ([]*TBizTradeEvent, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqString("event_type", filter.EventType)
	builder.EqString("biz_type", filter.BizType)
	builder.EqString("biz_id", filter.BizId)
	builder.EqInt64("event_status", filter.EventStatus)
	builder.GteInt64("create_times", filter.TimeStart)
	builder.LteInt64("create_times", filter.TimeEnd)

	where := builder.Where()
	args := builder.Args()

	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSQL, args...); err != nil {
		return nil, 0, err
	}

	listArgs := append([]any{}, args...)
	listSQL := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tBizTradeEventRows, m.table, where)
	if cursor > 0 {
		listSQL += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	listSQL += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)

	var list []*TBizTradeEvent
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSQL, listArgs...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
