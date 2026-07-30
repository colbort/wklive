package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TMarketPriceFormulaModel = (*customTMarketPriceFormulaModel)(nil)

type (
	// TMarketPriceFormulaModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTMarketPriceFormulaModel.
	TMarketPriceFormulaModel interface {
		tMarketPriceFormulaModel
		FindDue(context.Context, int64, int64) ([]*TMarketPriceFormula, error)
		ClaimTarget(context.Context, int64, int64, int64, int64) (bool, error)
		ReleaseTarget(context.Context, int64, int64, int64, int64) error
		ActivateVersion(context.Context, int64, int64) error
		RevokeVersion(context.Context, int64, int64) error
		FindPage(context.Context, PriceFormulaFilter, int64, int64) ([]*TMarketPriceFormula, int64, error)
	}

	customTMarketPriceFormulaModel struct {
		*defaultTMarketPriceFormulaModel
	}

	PriceFormulaFilter struct {
		Authority, SnapshotKind, CategoryCode, Market, Symbol string
		Status                                                int64
	}
)

// NewTMarketPriceFormulaModel returns a model for the database table.
func NewTMarketPriceFormulaModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TMarketPriceFormulaModel {
	return &customTMarketPriceFormulaModel{
		defaultTMarketPriceFormulaModel: newTMarketPriceFormulaModel(conn, c, opts...),
	}
}

// ActivateVersion atomically makes the selected immutable formula revision the
// sole active revision for its output. Formula content is changed by inserting
// a new revision, never by editing the active row in place.
func (m *defaultTMarketPriceFormulaModel) ActivateVersion(ctx context.Context, id, now int64) error {
	var revisions []TMarketPriceFormula
	err := m.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		var selected TMarketPriceFormula
		query := fmt.Sprintf("SELECT %s FROM %s WHERE id=? FOR UPDATE", tMarketPriceFormulaRows, m.table)
		if err := session.QueryRowCtx(ctx, &selected, query, id); err != nil {
			return err
		}
		if selected.Status == 3 {
			return errors.New("revoked price formula cannot be activated")
		}
		query = fmt.Sprintf("SELECT %s FROM %s WHERE authority=? AND snapshot_kind=? AND category_code=? AND market=? AND symbol=? FOR UPDATE", tMarketPriceFormulaRows, m.table)
		if err := session.QueryRowsCtx(ctx, &revisions, query, selected.Authority, selected.SnapshotKind, selected.CategoryCode, selected.Market, selected.Symbol); err != nil {
			return err
		}
		if _, err := session.ExecCtx(ctx, "UPDATE t_itick_price_formula SET status=2,version=version+1,update_times=? WHERE authority=? AND snapshot_kind=? AND category_code=? AND market=? AND symbol=? AND status=1 AND id<>?", now, selected.Authority, selected.SnapshotKind, selected.CategoryCode, selected.Market, selected.Symbol, id); err != nil {
			return err
		}
		_, err := session.ExecCtx(ctx, "UPDATE t_itick_price_formula SET status=1,version=version+1,update_times=? WHERE id=? AND status<>3", now, id)
		return err
	})
	if err != nil {
		return err
	}
	return m.DelCacheCtx(ctx, priceFormulaCacheKeys(revisions...)...)
}

// RevokeVersion is irreversible; revoked revisions cannot be activated again.
func (m *defaultTMarketPriceFormulaModel) RevokeVersion(ctx context.Context, id, now int64) error {
	row, err := m.FindOne(ctx, id)
	if err != nil {
		return err
	}
	result, err := m.ExecNoCacheCtx(ctx, "UPDATE t_itick_price_formula SET status=3,version=version+1,update_times=? WHERE id=? AND status<>3", now, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return sql.ErrNoRows
	}
	return m.DelCacheCtx(ctx, priceFormulaCacheKeys(*row)...)
}

func priceFormulaCacheKeys(rows ...TMarketPriceFormula) []string {
	keys := make([]string, 0, len(rows)*3)
	for _, row := range rows {
		keys = append(keys,
			fmt.Sprintf("%s%v", cacheTMarketPriceFormulaIdPrefix, row.Id),
			fmt.Sprintf("%s%v", cacheTMarketPriceFormulaFormulaNoPrefix, row.FormulaNo),
			fmt.Sprintf("%s%v:%v:%v:%v:%v:%v", cacheTMarketPriceFormulaAuthoritySnapshotKindCategoryCodeMarketSymbolFormulaVersionPrefix, row.Authority, row.SnapshotKind, row.CategoryCode, row.Market, row.Symbol, row.FormulaVersion),
		)
	}
	return keys
}

func (m *defaultTMarketPriceFormulaModel) FindPage(ctx context.Context, filter PriceFormulaFilter, cursor, limit int64) ([]*TMarketPriceFormula, int64, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	where, args := "id>?", []any{cursor}
	appendFilter := func(column, value string) {
		if value != "" {
			where += " AND " + column + "=?"
			args = append(args, value)
		}
	}
	appendFilter("authority", filter.Authority)
	appendFilter("snapshot_kind", filter.SnapshotKind)
	appendFilter("category_code", filter.CategoryCode)
	appendFilter("market", filter.Market)
	appendFilter("symbol", filter.Symbol)
	if filter.Status > 0 {
		where += " AND status=?"
		args = append(args, filter.Status)
	}
	var total int64
	countWhere := strings.Replace(where, "id>?", "id>0", 1)
	countArgs := args[1:]
	if err := m.QueryRowNoCacheCtx(ctx, &total, "SELECT COUNT(1) FROM t_itick_price_formula WHERE "+countWhere, countArgs...); err != nil {
		return nil, 0, err
	}
	args = append(args, limit)
	var rows []*TMarketPriceFormula
	err := m.QueryRowsNoCacheCtx(ctx, &rows, "SELECT "+tMarketPriceFormulaRows+" FROM t_itick_price_formula WHERE "+where+" ORDER BY id LIMIT ?", args...)
	return rows, total, err
}

func (m *defaultTMarketPriceFormulaModel) ReleaseTarget(ctx context.Context, id, claimedRunVersion, target, previous int64) error {
	_, err := m.ExecNoCacheCtx(ctx, "UPDATE t_itick_price_formula SET last_target_time=?,update_times=? WHERE id=? AND run_version=? AND last_target_time=?", previous, time.Now().UnixMilli(), id, claimedRunVersion, target)
	return err
}

func (m *defaultTMarketPriceFormulaModel) FindDue(ctx context.Context, now, limit int64) ([]*TMarketPriceFormula, error) {
	var rows []*TMarketPriceFormula
	err := m.QueryRowsNoCacheCtx(ctx, &rows, fmt.Sprintf("SELECT %s FROM %s WHERE status=1 AND last_target_time+interval_ms<=? ORDER BY id LIMIT ?", tMarketPriceFormulaRows, m.table), now, limit)
	return rows, err
}

func (m *defaultTMarketPriceFormulaModel) ClaimTarget(ctx context.Context, id, runVersion, target, now int64) (bool, error) {
	result, err := m.ExecNoCacheCtx(ctx, "UPDATE t_itick_price_formula SET last_target_time=?,run_version=run_version+1,update_times=? WHERE id=? AND run_version=? AND status=1 AND last_target_time<?", target, now, id, runVersion, target)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	if err == sql.ErrNoRows {
		return false, nil
	}
	return rows == 1, err
}
