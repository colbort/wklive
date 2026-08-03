package models

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TStakeOperationModel = (*customTStakeOperationModel)(nil)

type (
	// TStakeOperationModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTStakeOperationModel.
	TStakeOperationModel interface {
		tStakeOperationModel
		Claim(ctx context.Context, id, now int64) (bool, error)
		MarkRetryable(ctx context.Context, id, retryCount, nextRetryAt, status, now int64, lastError string) error
		FindRetryablePage(ctx context.Context, tenantId, now, cursor, limit int64) ([]*TStakeOperation, error)
		FindAdminPage(ctx context.Context, filter StakeOperationPageFilter, cursor, limit int64) ([]*TStakeOperation, int64, error)
		ResetForManualRetry(ctx context.Context, id, operatorID, now int64, reason string) (bool, error)
	}

	customTStakeOperationModel struct {
		*defaultTStakeOperationModel
	}
)

type StakeOperationPageFilter struct {
	TenantId      int64
	UserId        int64
	OperationType int64
	Status        int64
	OperationNo   string
	OrderNo       string
}

// NewTStakeOperationModel returns a model for the database table.
func NewTStakeOperationModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TStakeOperationModel {
	return &customTStakeOperationModel{
		defaultTStakeOperationModel: newTStakeOperationModel(conn, c, opts...),
	}
}

func (m *defaultTStakeOperationModel) Claim(ctx context.Context, id, now int64) (bool, error) {
	staleBefore := now - 60_000
	result, err := m.ExecNoCacheCtx(ctx, `UPDATE t_stake_operation
		SET status=2,version=version+1,update_times=?
		WHERE id=? AND ((status IN (1,4) AND next_retry_at<=?) OR (status=2 AND update_times<=?))`, now, id, now, staleBefore)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	return affected == 1, err
}

func (m *defaultTStakeOperationModel) MarkRetryable(ctx context.Context, id, retryCount, nextRetryAt, status, now int64, lastError string) error {
	_, err := m.ExecNoCacheCtx(ctx, `UPDATE t_stake_operation
		SET status=?,retry_count=?,next_retry_at=?,last_error=?,version=version+1,update_times=?
		WHERE id=? AND status=2`, status, retryCount, nextRetryAt, lastError, now, id)
	return err
}

func (m *defaultTStakeOperationModel) FindRetryablePage(ctx context.Context, tenantId, now, cursor, limit int64) ([]*TStakeOperation, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	query := fmt.Sprintf("SELECT %s FROM %s WHERE ((status IN (1,4) AND next_retry_at<=?) OR (status=2 AND update_times<=?))", tStakeOperationRows, m.table)
	args := []any{now, now - 60_000}
	if tenantId > 0 {
		query += " AND tenant_id=?"
		args = append(args, tenantId)
	}
	if cursor > 0 {
		query += " AND id>?"
		args = append(args, cursor)
	}
	query += " ORDER BY id ASC LIMIT ?"
	args = append(args, limit)
	var items []*TStakeOperation
	if err := m.QueryRowsNoCacheCtx(ctx, &items, query, args...); err != nil {
		return nil, err
	}
	return items, nil
}

func (m *defaultTStakeOperationModel) FindAdminPage(ctx context.Context, filter StakeOperationPageFilter, cursor, limit int64) ([]*TStakeOperation, int64, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	where := " WHERE 1=1"
	args := make([]any, 0, 8)
	if filter.TenantId > 0 {
		where += " AND tenant_id=?"
		args = append(args, filter.TenantId)
	}
	if filter.UserId > 0 {
		where += " AND user_id=?"
		args = append(args, filter.UserId)
	}
	if filter.OperationType > 0 {
		where += " AND operation_type=?"
		args = append(args, filter.OperationType)
	}
	if filter.Status > 0 {
		where += " AND status=?"
		args = append(args, filter.Status)
	}
	if filter.OperationNo != "" {
		where += " AND operation_no LIKE ?"
		args = append(args, "%"+filter.OperationNo+"%")
	}
	if filter.OrderNo != "" {
		where += " AND order_no LIKE ?"
		args = append(args, "%"+filter.OrderNo+"%")
	}

	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, "SELECT COUNT(*) FROM "+m.table+where, args...); err != nil {
		return nil, 0, err
	}
	pageWhere := where
	pageArgs := append([]any(nil), args...)
	if cursor > 0 {
		pageWhere += " AND id<?"
		pageArgs = append(pageArgs, cursor)
	}
	pageArgs = append(pageArgs, limit)
	query := fmt.Sprintf("SELECT %s FROM %s%s ORDER BY id DESC LIMIT ?", tStakeOperationRows, m.table, pageWhere)
	var items []*TStakeOperation
	if err := m.QueryRowsNoCacheCtx(ctx, &items, query, pageArgs...); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (m *defaultTStakeOperationModel) ResetForManualRetry(ctx context.Context, id, operatorID, now int64, reason string) (bool, error) {
	item, err := m.FindOne(ctx, id)
	if err != nil {
		return false, err
	}
	idKey := fmt.Sprintf("%s%v", cacheTStakeOperationIdPrefix, item.Id)
	operationKey := fmt.Sprintf("%s%v:%v", cacheTStakeOperationTenantIdOperationNoPrefix, item.TenantId, item.OperationNo)
	requestKey := fmt.Sprintf("%s%v:%v:%v:%v", cacheTStakeOperationTenantIdUserIdOperationTypeRequestNoPrefix, item.TenantId, item.UserId, item.OperationType, item.RequestNo)
	result, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		return conn.ExecCtx(ctx, `UPDATE t_stake_operation
			SET status=4,next_retry_at=?,last_error=CONCAT('manual retry by ',?,': ',?),version=version+1,update_times=?
			WHERE id=? AND status IN (4,5)`, now, operatorID, reason, now, id)
	}, idKey, operationKey, requestKey)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	return affected == 1, err
}
