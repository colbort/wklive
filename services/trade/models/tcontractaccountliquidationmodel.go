package models

import (
	"context"
	"database/sql"
	"fmt"

	"wklive/common/sqlutil"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TContractAccountLiquidationModel = (*customTContractAccountLiquidationModel)(nil)

const (
	ContractAccountLiquidationStatusPending       int64 = 1
	ContractAccountLiquidationStatusAssetSettling int64 = 2
	ContractAccountLiquidationStatusClosing       int64 = 3
	ContractAccountLiquidationStatusCompleted     int64 = 4
	ContractAccountLiquidationStatusManualReview  int64 = 5
	ContractAccountLiquidationStatusInsuranceFund int64 = 6
	ContractAccountLiquidationStatusADL           int64 = 7

	ContractAccountLiquidationItemStatusLocked int64 = 1
	ContractAccountLiquidationItemStatusClosed int64 = 2
)

type (
	// TContractAccountLiquidationModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTContractAccountLiquidationModel.
	TContractAccountLiquidationModel interface {
		tContractAccountLiquidationModel
		FindOneForUpdate(ctx context.Context, id int64) (*TContractAccountLiquidation, error)
		FindActiveByRiskUnit(ctx context.Context, tenantID, userID int64, marginAsset string) (*TContractAccountLiquidation, error)
		FindRecoverable(ctx context.Context, tenantID, limit int64) ([]*TContractAccountLiquidation, error)
		CountUnfinishedByRiskUnit(ctx context.Context, tenantID, userID int64, marginAsset string) (int64, error)
		FindPage(ctx context.Context, filter AdminPageFilter, cursor, limit int64) ([]*TContractAccountLiquidation, int64, error)
	}

	customTContractAccountLiquidationModel struct {
		*defaultTContractAccountLiquidationModel
		conn sqlx.SqlConn
	}
)

// NewTContractAccountLiquidationModel returns a model for the database table.
func NewTContractAccountLiquidationModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TContractAccountLiquidationModel {
	return &customTContractAccountLiquidationModel{
		defaultTContractAccountLiquidationModel: newTContractAccountLiquidationModel(conn, c, opts...),
		conn:                                    conn,
	}
}

func (m *customTContractAccountLiquidationModel) FindOneForUpdate(
	ctx context.Context, id int64,
) (*TContractAccountLiquidation, error) {
	return m.findOneNoCache(ctx, "id=?", true, id)
}

func (m *customTContractAccountLiquidationModel) FindActiveByRiskUnit(
	ctx context.Context, tenantID, userID int64, marginAsset string,
) (*TContractAccountLiquidation, error) {
	return m.findOneNoCache(
		ctx,
		"tenant_id=? AND user_id=? AND margin_asset=? AND status IN (1,2,3,5,6,7) ORDER BY id DESC",
		false, tenantID, userID, marginAsset,
	)
}

func (m *customTContractAccountLiquidationModel) findOneNoCache(
	ctx context.Context, where string, forUpdate bool, args ...any,
) (*TContractAccountLiquidation, error) {
	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE %s LIMIT 1",
		tContractAccountLiquidationRows, m.table, where,
	)
	if forUpdate {
		query += " FOR UPDATE"
	}
	var row TContractAccountLiquidation
	if err := m.conn.QueryRowCtx(ctx, &row, query, args...); err != nil {
		if err == sqlx.ErrNotFound || err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &row, nil
}

func (m *customTContractAccountLiquidationModel) FindRecoverable(
	ctx context.Context, tenantID, limit int64,
) ([]*TContractAccountLiquidation, error) {
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	var rows []*TContractAccountLiquidation
	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE (?=0 OR tenant_id=?) AND status IN (1,2,3,6,7) ORDER BY update_times,id LIMIT ?",
		tContractAccountLiquidationRows, m.table,
	)
	if err := m.conn.QueryRowsCtx(ctx, &rows, query, tenantID, tenantID, limit); err != nil {
		return nil, err
	}
	return rows, nil
}

func (m *customTContractAccountLiquidationModel) CountUnfinishedByRiskUnit(
	ctx context.Context, tenantID, userID int64, marginAsset string,
) (int64, error) {
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, fmt.Sprintf(`SELECT COUNT(1)
FROM %s
WHERE tenant_id=? AND user_id=? AND margin_asset=? AND status<>?`, m.table),
		tenantID, userID, marginAsset, ContractAccountLiquidationStatusCompleted)
	return count, err
}

func (m *customTContractAccountLiquidationModel) FindPage(
	ctx context.Context, filter AdminPageFilter, cursor, limit int64,
) ([]*TContractAccountLiquidation, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	builder := adminPageBuilder(filter, "create_times")
	where, args := builder.Where(), builder.Args()
	var total int64
	if err := m.conn.QueryRowCtx(
		ctx, &total, fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where), args...,
	); err != nil {
		return nil, 0, err
	}
	listArgs := append([]any{}, args...)
	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE %s",
		tContractAccountLiquidationRows, m.table, where,
	)
	if cursor > 0 {
		query += " AND id<?"
		listArgs = append(listArgs, cursor)
	}
	query += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)
	var rows []*TContractAccountLiquidation
	if err := m.conn.QueryRowsCtx(ctx, &rows, query, listArgs...); err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}
