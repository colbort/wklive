package models

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"wklive/common/sqlutil"
)

var _ TTradeFillModel = (*customTTradeFillModel)(nil)

type (
	TradeFillPageFilter struct {
		TenantId    int64
		UserId      int64
		SymbolId    int64
		ProductType int64
		TimeStart   int64
		TimeEnd     int64
	}

	// TTradeFillModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTTradeFillModel.
	TTradeFillModel interface {
		tTradeFillModel
		FindPage(ctx context.Context, filter TradeFillPageFilter, cursor int64, limit int64) ([]*TTradeFill, int64, error)
		FindLastPrice(ctx context.Context, tenantId, symbolId, marketType int64) (decimal.Decimal, error)
		CountUnsettledByOrder(ctx context.Context, tenantId, orderId int64) (int64, error)
		FindSettlementReady(ctx context.Context, tenantId, cursor, limit int64) ([]*TTradeFill, error)
		UpdateRealizedPnl(ctx context.Context, id int64, realizedPnl decimal.Decimal) error
	}

	customTTradeFillModel struct {
		*defaultTTradeFillModel
	}
)

// NewTTradeFillModel returns a model for the database table.
func NewTTradeFillModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TTradeFillModel {
	return &customTTradeFillModel{
		defaultTTradeFillModel: newTTradeFillModel(conn, c, opts...),
	}
}

func (m *defaultTTradeFillModel) CountUnsettledByOrder(ctx context.Context, tenantId, orderId int64) (int64, error) {
	var count int64
	query := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE tenant_id = ? AND order_id = ? AND settlement_status <> 3", m.table)
	if err := m.QueryRowNoCacheCtx(ctx, &count, query, tenantId, orderId); err != nil {
		return 0, err
	}
	return count, nil
}

func (m *defaultTTradeFillModel) UpdateRealizedPnl(ctx context.Context, id int64, realizedPnl decimal.Decimal) error {
	data, err := m.FindOne(ctx, id)
	if err != nil {
		return err
	}
	idKey := fmt.Sprintf("%s%v", cacheTTradeFillIdPrefix, data.Id)
	fillNoKey := fmt.Sprintf("%s%v:%v", cacheTTradeFillTenantIdFillNoPrefix, data.TenantId, data.FillNo)
	matchOrderKey := fmt.Sprintf("%s%v:%v:%v", cacheTTradeFillTenantIdMatchNoOrderIdPrefix, data.TenantId, data.MatchNo, data.OrderId)
	_, err = m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		query := fmt.Sprintf("UPDATE %s SET realized_pnl = ? WHERE id = ?", m.table)
		return conn.ExecCtx(ctx, query, realizedPnl, id)
	}, idKey, fillNoKey, matchOrderKey)
	return err
}

// FindSettlementReady returns fills whose Asset instructions have all
// succeeded but whose final fill/order state was not committed yet. This is
// the recovery path for the race where the last Asset instruction succeeds
// before a derivative POSITION_FILL_REQUIRED event is acknowledged.
func (m *defaultTTradeFillModel) FindSettlementReady(ctx context.Context, tenantId, cursor, limit int64) ([]*TTradeFill, error) {
	limit = sqlutil.NormalizeLimit(limit)
	tenantClause := ""
	args := make([]any, 0, 3)
	if tenantId > 0 {
		tenantClause = "f.tenant_id = ? AND "
		args = append(args, tenantId)
	}
	query := fmt.Sprintf(`SELECT %s
		FROM %s AS f
		WHERE %sf.id > ?
		  AND f.settlement_status IN (1, 2, 4)
		  AND EXISTS (
			SELECT 1 FROM t_trade_settlement_instruction AS i
			WHERE i.tenant_id = f.tenant_id AND i.fill_id = f.id
		  )
		  AND NOT EXISTS (
			SELECT 1 FROM t_trade_settlement_instruction AS i
			WHERE i.tenant_id = f.tenant_id AND i.fill_id = f.id AND i.status <> 3
		  )
		ORDER BY f.id ASC LIMIT ?`, prefixedColumns(tTradeFillRows, "f"), m.table, tenantClause)
	args = append(args, cursor, limit)
	var fills []*TTradeFill
	if err := m.QueryRowsNoCacheCtx(ctx, &fills, query, args...); err != nil {
		return nil, err
	}
	return fills, nil
}

func prefixedColumns(columns, alias string) string {
	parts := strings.Split(columns, ",")
	for i, part := range parts {
		parts[i] = alias + "." + strings.TrimSpace(part)
	}
	return strings.Join(parts, ",")
}

func (m *defaultTTradeFillModel) FindPage(ctx context.Context, filter TradeFillPageFilter, cursor int64, limit int64) ([]*TTradeFill, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqInt64("user_id", filter.UserId)
	builder.EqInt64("symbol_id", filter.SymbolId)
	builder.EqInt64("product_type", filter.ProductType)
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
	listSQL := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tTradeFillRows, m.table, where)
	if cursor > 0 {
		listSQL += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	listSQL += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)

	var list []*TTradeFill
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSQL, listArgs...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (m *defaultTTradeFillModel) FindLastPrice(ctx context.Context, tenantId, symbolId, productType int64) (decimal.Decimal, error) {
	// decimal.Decimal is itself a struct. Passing it directly to go-zero's
	// strict ORM scanner makes the single price column look like an
	// incomplete multi-field destination and returns ErrNotMatchDestination.
	// Wrap it in a tagged row so the selected column has an explicit target.
	var row struct {
		Price decimal.Decimal `db:"price"`
	}
	sql := fmt.Sprintf("SELECT `price` FROM %s WHERE `tenant_id` = ? AND `symbol_id` = ? AND `product_type` = ? ORDER BY `match_time` DESC, `id` DESC LIMIT 1", m.table)
	err := m.QueryRowNoCacheCtx(ctx, &row, sql, tenantId, symbolId, productType)
	switch err {
	case nil:
		return row.Price, nil
	case sqlc.ErrNotFound:
		return decimal.Zero, ErrNotFound
	default:
		return decimal.Zero, err
	}
}
