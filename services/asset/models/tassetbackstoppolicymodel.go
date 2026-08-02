package models

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TAssetBackstopPolicyModel = (*customTAssetBackstopPolicyModel)(nil)

type (
	TAssetBackstopPolicyModel interface {
		tAssetBackstopPolicyModel
		FindOneForUpdate(ctx context.Context, id int64) (*TAssetBackstopPolicy, error)
		FindOneByRequestNoForUpdate(ctx context.Context, tenantID int64, requestNo string) (*TAssetBackstopPolicy, error)
		FindEffectiveForUpdate(ctx context.Context, tenantID int64, coin string, now int64) (*TAssetBackstopPolicy, error)
		NextVersion(ctx context.Context, tenantID int64, coin string) (int64, error)
		FindPage(ctx context.Context, tenantID int64, coin string, status, cursor, limit int64) ([]*TAssetBackstopPolicy, int64, error)
	}

	customTAssetBackstopPolicyModel struct {
		*defaultTAssetBackstopPolicyModel
	}
)

func NewTAssetBackstopPolicyModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TAssetBackstopPolicyModel {
	return &customTAssetBackstopPolicyModel{
		defaultTAssetBackstopPolicyModel: newTAssetBackstopPolicyModel(conn, c, opts...),
	}
}

func (m *defaultTAssetBackstopPolicyModel) FindOneByRequestNoForUpdate(
	ctx context.Context,
	tenantID int64,
	requestNo string,
) (*TAssetBackstopPolicy, error) {
	var row TAssetBackstopPolicy
	query := fmt.Sprintf("SELECT %s FROM %s WHERE tenant_id=? AND request_no=? FOR UPDATE",
		tAssetBackstopPolicyRows, m.table)
	if err := m.QueryRowNoCacheCtx(ctx, &row, query, tenantID, requestNo); err != nil {
		if err == sql.ErrNoRows || err == sqlc.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &row, nil
}

func (m *defaultTAssetBackstopPolicyModel) FindPage(
	ctx context.Context,
	tenantID int64,
	coin string,
	status int64,
	cursor int64,
	limit int64,
) ([]*TAssetBackstopPolicy, int64, error) {
	if limit <= 0 || limit > 100 {
		limit = 100
	}
	where := "tenant_id=?"
	args := []any{tenantID}
	if coin != "" {
		where += " AND coin=?"
		args = append(args, coin)
	}
	if status != 0 {
		where += " AND status=?"
		args = append(args, status)
	}
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}
	listArgs := append([]any{}, args...)
	if cursor > 0 {
		where += " AND id<?"
		listArgs = append(listArgs, cursor)
	}
	listArgs = append(listArgs, limit)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s ORDER BY id DESC LIMIT ?",
		tAssetBackstopPolicyRows, m.table, where)
	var rows []*TAssetBackstopPolicy
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, listArgs...); err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (m *defaultTAssetBackstopPolicyModel) FindOneForUpdate(
	ctx context.Context,
	id int64,
) (*TAssetBackstopPolicy, error) {
	var row TAssetBackstopPolicy
	query := fmt.Sprintf("SELECT %s FROM %s WHERE id=? FOR UPDATE", tAssetBackstopPolicyRows, m.table)
	if err := m.QueryRowNoCacheCtx(ctx, &row, query, id); err != nil {
		if err == sql.ErrNoRows || err == sqlc.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &row, nil
}

func (m *defaultTAssetBackstopPolicyModel) FindEffectiveForUpdate(
	ctx context.Context,
	tenantID int64,
	coin string,
	now int64,
) (*TAssetBackstopPolicy, error) {
	var row TAssetBackstopPolicy
	query := fmt.Sprintf(`SELECT %s FROM %s
WHERE tenant_id=? AND coin=? AND status=2
  AND effective_from<=? AND effective_until>?
ORDER BY version DESC LIMIT 1 FOR UPDATE`, tAssetBackstopPolicyRows, m.table)
	if err := m.QueryRowNoCacheCtx(ctx, &row, query, tenantID, coin, now, now); err != nil {
		if err == sql.ErrNoRows || err == sqlc.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &row, nil
}

func (m *defaultTAssetBackstopPolicyModel) NextVersion(
	ctx context.Context,
	tenantID int64,
	coin string,
) (int64, error) {
	var next int64
	query := fmt.Sprintf("SELECT COALESCE(MAX(version),0)+1 FROM %s WHERE tenant_id=? AND coin=?", m.table)
	if err := m.QueryRowNoCacheCtx(ctx, &next, query, tenantID, coin); err != nil {
		return 0, err
	}
	return next, nil
}
