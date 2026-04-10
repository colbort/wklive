package models

import (
	"context"
	"fmt"
)

type AssetFlowModel interface {
	tAssetFlowModel
	FindPage(ctx context.Context, tenantId int64, userId int64, walletType int64, coin string, bizType string, sceneType string, bizNo string, startTime int64, endTime int64, cursor int64, limit int64) ([]*TAssetFlow, int64, error)
}

func (m *defaultTAssetFlowModel) FindPage(ctx context.Context, tenantId int64, userId int64, walletType int64, coin string, bizType string, sceneType string, bizNo string, startTime int64, endTime int64, cursor int64, limit int64) ([]*TAssetFlow, int64, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	where := "1=1"
	args := make([]any, 0, 10)
	if tenantId > 0 {
		where += " AND tenant_id = ?"
		args = append(args, tenantId)
	}
	if userId > 0 {
		where += " AND user_id = ?"
		args = append(args, userId)
	}
	if walletType > 0 {
		where += " AND wallet_type = ?"
		args = append(args, walletType)
	}
	if coin != "" {
		where += " AND coin = ?"
		args = append(args, coin)
	}
	if bizType != "" {
		where += " AND biz_type = ?"
		args = append(args, bizType)
	}
	if sceneType != "" {
		where += " AND scene_type = ?"
		args = append(args, sceneType)
	}
	if bizNo != "" {
		where += " AND biz_no = ?"
		args = append(args, bizNo)
	}
	if startTime > 0 {
		where += " AND create_times >= ?"
		args = append(args, startTime)
	}
	if endTime > 0 {
		where += " AND create_times <= ?"
		args = append(args, endTime)
	}

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
