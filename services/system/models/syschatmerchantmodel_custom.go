package models

import (
	"context"
	"fmt"
	"wklive/common/sqlutil"
)

type ChatMerchantPageFilter struct {
	Keyword      string
	Enabled      int64
	MerchantName string
	MerchantCode string
	ContactName  string
	ContactPhone string
	ContactEmail string
}

type ChatMerchantModel interface {
	sysChatMerchantModel
	FindPage(ctx context.Context, filter ChatMerchantPageFilter, cursor int64, limit int64) ([]*SysChatMerchant, int64, error)
}

func (m *customSysChatMerchantModel) FindPage(
	ctx context.Context,
	filter ChatMerchantPageFilter,
	cursor int64,
	limit int64,
) ([]*SysChatMerchant, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	if filter.Keyword != "" {
		like := "%" + filter.Keyword + "%"
		builder.And(
			"(merchant_name LIKE ? OR merchant_code LIKE ? OR contact_name LIKE ? OR contact_phone LIKE ? OR contact_email LIKE ?)",
			like, like, like, like, like,
		)
	}
	builder.EqInt64("enabled", filter.Enabled)
	builder.EqString("merchant_name", filter.MerchantName)
	builder.EqString("merchant_code", filter.MerchantCode)
	builder.EqString("contact_name", filter.ContactName)
	builder.EqString("contact_phone", filter.ContactPhone)
	builder.EqString("contact_email", filter.ContactEmail)

	where := builder.Where()
	args := builder.Args()

	var total int64
	countSql := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSql, args...); err != nil {
		return nil, 0, err
	}

	listSql := fmt.Sprintf(
		"SELECT %s FROM %s WHERE %s ORDER BY id DESC LIMIT ?,?",
		sysChatMerchantRows,
		m.table,
		where,
	)
	listArgs := append(append([]any{}, args...), cursor, limit)

	var list []*SysChatMerchant
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
