package models

import (
	"context"
	"fmt"

	"wklive/common/sqlutil"
)

type OptionBillPageFilter struct {
	TenantId        int64
	Uid             int64
	AccountId       int64
	BizNo           string
	RefType         int64
	CreateTimeStart int64
	CreateTimeEnd   int64
}

type OptionBillModel interface {
	tOptionBillModel
	FindPage(ctx context.Context, filter OptionBillPageFilter, cursor int64, limit int64) ([]*TOptionBill, int64, error)
}

func (m *defaultTOptionBillModel) FindPage(ctx context.Context, filter OptionBillPageFilter, cursor int64, limit int64) ([]*TOptionBill, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqInt64("uid", filter.Uid)
	builder.EqInt64("account_id", filter.AccountId)
	builder.EqString("biz_no", filter.BizNo)
	builder.EqInt64("ref_type", filter.RefType)
	builder.GteInt64("create_times", filter.CreateTimeStart)
	builder.LteInt64("create_times", filter.CreateTimeEnd)

	where := builder.Where()
	args := builder.Args()

	var total int64
	countSql := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSql, args...); err != nil {
		return nil, 0, err
	}

	listArgs := append([]any{}, args...)
	listSql := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tOptionBillRows, m.table, where)
	if cursor > 0 {
		listSql += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	listSql += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)

	var list []*TOptionBill
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
