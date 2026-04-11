package models

import (
	"context"
	"fmt"

	"wklive/common/sqlutil"
)

type StakeProductModel interface {
	tStakeProductModel
	FindPage(ctx context.Context, tenantID int64, cursor, limit int64, productNo, productName, coinSymbol string, productType, status int64) ([]*TStakeProduct, int64, error)
}

func (m *defaultTStakeProductModel) FindPage(ctx context.Context, tenantID int64, cursor, limit int64, productNo, productName, coinSymbol string, productType, status int64) ([]*TStakeProduct, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	builder.And("tenant_id = ?", tenantID)
	builder.EqString("product_no", productNo)
	if productName != "" {
		builder.LikeString("product_name", "%"+productName+"%")
	}
	builder.EqString("coin_symbol", coinSymbol)
	builder.EqInt64("product_type", productType)
	builder.EqInt64("status", status)

	where := builder.Where()
	args := builder.Args()

	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSQL, args...); err != nil {
		return nil, 0, err
	}

	listArgs := append([]any{}, args...)
	listSQL := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tStakeProductRows, m.table, where)
	if cursor > 0 {
		listSQL += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	listSQL += " ORDER BY sort DESC, id DESC LIMIT ?"
	listArgs = append(listArgs, limit)

	var list []*TStakeProduct
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSQL, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
