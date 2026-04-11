package models

import (
	"context"
	"fmt"

	"wklive/common/sqlutil"
)

type TenantPayChannelModel interface {
	tTenantPayChannelModel
	FindPage(ctx context.Context, tenantId int64, platformId int64, productId int64, accountId int64, keyword string, status int64, cursor int64, limit int64) ([]*TTenantPayChannel, int64, error)
}

func (m *defaultTTenantPayChannelModel) FindPage(ctx context.Context, tenantId int64, platformId int64, productId int64, accountId int64, keyword string, status int64, cursor int64, limit int64) ([]*TTenantPayChannel, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", tenantId)
	builder.EqInt64("platform_id", platformId)
	builder.EqInt64("product_id", productId)
	builder.EqInt64("account_id", accountId)
	if keyword != "" {
		builder.And("(channel_code LIKE ? OR channel_name LIKE ? OR display_name LIKE ?)", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	builder.EqInt64("status", status)

	where := builder.Where()
	args := builder.Args()

	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSQL, args...); err != nil {
		return nil, 0, err
	}

	listArgs := append([]any{}, args...)
	listSQL := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tTenantPayChannelRows, m.table, where)
	if cursor > 0 {
		listSQL += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	listSQL += " ORDER BY sort DESC, id DESC LIMIT ?"
	listArgs = append(listArgs, limit)

	var list []*TTenantPayChannel
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSQL, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
