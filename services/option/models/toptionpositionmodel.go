package models

import (
	"context"
	"fmt"
	"strings"

	"wklive/common/sqlutil"
	"wklive/proto/common"
	"wklive/proto/option"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionPositionModel = (*customTOptionPositionModel)(nil)

type (
	OptionPositionPageFilter struct {
		TenantId   int64
		UserId     int64
		AccountId  int64
		ContractId int64
		Side       int64
		Status     int64
		Statuses   []int64
	}

	OptionOpenInterest struct {
		ContractId int64           `db:"contract_id"`
		LongQty    decimal.Decimal `db:"long_qty"`
		ShortQty   decimal.Decimal `db:"short_qty"`
		AsOf       int64           `db:"as_of"`
	}

	// TOptionPositionModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionPositionModel.
	TOptionPositionModel interface {
		tOptionPositionModel
		FindPage(ctx context.Context, filter OptionPositionPageFilter, cursor int64, limit int64) ([]*TOptionPosition, int64, error)
		FindAssignableShortsPage(ctx context.Context, tenantId, contractId, afterCreateTimes, afterId, limit int64) ([]*TOptionPosition, error)
		FindOneForUpdate(ctx context.Context, id int64) (*TOptionPosition, error)
		FindOneByTenantIdUserIdAccountIdContractIdSideForUpdate(ctx context.Context, tenantId, userId, accountId, contractId, side int64) (*TOptionPosition, error)
		SumHoldingQty(ctx context.Context, tenantId, userId, contractId, side int64) (decimal.Decimal, error)
		CountHoldingByContract(ctx context.Context, tenantId, contractId int64) (int64, error)
		FindHoldingBatch(ctx context.Context, tenantId, contractId, afterId, limit int64) ([]*TOptionPosition, error)
		FindOpenInterestByContracts(ctx context.Context, tenantId int64, contractIDs []int64) ([]*OptionOpenInterest, error)
	}

	customTOptionPositionModel struct {
		*defaultTOptionPositionModel
	}
)

func (m *customTOptionPositionModel) FindOpenInterestByContracts(
	ctx context.Context, tenantId int64, contractIDs []int64,
) ([]*OptionOpenInterest, error) {
	if len(contractIDs) == 0 {
		return []*OptionOpenInterest{}, nil
	}
	args := make([]any, 0, len(contractIDs)+4)
	args = append(args,
		int64(common.PositionSide_POSITION_SIDE_LONG),
		int64(common.PositionSide_POSITION_SIDE_SHORT),
		tenantId,
		int64(option.PositionStatus_POSITION_STATUS_HOLDING),
	)
	for _, id := range contractIDs {
		args = append(args, id)
	}
	placeholders := strings.TrimRight(strings.Repeat("?,", len(contractIDs)), ",")
	query := fmt.Sprintf(`SELECT contract_id,
  COALESCE(SUM(CASE WHEN side=? THEN position_qty ELSE 0 END),0) AS long_qty,
  COALESCE(SUM(CASE WHEN side=? THEN position_qty ELSE 0 END),0) AS short_qty,
  COALESCE(MAX(update_times),0) AS as_of
FROM %s
WHERE tenant_id=? AND status=? AND position_qty>0 AND contract_id IN (%s)
GROUP BY contract_id`, m.table, placeholders)
	var items []*OptionOpenInterest
	err := m.QueryRowsNoCacheCtx(ctx, &items, query, args...)
	return items, err
}

func (m *customTOptionPositionModel) CountHoldingByContract(
	ctx context.Context, tenantId, contractId int64,
) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE tenant_id=? AND contract_id=? AND status=? AND position_qty>0",
		m.table)
	var total int64
	err := m.QueryRowNoCacheCtx(ctx, &total, query, tenantId, contractId,
		int64(option.PositionStatus_POSITION_STATUS_HOLDING))
	return total, err
}

func (m *customTOptionPositionModel) FindHoldingBatch(
	ctx context.Context, tenantId, contractId, afterId, limit int64,
) ([]*TOptionPosition, error) {
	limit = sqlutil.NormalizeLimit(limit)
	query := fmt.Sprintf(`SELECT %s FROM %s
WHERE tenant_id=? AND contract_id=? AND status=? AND position_qty>0 AND id>?
ORDER BY id LIMIT ?`, tOptionPositionRows, m.table)
	var items []*TOptionPosition
	err := m.QueryRowsNoCacheCtx(ctx, &items, query, tenantId, contractId,
		int64(option.PositionStatus_POSITION_STATUS_HOLDING), afterId, limit)
	return items, err
}

