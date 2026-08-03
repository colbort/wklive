package models

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"wklive/common/sqlutil"
)

var _ TAssetIdempotentModel = (*customTAssetIdempotentModel)(nil)

type (
	// TAssetIdempotentModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTAssetIdempotentModel.
	TAssetIdempotentModel interface {
		tAssetIdempotentModel
		FindPage(ctx context.Context, cursor int64, limit int64) ([]*TAssetIdempotent, int64, error)
		MarkSuccess(ctx context.Context, tenantId int64, bizType, sceneType, bizNo string, updateTimes int64) error
	}

	customTAssetIdempotentModel struct {
		*defaultTAssetIdempotentModel
	}
)

// NewTAssetIdempotentModel returns a model for the database table.
func NewTAssetIdempotentModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TAssetIdempotentModel {
	return &customTAssetIdempotentModel{
		defaultTAssetIdempotentModel: newTAssetIdempotentModel(conn, c, opts...),
	}
}

// MarkSuccess avoids a cached read immediately after Insert. PrepareAssetIdempotent
// necessarily creates a negative cache entry; reading the same unique key in
// the same SQL transaction can otherwise return sql.ErrNoRows and roll back a
// valid asset mutation.
func (m *defaultTAssetIdempotentModel) MarkSuccess(ctx context.Context, tenantId int64, bizType, sceneType, bizNo string, updateTimes int64) error {
	key := fmt.Sprintf("%s%v:%v:%v:%v", cacheTAssetIdempotentTenantIdBizTypeSceneTypeBizNoPrefix, tenantId, bizType, sceneType, bizNo)
	query := fmt.Sprintf("UPDATE %s SET status=?, remark=?, update_times=? WHERE tenant_id=? AND biz_type=? AND scene_type=? AND biz_no=?", m.table)
	result, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		return conn.ExecCtx(ctx, query, 2, "success", updateTimes, tenantId, bizType, sceneType, bizNo)
	}, key)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 1 {
		return sql.ErrNoRows
	}
	return nil
}

func (m *defaultTAssetIdempotentModel) FindPage(ctx context.Context, cursor int64, limit int64) ([]*TAssetIdempotent, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	where := builder.Where()
	args := builder.Args()

	// ---- total ----
	var total int64
	countSql := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSql, args...); err != nil {
		return nil, 0, err
	}

	listArgs := append([]any{}, args...)
	var listSql string

	if cursor <= 0 {
		listSql = fmt.Sprintf(
			`SELECT %s
			FROM %s
			WHERE %s
			ORDER BY id DESC
			LIMIT ?`,
			tAssetIdempotentRows, m.table, where,
		)
		listArgs = append(listArgs, limit)
	} else {
		listSql = fmt.Sprintf(
			`SELECT %s
			FROM %s
			WHERE %s AND id < ?
			ORDER BY id DESC
			LIMIT ?`,
			tAssetIdempotentRows, m.table, where,
		)
		listArgs = append(listArgs, cursor, limit)
	}

	var list []*TAssetIdempotent
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
