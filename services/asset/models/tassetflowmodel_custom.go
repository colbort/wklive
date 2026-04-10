package models

import (
	"context"
	"fmt"
	"wklive/common/sqlutil"
)

type AssetFlowModel interface {
	tAssetFlowModel
	FindPage(ctx context.Context, tenantId int64, userId int64, walletType int64, coin string, bizType string, sceneType string, bizNo string, startTime int64, endTime int64, cursor int64, limit int64) ([]*TAssetFlow, int64, error)
}

func (m *defaultTAssetFlowModel) FindPage(ctx context.Context, tenantId int64, userId int64, walletType int64, coin string, bizType string, sceneType string, bizNo string, startTime int64, endTime int64, cursor int64, limit int64) ([]*TAssetFlow, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", tenantId)
	builder.EqInt64("user_id", userId)
	builder.EqInt64("wallet_type", walletType)
	builder.EqString("coin", coin)
	builder.EqString("biz_type", bizType)
	builder.EqString("scene_type", sceneType)
	builder.EqString("biz_no", bizNo)
	builder.GteInt64("create_times", startTime)
	builder.LteInt64("create_times", endTime)

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
