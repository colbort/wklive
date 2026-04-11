package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ItickKlineSyncProgressModel interface {
	tItickKlineSyncProgressModel
	FindOrCreate(ctx context.Context, categoryCode, market, symbol, interval string) (*TItickKlineSyncProgress, error)
	UpdateSyncStart(ctx context.Context, id int64, mode string, now int64) error
	UpdateSyncSuccess(ctx context.Context, id int64, mode string, latestTs, oldestTs, fullSynced, now int64, message string) error
	UpdateSyncFail(ctx context.Context, id int64, mode string, now int64, message string) error
}

func (m *defaultTItickKlineSyncProgressModel) FindOrCreate(ctx context.Context, categoryCode, market, symbol, interval string) (*TItickKlineSyncProgress, error) {
	exist, err := m.FindOneByCategoryCodeMarketSymbolInterval(ctx, categoryCode, market, symbol, interval)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, err
	}
	if exist != nil {
		return exist, nil
	}

	now := utils.NowMillis()
	_, err = m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		return conn.ExecCtx(ctx, fmt.Sprintf(`
		insert ignore into %s
		(category_code, market, symbol, `+"`interval`"+`, latest_ts, oldest_ts, full_synced, sync_status, last_sync_mode, last_sync_message, last_success_time, last_fail_time, create_times, update_times)
		values (?, ?, ?, ?, 0, 0, 0, 0, '', '', 0, 0, ?, ?)
	`, m.table), categoryCode, market, symbol, interval, now, now)
	})
	if err != nil {
		return nil, err
	}

	return m.FindOneByCategoryCodeMarketSymbolInterval(ctx, categoryCode, market, symbol, interval)
}

func (m *defaultTItickKlineSyncProgressModel) UpdateSyncStart(ctx context.Context, id int64, mode string, now int64) error {
	query := fmt.Sprintf(`
		update %s
		set sync_status = 1,
			last_sync_mode = ?,
			last_sync_message = '同步中',
			update_times = ?
		where id = ?
	`, m.table)

	_, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		return conn.ExecCtx(ctx, query, mode, now, id)
	})
	return err
}

func (m *defaultTItickKlineSyncProgressModel) UpdateSyncSuccess(ctx context.Context, id int64, mode string, latestTs, oldestTs, fullSynced, now int64, message string) error {
	query := fmt.Sprintf(`
		update %s
		set sync_status = 2,
			last_sync_mode = ?,
			latest_ts = ?,
			oldest_ts = ?,
			full_synced = ?,
			last_success_time = ?,
			last_sync_message = ?,
			update_times = ?
		where id = ?
	`, m.table)

	_, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		return conn.ExecCtx(ctx, query, mode, latestTs, oldestTs, fullSynced, now, message, now, id)
	})
	return err
}

func (m *defaultTItickKlineSyncProgressModel) UpdateSyncFail(ctx context.Context, id int64, mode string, now int64, message string) error {
	query := fmt.Sprintf(`
		update %s
		set sync_status = 3,
			last_sync_mode = ?,
			last_fail_time = ?,
			last_sync_message = ?,
			update_times = ?
		where id = ?
	`, m.table)

	_, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		return conn.ExecCtx(ctx, query, mode, now, message, now, id)
	})
	return err
}
