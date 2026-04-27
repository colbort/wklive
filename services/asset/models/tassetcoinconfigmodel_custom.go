package models

import (
	"context"
	"fmt"
	"wklive/common/sqlutil"
)

type AssetCoinConfigModel interface {
	tAssetCoinConfigModel
	FindPage(ctx context.Context, tenantId int64, walletType int64, coin string, symbol string, coinType int64, appVisible int64, rechargeEnabled int64, withdrawEnabled int64, transferEnabled int64, status int64, cursor int64, limit int64) ([]*TAssetCoinConfig, int64, error)
	FindVisibleByOperation(ctx context.Context, tenantId int64, walletType int64, operationType int64, coinType int64) ([]*TAssetCoinConfig, error)
}

const (
	assetCoinSwitchOff = int64(1)
	assetCoinSwitchOn  = int64(2)

	assetCoinStatusDisabled = int64(1)
	assetCoinStatusEnabled  = int64(2)

	assetCoinOperationRecharge = int64(1)
	assetCoinOperationWithdraw = int64(2)
	assetCoinOperationTransfer = int64(3)
)

func (m *defaultTAssetCoinConfigModel) FindPage(ctx context.Context, tenantId int64, walletType int64, coin string, symbol string, coinType int64, appVisible int64, rechargeEnabled int64, withdrawEnabled int64, transferEnabled int64, status int64, cursor int64, limit int64) ([]*TAssetCoinConfig, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", tenantId)
	builder.EqInt64("wallet_type", walletType)
	builder.EqString("coin", coin)
	builder.EqString("symbol", symbol)
	builder.EqInt64("coin_type", coinType)
	appendSwitchFilter(builder, "app_visible", appVisible)
	appendSwitchFilter(builder, "recharge_enabled", rechargeEnabled)
	appendSwitchFilter(builder, "withdraw_enabled", withdrawEnabled)
	appendSwitchFilter(builder, "transfer_enabled", transferEnabled)
	appendStatusFilter(builder, status)

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
			ORDER BY sort ASC, id DESC
			LIMIT ?`,
			tAssetCoinConfigRows, m.table, where,
		)
		listArgs = append(listArgs, limit)
	} else {
		listSql = fmt.Sprintf(
			`SELECT %s
			FROM %s
			WHERE %s AND id < ?
			ORDER BY sort ASC, id DESC
			LIMIT ?`,
			tAssetCoinConfigRows, m.table, where,
		)
		listArgs = append(listArgs, cursor, limit)
	}

	var list []*TAssetCoinConfig
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (m *defaultTAssetCoinConfigModel) FindVisibleByOperation(ctx context.Context, tenantId int64, walletType int64, operationType int64, coinType int64) ([]*TAssetCoinConfig, error) {
	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", tenantId)
	builder.EqInt64("wallet_type", walletType)
	builder.EqInt64("coin_type", coinType)
	builder.And("app_visible = ?", int64(1))
	builder.And("status = ?", int64(1))
	switch operationType {
	case assetCoinOperationRecharge:
		builder.And("recharge_enabled = ?", int64(1))
	case assetCoinOperationWithdraw:
		builder.And("withdraw_enabled = ?", int64(1))
	case assetCoinOperationTransfer:
		builder.And("transfer_enabled = ?", int64(1))
	}

	where := builder.Where()
	args := builder.Args()

	query := fmt.Sprintf(
		`SELECT %s
		FROM %s
		WHERE %s
		ORDER BY sort ASC, id DESC`,
		tAssetCoinConfigRows, m.table, where,
	)

	var list []*TAssetCoinConfig
	if err := m.QueryRowsNoCacheCtx(ctx, &list, query, args...); err != nil {
		return nil, err
	}

	return list, nil
}

func appendSwitchFilter(builder *sqlutil.PageQueryBuilder, column string, value int64) {
	switch value {
	case assetCoinSwitchOff:
		builder.And(fmt.Sprintf("%s = ?", column), int64(0))
	case assetCoinSwitchOn:
		builder.And(fmt.Sprintf("%s = ?", column), int64(1))
	}
}

func appendStatusFilter(builder *sqlutil.PageQueryBuilder, status int64) {
	switch status {
	case assetCoinStatusDisabled:
		builder.And("status = ?", int64(0))
	case assetCoinStatusEnabled:
		builder.And("status = ?", int64(1))
	}
}
