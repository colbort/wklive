package models

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"sort"
	"wklive/common/sqlutil"
)

var _ TAssetCoinConfigModel = (*customTAssetCoinConfigModel)(nil)

type (
	AssetCoinConfigPageFilter struct {
		TenantId        int64
		WalletType      int64
		Coin            string
		Symbol          string
		CoinType        int64
		ChainCode       int64
		AppVisible      int64
		RechargeEnabled int64
		WithdrawEnabled int64
		TransferEnabled int64
		Enabled         int64
	}

	// TAssetCoinConfigModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTAssetCoinConfigModel.
	TAssetCoinConfigModel interface {
		tAssetCoinConfigModel
		FindPage(ctx context.Context, filter AssetCoinConfigPageFilter, cursor int64, limit int64) ([]*TAssetCoinConfig, int64, error)
		FindVisibleByOperation(ctx context.Context, tenantId int64, walletType int64, operationType int64, coinType int64) ([]*TAssetCoinConfig, error)
	}

	customTAssetCoinConfigModel struct {
		*defaultTAssetCoinConfigModel
	}
)

// NewTAssetCoinConfigModel returns a model for the database table.
func NewTAssetCoinConfigModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TAssetCoinConfigModel {
	return &customTAssetCoinConfigModel{
		defaultTAssetCoinConfigModel: newTAssetCoinConfigModel(conn, c, opts...),
	}
}

const (
	assetCoinOperationRecharge = int64(1)
	assetCoinOperationWithdraw = int64(2)
	assetCoinOperationTransfer = int64(3)
)

func (m *defaultTAssetCoinConfigModel) FindPage(ctx context.Context, filter AssetCoinConfigPageFilter, cursor int64, limit int64) ([]*TAssetCoinConfig, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqInt64("wallet_type", filter.WalletType)
	builder.EqString("coin", filter.Coin)
	builder.EqString("symbol", filter.Symbol)
	builder.EqInt64("coin_type", filter.CoinType)
	builder.EqInt64("chain_code", filter.ChainCode)
	appendSwitchFilter(builder, "app_visible", filter.AppVisible)
	appendSwitchFilter(builder, "recharge_enabled", filter.RechargeEnabled)
	appendSwitchFilter(builder, "withdraw_enabled", filter.WithdrawEnabled)
	appendSwitchFilter(builder, "transfer_enabled", filter.TransferEnabled)
	appendEnabledFilter(builder, filter.Enabled)

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
	builder.And("enabled = ?", int64(1))
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

	// 如果operationType <= 0, USDT，USDC 只应该返回一个，USDT有ERC20和TRC20两种类型，但APP端不区分
	if operationType <= 0 {
		usdtList := make([]*TAssetCoinConfig, 0)
		usdcList := make([]*TAssetCoinConfig, 0)
		newList := make([]*TAssetCoinConfig, 0, len(list))
		for _, item := range list {
			switch item.Coin {
			case "USDT":
				usdtList = append(usdtList, item)
			case "USDC":
				usdcList = append(usdcList, item)
			default:
				newList = append(newList, item)
			}
		}
		if len(usdtList) > 0 {
			newList = append(newList, usdtList[0])
		}
		if len(usdcList) > 0 {
			newList = append(newList, usdcList[0])
		}
		list = newList
	}

	sort.SliceStable(list, func(i, j int) bool {
		if list[i].Sort == list[j].Sort {
			return list[i].Id > list[j].Id
		}
		return list[i].Sort < list[j].Sort
	})

	return list, nil
}

func appendSwitchFilter(builder *sqlutil.PageQueryBuilder, column string, value int64) {
	if value != 0 {
		builder.And(fmt.Sprintf("%s = ?", column), value)
	}
}

func appendEnabledFilter(builder *sqlutil.PageQueryBuilder, enabled int64) {
	if enabled != 0 {
		builder.And("enabled = ?", enabled)
	}
}
