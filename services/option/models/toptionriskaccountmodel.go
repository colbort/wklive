package models

import (
	"context"
	"database/sql"
	"fmt"

	"wklive/common/sqlutil"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionRiskAccountModel = (*customTOptionRiskAccountModel)(nil)

type (
	// TOptionRiskAccountModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionRiskAccountModel.
	TOptionRiskAccountModel interface {
		tOptionRiskAccountModel
		FindByTenant(ctx context.Context, tenantId int64) ([]*TOptionRiskAccount, error)
		FindPage(ctx context.Context, filter OptionRiskAccountPageFilter, cursor, limit int64) ([]*TOptionRiskAccount, int64, error)
		EnsureAndFindOneForUpdate(ctx context.Context, tenantId, userId, accountId int64, settleCoin string, now int64) (*TOptionRiskAccount, error)
	}

	customTOptionRiskAccountModel struct {
		*defaultTOptionRiskAccountModel
	}
)

func (m *defaultTOptionRiskAccountModel) EnsureAndFindOneForUpdate(
	ctx context.Context,
	tenantId, userId, accountId int64,
	settleCoin string,
	now int64,
) (*TOptionRiskAccount, error) {
	identityKey := fmt.Sprintf("%s%v:%v:%v:%v",
		cacheTOptionRiskAccountTenantIdUserIdAccountIdSettleCoinPrefix,
		tenantId, userId, accountId, settleCoin,
	)
	if _, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		return conn.ExecCtx(ctx, `
INSERT IGNORE INTO t_option_risk_account
(tenant_id,user_id,account_id,settle_coin,status,create_times,update_times)
VALUES (?,?,?,?,?,?,?)`,
			tenantId, userId, accountId, settleCoin, 1, now, now,
		)
	}, identityKey); err != nil {
		return nil, err
	}
	query := fmt.Sprintf(`SELECT %s FROM %s
WHERE tenant_id = ? AND user_id = ? AND account_id = ? AND settle_coin = ?
LIMIT 1 FOR UPDATE`, tOptionRiskAccountRows, m.table)
	var item TOptionRiskAccount
	if err := m.QueryRowNoCacheCtx(ctx, &item, query, tenantId, userId, accountId, settleCoin); err != nil {
		return nil, err
	}
	return &item, nil
}

type OptionRiskAccountPageFilter struct {
	TenantId   int64
	UserId     int64
	AccountId  int64
	SettleCoin string
	Status     int64
}

func (m *defaultTOptionRiskAccountModel) FindPage(ctx context.Context, filter OptionRiskAccountPageFilter, cursor, limit int64) ([]*TOptionRiskAccount, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqInt64("user_id", filter.UserId)
	builder.EqInt64("account_id", filter.AccountId)
	builder.EqString("settle_coin", filter.SettleCoin)
	builder.EqInt64("status", filter.Status)
	where, args := builder.Where(), builder.Args()
	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where), args...); err != nil {
		return nil, 0, err
	}
	listArgs := append([]any{}, args...)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tOptionRiskAccountRows, m.table, where)
	if cursor > 0 {
		query += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	query += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)
	var list []*TOptionRiskAccount
	err := m.QueryRowsNoCacheCtx(ctx, &list, query, listArgs...)
	return list, total, err
}

func (m *defaultTOptionRiskAccountModel) FindByTenant(ctx context.Context, tenantId int64) ([]*TOptionRiskAccount, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE (? = 0 OR tenant_id = ?) ORDER BY id", tOptionRiskAccountRows, m.table)
	var list []*TOptionRiskAccount
	err := m.QueryRowsNoCacheCtx(ctx, &list, query, tenantId, tenantId)
	return list, err
}

// NewTOptionRiskAccountModel returns a model for the database table.
func NewTOptionRiskAccountModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionRiskAccountModel {
	return &customTOptionRiskAccountModel{
		defaultTOptionRiskAccountModel: newTOptionRiskAccountModel(conn, c, opts...),
	}
}