func (m *customTOptionPositionModel) SumHoldingQty(
	ctx context.Context, tenantId, userId, contractId, side int64,
) (decimal.Decimal, error) {
	userClause := ""
	args := []any{tenantId, contractId, side, int64(option.PositionStatus_POSITION_STATUS_HOLDING)}
	if userId > 0 {
		userClause = " AND user_id = ?"
		args = append(args, userId)
	}
	query := fmt.Sprintf(`SELECT COALESCE(SUM(position_qty), 0) AS total FROM %s
WHERE tenant_id = ? AND contract_id = ? AND side = ? AND status = ?%s`, m.table, userClause)
	var aggregate decimalAggregate
	if err := m.QueryRowNoCacheCtx(ctx, &aggregate, query, args...); err != nil {
		return decimal.Zero, err
	}
	return aggregate.Decimal()
}

func (m *customTOptionPositionModel) FindAssignableShortsPage(
	ctx context.Context,
	tenantId, contractId, afterCreateTimes, afterId, limit int64,
) ([]*TOptionPosition, error) {
	if limit <= 0 || limit > 500 {
		limit = 500
	}
	query := fmt.Sprintf(`SELECT %s FROM %s
WHERE tenant_id = ? AND contract_id = ? AND side = ? AND status = ? AND position_qty > 0
	  AND (create_times > ? OR (create_times = ? AND id > ?))
ORDER BY create_times ASC, id ASC LIMIT ?`, tOptionPositionRows, m.table)
	var list []*TOptionPosition
	err := m.QueryRowsNoCacheCtx(
		ctx, &list, query,
		tenantId, contractId,
		int64(common.PositionSide_POSITION_SIDE_SHORT),
		int64(option.PositionStatus_POSITION_STATUS_HOLDING),
		afterCreateTimes, afterCreateTimes, afterId, limit,
	)
	return list, err
}

func (m *customTOptionPositionModel) FindOneForUpdate(ctx context.Context, id int64) (*TOptionPosition, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE id = ? LIMIT 1 FOR UPDATE", tOptionPositionRows, m.table)
	var item TOptionPosition
	if err := m.QueryRowNoCacheCtx(ctx, &item, query, id); err != nil {
		return nil, err
	}
	return &item, nil
}

func (m *customTOptionPositionModel) FindOneByTenantIdUserIdAccountIdContractIdSideForUpdate(
	ctx context.Context, tenantId, userId, accountId, contractId, side int64,
) (*TOptionPosition, error) {
	query := fmt.Sprintf(`SELECT %s FROM %s
WHERE tenant_id = ? AND user_id = ? AND account_id = ? AND contract_id = ? AND side = ?
LIMIT 1 FOR UPDATE`, tOptionPositionRows, m.table)
	var item TOptionPosition
	if err := m.QueryRowNoCacheCtx(
		ctx, &item, query, tenantId, userId, accountId, contractId, side,
	); err != nil {
		return nil, err
	}
	return &item, nil
}

// NewTOptionPositionModel returns a model for the database table.
func NewTOptionPositionModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionPositionModel {
	return &customTOptionPositionModel{
		defaultTOptionPositionModel: newTOptionPositionModel(conn, c, opts...),
	}
}

func (m *customTOptionPositionModel) FindPage(ctx context.Context, filter OptionPositionPageFilter, cursor int64, limit int64) ([]*TOptionPosition, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqInt64("user_id", filter.UserId)
	builder.EqInt64("account_id", filter.AccountId)
	builder.EqInt64("contract_id", filter.ContractId)
	builder.EqInt64("side", filter.Side)
	builder.EqInt64("status", filter.Status)
	builder.InInt64("status", filter.Statuses)

	where := builder.Where()
	args := builder.Args()

	var total int64
	countSql := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSql, args...); err != nil {
		return nil, 0, err
	}

	listArgs := append([]any{}, args...)
	listSql := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tOptionPositionRows, m.table, where)
	if cursor > 0 {
		listSql += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	listSql += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)

	var list []*TOptionPosition
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
