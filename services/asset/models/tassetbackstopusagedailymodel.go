package models

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TAssetBackstopUsageDailyModel = (*customTAssetBackstopUsageDailyModel)(nil)

type (
	TAssetBackstopUsageDailyModel interface {
		tAssetBackstopUsageDailyModel
		FindOneForUpdate(ctx context.Context, tenantID int64, coin, usageDay string) (*TAssetBackstopUsageDaily, error)
		AddCovered(ctx context.Context, row *TAssetBackstopUsageDaily, amount decimal.Decimal, policyID, now int64) error
	}

	customTAssetBackstopUsageDailyModel struct {
		*defaultTAssetBackstopUsageDailyModel
	}
)

func NewTAssetBackstopUsageDailyModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TAssetBackstopUsageDailyModel {
	return &customTAssetBackstopUsageDailyModel{
		defaultTAssetBackstopUsageDailyModel: newTAssetBackstopUsageDailyModel(conn, c, opts...),
	}
}

func (m *customTAssetBackstopUsageDailyModel) FindOneForUpdate(
	ctx context.Context,
	tenantID int64,
	coin string,
	usageDay string,
) (*TAssetBackstopUsageDaily, error) {
	var row TAssetBackstopUsageDaily
	query := fmt.Sprintf(`SELECT %s FROM %s
WHERE tenant_id=? AND coin=? AND usage_day=? FOR UPDATE`, tAssetBackstopUsageDailyRows, m.table)
	if err := m.QueryRowNoCacheCtx(ctx, &row, query, tenantID, coin, usageDay); err != nil {
		if err == sql.ErrNoRows || err == sqlc.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &row, nil
}

func (m *customTAssetBackstopUsageDailyModel) AddCovered(
	ctx context.Context,
	row *TAssetBackstopUsageDaily,
	amount decimal.Decimal,
	policyID int64,
	now int64,
) error {
	idKey := fmt.Sprintf("%s%v", cacheTAssetBackstopUsageDailyIdPrefix, row.Id)
	uniqueKey := fmt.Sprintf("%s%v:%v:%v", cacheTAssetBackstopUsageDailyTenantIdCoinUsageDayPrefix,
		row.TenantId, row.Coin, row.UsageDay)
	_, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		return conn.ExecCtx(ctx, `UPDATE t_asset_backstop_usage_daily
SET covered_amount=covered_amount+?,last_policy_id=?,update_times=? WHERE id=?`,
			amount, policyID, now, row.Id)
	}, idKey, uniqueKey)
	return err
}
