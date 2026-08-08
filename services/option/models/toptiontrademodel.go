package models

import (
	"context"
	"fmt"
	"strings"

	"wklive/common/sqlutil"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionTradeModel = (*customTOptionTradeModel)(nil)

type (
	OptionTradePageFilter struct {
		TenantId       int64
		ContractId     int64
		UserId         int64
		AccountId      int64
		TradeNo        string
		TradeTimeStart int64
		TradeTimeEnd   int64
	}

	OptionTradeStatistics struct {
		ContractId int64           `db:"contract_id"`
		Volume     decimal.Decimal `db:"volume"`
		Turnover   decimal.Decimal `db:"turnover"`
		TradeCount int64           `db:"trade_count"`
		LastTrade  int64           `db:"last_trade"`
	}

	// TOptionTradeModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionTradeModel.
	TOptionTradeModel interface {
		tOptionTradeModel
		FindPage(ctx context.Context, filter OptionTradePageFilter, cursor int64, limit int64) ([]*TOptionTrade, int64, error)
		FindStatisticsByContracts(ctx context.Context, tenantId int64, contractIDs []int64, startTime, endTime int64) ([]*OptionTradeStatistics, error)
		FindLastMatchSequence(ctx context.Context, tenantId, contractId int64) (int64, error)
		FindByComboOrderID(ctx context.Context, tenantId, comboOrderId, limit int64) ([]*TOptionTrade, int64, error)
	}

	customTOptionTradeModel struct {
		*defaultTOptionTradeModel
	}
)

func (m *customTOptionTradeModel) FindStatisticsByContracts(
	ctx context.Context, tenantId int64, contractIDs []int64, startTime, endTime int64,
) ([]*OptionTradeStatistics, error) {
	if len(contractIDs) == 0 {
		return []*OptionTradeStatistics{}, nil
	}
	args := make([]any, 0, len(contractIDs)+3)
	args = append(args, tenantId, startTime, endTime)
	for _, id := range contractIDs {
		args = append(args, id)
	}
	placeholders := strings.TrimRight(strings.Repeat("?,", len(contractIDs)), ",")
	query := fmt.Sprintf(`SELECT contract_id,
  COALESCE(SUM(qty),0) AS volume,
  COALESCE(SUM(turnover),0) AS turnover,
  COUNT(1) AS trade_count,
  COALESCE(MAX(trade_time),0) AS last_trade
FROM %s
WHERE tenant_id=? AND trade_time>=? AND trade_time<=? AND contract_id IN (%s)
GROUP BY contract_id`, m.table, placeholders)
	var items []*OptionTradeStatistics
	err := m.QueryRowsNoCacheCtx(ctx, &items, query, args...)
	return items, err
}

func (m *customTOptionTradeModel) FindLastMatchSequence(
	ctx context.Context, tenantId, contractId int64,
) (int64, error) {
	query := fmt.Sprintf(
		"SELECT COALESCE(MAX(match_sequence),0) FROM %s WHERE tenant_id=? AND contract_id=?",
		m.table,
	)
	var sequence int64
	err := m.QueryRowNoCacheCtx(ctx, &sequence, query, tenantId, contractId)
	return sequence, err
}

func (m *customTOptionTradeModel) FindByComboOrderID(
	ctx context.Context, tenantId, comboOrderId, limit int64,
) ([]*TOptionTrade, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	where := `trade.tenant_id=? AND trade.combo_match_no<>''
  AND EXISTS (
    SELECT 1 FROM t_option_order AS child
    WHERE child.tenant_id=trade.tenant_id AND child.combo_order_id=?
      AND (child.id=trade.buy_order_id OR child.id=trade.sell_order_id)
  )`
	var total int64
	if err := m.QueryRowNoCacheCtx(
		ctx, &total, fmt.Sprintf("SELECT COUNT(1) FROM %s AS trade WHERE %s", m.table, where),
		tenantId, comboOrderId,
	); err != nil {
		return nil, 0, err
	}
	query := fmt.Sprintf(`SELECT %s FROM %s AS trade
WHERE %s
ORDER BY trade.combo_match_no,trade.combo_leg_no,trade.id
LIMIT ?`, tOptionTradeRows, m.table, where)
	var list []*TOptionTrade
	if err := m.QueryRowsNoCacheCtx(ctx, &list, query, tenantId, comboOrderId, limit); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// NewTOptionTradeModel returns a model for the database table.
func NewTOptionTradeModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionTradeModel {
	return &customTOptionTradeModel{
		defaultTOptionTradeModel: newTOptionTradeModel(conn, c, opts...),
	}
}

func (m *customTOptionTradeModel) FindPage(ctx context.Context, filter OptionTradePageFilter, cursor int64, limit int64) ([]*TOptionTrade, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqInt64("contract_id", filter.ContractId)
	builder.EqString("trade_no", filter.TradeNo)
	builder.GteInt64("trade_time", filter.TradeTimeStart)
	builder.LteInt64("trade_time", filter.TradeTimeEnd)

	where := builder.Where()
	args := builder.Args()
	if filter.UserId != 0 {
		where += " AND (buy_user_id = ? OR sell_user_id = ?)"
		args = append(args, filter.UserId, filter.UserId)
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
