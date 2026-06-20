package models

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"wklive/common/sqlutil"
)

var _ TAssetFlowModel = (*customTAssetFlowModel)(nil)

type (
	AssetFlowPageFilter struct {
		TenantId   int64
		UserId     int64
		WalletType int64
		Coin       string
		BizType    string
		SceneType  string
		BizNo      string
		StartTime  int64
		EndTime    int64
	}

	// TAssetFlowModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTAssetFlowModel.
	TAssetFlowModel interface {
		tAssetFlowModel
		FindPage(ctx context.Context, filter AssetFlowPageFilter, cursor int64, limit int64) ([]*TAssetFlow, int64, error)
	}

	customTAssetFlowModel struct {
		*defaultTAssetFlowModel
	}
)

// NewTAssetFlowModel returns a model for the database table.
func NewTAssetFlowModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TAssetFlowModel {
	return &customTAssetFlowModel{
		defaultTAssetFlowModel: newTAssetFlowModel(conn, c, opts...),
	}
}

func (m *defaultTAssetFlowModel) FindPage(ctx context.Context, filter AssetFlowPageFilter, cursor int64, limit int64) ([]*TAssetFlow, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqInt64("user_id", filter.UserId)
	builder.EqInt64("wallet_type", filter.WalletType)
	builder.EqString("coin", filter.Coin)
	builder.EqString("biz_type", filter.BizType)
	builder.EqString("scene_type", filter.SceneType)
	builder.EqString("biz_no", filter.BizNo)
	builder.GteInt64("create_times", filter.StartTime)
	builder.LteInt64("create_times", filter.EndTime)

	where := builder.Where()
	args := builder.Args()

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
			tAssetFlowRows, m.table, where,
		)
		listArgs = append(listArgs, limit)
	} else {
		listSql = fmt.Sprintf(
			`SELECT %s
            FROM %s
            WHERE %s AND id < ?
            ORDER BY id DESC
            LIMIT ?`,
			tAssetFlowRows, m.table, where,
		)
		listArgs = append(listArgs, cursor, limit)
	}

	var list []*TAssetFlow
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
