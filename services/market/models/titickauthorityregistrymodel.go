package models

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TItickAuthorityRegistryModel = (*customTItickAuthorityRegistryModel)(nil)

type (
	// TItickAuthorityRegistryModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTItickAuthorityRegistryModel.
	TItickAuthorityRegistryModel interface {
		tItickAuthorityRegistryModel
		FindEnabled(context.Context, string) (*TItickAuthorityRegistry, error)
		Create(context.Context, *TItickAuthorityRegistry) (int64, error)
		FindPage(context.Context, AuthorityRegistryFilter, int64, int64) ([]*TItickAuthorityRegistry, int64, error)
		CountActiveFormulaReferences(context.Context, string) (int64, error)
		UpdateConfigVersioned(context.Context, int64, int64, string, int64, int64) (bool, error)
	}

	customTItickAuthorityRegistryModel struct {
		*defaultTItickAuthorityRegistryModel
	}

	AuthorityRegistryFilter struct {
		Authority, ProviderCode, ProducerType, SnapshotKind string
		Status                                              int64
	}
)

// NewTItickAuthorityRegistryModel returns a model for the database table.
func NewTItickAuthorityRegistryModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TItickAuthorityRegistryModel {
	return &customTItickAuthorityRegistryModel{
		defaultTItickAuthorityRegistryModel: newTItickAuthorityRegistryModel(conn, c, opts...),
	}
}

func (r *TItickAuthorityRegistry) Allows(kind string) bool {
	if r == nil || r.Status != 1 {
		return false
	}
	var kinds []string
	if json.Unmarshal([]byte(r.AllowedKinds), &kinds) != nil {
		return false
	}
	kind = strings.ToUpper(strings.TrimSpace(kind))
	for _, allowed := range kinds {
		if strings.ToUpper(strings.TrimSpace(allowed)) == kind {
			return true
		}
	}
	return false
}

func (m *defaultTItickAuthorityRegistryModel) FindEnabled(ctx context.Context, authority string) (*TItickAuthorityRegistry, error) {
	var row TItickAuthorityRegistry
	err := m.QueryRowNoCacheCtx(ctx, &row, `SELECT id,authority,provider_code,producer_type,allowed_kinds,status,version,create_times,update_times
FROM t_itick_authority_registry WHERE authority=? AND status=1 LIMIT 1`, strings.ToLower(strings.TrimSpace(authority)))
	if errors.Is(err, sqlx.ErrNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (m *defaultTItickAuthorityRegistryModel) Create(ctx context.Context, row *TItickAuthorityRegistry) (int64, error) {
	result, err := m.Insert(ctx, row)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (m *defaultTItickAuthorityRegistryModel) FindPage(
	ctx context.Context,
	filter AuthorityRegistryFilter,
	cursor, limit int64,
) ([]*TItickAuthorityRegistry, int64, error) {
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
	appendFilter("provider_code", filter.ProviderCode)
	appendFilter("producer_type", filter.ProducerType)
	if filter.SnapshotKind != "" {
		where += " AND JSON_CONTAINS(allowed_kinds,JSON_QUOTE(?))"
		args = append(args, filter.SnapshotKind)
	}
	if filter.Status > 0 {
		where += " AND status=?"
		args = append(args, filter.Status)
	}

	countWhere := strings.Replace(where, "id>?", "id>0", 1)
	countArgs := args[1:]
	var total int64
	if err := m.QueryRowNoCacheCtx(
		ctx,
		&total,
		"SELECT COUNT(1) FROM t_itick_authority_registry WHERE "+countWhere,
		countArgs...,
	); err != nil {
		return nil, 0, err
	}

	args = append(args, limit)
	var rows []*TItickAuthorityRegistry
	err := m.QueryRowsNoCacheCtx(
		ctx,
		&rows,
		"SELECT "+tItickAuthorityRegistryRows+
			" FROM t_itick_authority_registry WHERE "+where+" ORDER BY id LIMIT ?",
		args...,
	)
	return rows, total, err
}

func (m *defaultTItickAuthorityRegistryModel) CountActiveFormulaReferences(
	ctx context.Context,
	authority string,
) (int64, error) {
	const query = `
SELECT COUNT(1)
FROM t_itick_price_formula AS f
WHERE f.status=1
  AND (
    f.authority=?
    OR EXISTS (
      SELECT 1
      FROM JSON_TABLE(
        f.components,
        '$[*]' COLUMNS(authority VARCHAR(32) PATH '$.authority')
      ) AS component
      WHERE component.authority=?
    )
  )`
	var count int64
	err := m.QueryRowNoCacheCtx(ctx, &count, query, authority, authority)
	return count, err
}

func (m *defaultTItickAuthorityRegistryModel) UpdateConfigVersioned(
	ctx context.Context,
	id, expectedVersion int64,
	allowedKinds string,
	status, now int64,
) (bool, error) {
	row, err := m.FindOne(ctx, id)
	if err != nil {
		return false, err
	}
	result, err := m.ExecNoCacheCtx(
		ctx,
		`UPDATE t_itick_authority_registry
SET allowed_kinds=?,status=?,version=version+1,update_times=?
WHERE id=? AND version=?`,
		allowedKinds,
		status,
		now,
		id,
		expectedVersion,
	)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	if affected != 1 {
		return false, nil
	}
	err = m.DelCacheCtx(
		ctx,
		fmt.Sprintf("%s%v", cacheTItickAuthorityRegistryIdPrefix, row.Id),
		fmt.Sprintf("%s%v", cacheTItickAuthorityRegistryAuthorityPrefix, row.Authority),
	)
	return err == nil, err
}
