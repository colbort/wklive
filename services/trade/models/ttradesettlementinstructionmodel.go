package models

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"strings"
	"wklive/common/sqlutil"
)

var _ TTradeSettlementInstructionModel = (*customTTradeSettlementInstructionModel)(nil)

type (
	// TTradeSettlementInstructionModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTTradeSettlementInstructionModel.
	TTradeSettlementInstructionModel interface {
		tTradeSettlementInstructionModel
		FindPage(ctx context.Context, filter AdminPageFilter, cursor, limit int64) ([]*TTradeSettlementInstruction, int64, error)
		FindPendingFillSettlements(ctx context.Context, tenantId, now, limit int64) ([]*TTradeSettlementInstruction, error)
		FindPendingOrderReleases(ctx context.Context, tenantId, now, limit int64) ([]*TTradeSettlementInstruction, error)
		FindByFillId(ctx context.Context, tenantId, fillId int64) ([]*TTradeSettlementInstruction, error)
		CountUnfinishedByOrder(ctx context.Context, tenantId, orderId int64) (int64, error)
		CountAllUnfinishedByOrder(ctx context.Context, tenantId, orderId int64) (int64, error)
		Claim(ctx context.Context, id, now int64) (bool, error)
	}

	customTTradeSettlementInstructionModel struct {
		*defaultTTradeSettlementInstructionModel
	}
)

// NewTTradeSettlementInstructionModel returns a model for the database table.
func NewTTradeSettlementInstructionModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TTradeSettlementInstructionModel {
	return &customTTradeSettlementInstructionModel{
		defaultTTradeSettlementInstructionModel: newTTradeSettlementInstructionModel(conn, c, opts...),
	}
}

func (m *defaultTTradeSettlementInstructionModel) CountAllUnfinishedByOrder(ctx context.Context, tenantId, orderId int64) (int64, error) {
	var count int64
	query := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE tenant_id = ? AND order_id = ? AND status <> 3", m.table)
	if err := m.QueryRowNoCacheCtx(ctx, &count, query, tenantId, orderId); err != nil {
		return 0, err
	}
	return count, nil
}

func (m *defaultTTradeSettlementInstructionModel) FindPendingOrderReleases(ctx context.Context, tenantId, now, limit int64) ([]*TTradeSettlementInstruction, error) {
	limit = sqlutil.NormalizeLimit(limit)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE (? = 0 OR tenant_id = ?) AND biz_type = 'order' AND action = 2 AND ((status IN (1, 4) AND (next_retry_at = 0 OR next_retry_at <= ?)) OR (status = 2 AND update_times <= ?)) ORDER BY id ASC LIMIT ?", tTradeSettlementInstructionRows, m.table)
	var list []*TTradeSettlementInstruction
	if err := m.QueryRowsNoCacheCtx(ctx, &list, query, tenantId, tenantId, now, now-60*1000, limit); err != nil {
		return nil, err
	}
	return list, nil
}

func (m *defaultTTradeSettlementInstructionModel) CountUnfinishedByOrder(ctx context.Context, tenantId, orderId int64) (int64, error) {
	var count int64
	query := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE tenant_id = ? AND order_id = ? AND status <> 3 AND action <> 2", m.table)
	if err := m.QueryRowNoCacheCtx(ctx, &count, query, tenantId, orderId); err != nil {
		return 0, err
	}
	return count, nil
}

func (m *defaultTTradeSettlementInstructionModel) FindPendingFillSettlements(ctx context.Context, tenantId, now, limit int64) ([]*TTradeSettlementInstruction, error) {
	limit = sqlutil.NormalizeLimit(limit)
	staleBefore := now - 60*1000
	query := fmt.Sprintf("SELECT %s FROM %s i JOIN t_trade_fill f ON f.tenant_id = i.tenant_id AND f.id = i.fill_id WHERE (? = 0 OR i.tenant_id = ?) AND i.biz_type = 'fill' AND f.product_type IN (1, 2) AND ((i.status IN (1, 4) AND (i.next_retry_at = 0 OR i.next_retry_at <= ?)) OR (i.status = 2 AND i.update_times <= ?)) AND NOT EXISTS (SELECT 1 FROM %s p WHERE p.tenant_id = i.tenant_id AND p.fill_id = i.fill_id AND p.step_no < i.step_no AND p.status <> 3) ORDER BY i.id ASC LIMIT ?", prefixedSettlementInstructionRows("i"), m.table, m.table)
	var list []*TTradeSettlementInstruction
	if err := m.QueryRowsNoCacheCtx(ctx, &list, query, tenantId, tenantId, now, staleBefore, limit); err != nil {
		return nil, err
	}
	return list, nil
}

func (m *defaultTTradeSettlementInstructionModel) FindByFillId(ctx context.Context, tenantId, fillId int64) ([]*TTradeSettlementInstruction, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE tenant_id = ? AND fill_id = ? ORDER BY step_no ASC, id ASC", tTradeSettlementInstructionRows, m.table)
	var list []*TTradeSettlementInstruction
	if err := m.QueryRowsNoCacheCtx(ctx, &list, query, tenantId, fillId); err != nil {
		return nil, err
	}
	return list, nil
}

func (m *defaultTTradeSettlementInstructionModel) Claim(ctx context.Context, id, now int64) (bool, error) {
	item, err := m.FindOne(ctx, id)
	if err != nil {
		return false, err
	}
	idKey := fmt.Sprintf("%s%v", cacheTTradeSettlementInstructionIdPrefix, id)
	uniqueKey := fmt.Sprintf("%s%v:%v", cacheTTradeSettlementInstructionTenantIdInstructionNoPrefix, item.TenantId, item.InstructionNo)
	result, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		query := fmt.Sprintf("UPDATE %s SET status = 2, update_times = ? WHERE id = ? AND ((status IN (1, 4) AND (next_retry_at = 0 OR next_retry_at <= ?)) OR (status = 2 AND update_times <= ?))", m.table)
		return conn.ExecCtx(ctx, query, now, id, now, now-60*1000)
	}, idKey, uniqueKey)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	return rows == 1, err
}

func prefixedSettlementInstructionRows(alias string) string {
	rows := make([]string, 0, len(tTradeSettlementInstructionFieldNames))
	for _, field := range tTradeSettlementInstructionFieldNames {
		rows = append(rows, alias+"."+field)
	}
	return strings.Join(rows, ",")
}

func (m *defaultTTradeSettlementInstructionModel) FindPage(ctx context.Context, filter AdminPageFilter, cursor, limit int64) ([]*TTradeSettlementInstruction, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	b := adminPageBuilder(filter, "")
	where, args := b.Where(), b.Args()
	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where), args...); err != nil {
		return nil, 0, err
	}
	la := append([]any{}, args...)
	q := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tTradeSettlementInstructionRows, m.table, where)
	if cursor > 0 {
		q += " AND id < ?"
		la = append(la, cursor)
	}
	q += " ORDER BY id DESC LIMIT ?"
	la = append(la, limit)
	var list []*TTradeSettlementInstruction
	if err := m.QueryRowsNoCacheCtx(ctx, &list, q, la...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
