package models

import (
	"context"
	"fmt"

	"wklive/common/sqlutil"
)

type OptionTradePageFilter struct {
	TenantId       int64
	ContractId     int64
	Uid            int64
	AccountId      int64
	TradeNo        string
	TradeTimeStart int64
	TradeTimeEnd   int64
}

type OptionTradeModel interface {
	tOptionTradeModel
	FindPage(ctx context.Context, filter OptionTradePageFilter, cursor int64, limit int64) ([]*TOptionTrade, int64, error)
}

func (m *defaultTOptionTradeModel) FindPage(ctx context.Context, filter OptionTradePageFilter, cursor int64, limit int64) ([]*TOptionTrade, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqInt64("contract_id", filter.ContractId)
	builder.EqString("trade_no", filter.TradeNo)
	builder.GteInt64("trade_time", filter.TradeTimeStart)
	builder.LteInt64("trade_time", filter.TradeTimeEnd)

	where := builder.Where()
	args := builder.Args()
	if filter.Uid != 0 {
		where += " AND (buy_uid = ? OR sell_uid = ?)"
		args = append(args, filter.Uid, filter.Uid)
	}
	if filter.AccountId != 0 {
		where += " AND (buy_account_id = ? OR sell_account_id = ?)"
		args = append(args, filter.AccountId, filter.AccountId)
	}

	var total int64
	countSql := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSql, args...); err != nil {
		return nil, 0, err
	}

	listArgs := append([]any{}, args...)
	listSql := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tOptionTradeRows, m.table, where)
	if cursor > 0 {
		listSql += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	listSql += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)

	var list []*TOptionTrade
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
