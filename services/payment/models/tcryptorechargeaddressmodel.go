package models

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"wklive/common/sqlutil"
)

var _ TCryptoRechargeAddressModel = (*customTCryptoRechargeAddressModel)(nil)

type (
	CryptoRechargeAddressPageFilter struct {
		TenantId    int64
		UserId      int64
		WalletType  int64
		Coin        string
		ChainCode   int64
		Address     string
		Status      int64
		AddressType int64
	}

	// TCryptoRechargeAddressModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTCryptoRechargeAddressModel.
	TCryptoRechargeAddressModel interface {
		tCryptoRechargeAddressModel
		FindPage(ctx context.Context, filter CryptoRechargeAddressPageFilter, cursor int64, limit int64) ([]*TCryptoRechargeAddress, int64, error)
		FindOneAssignable(ctx context.Context, tenantId int64, walletType int64, coin string, chainCode int64) (*TCryptoRechargeAddress, error)
		FindAssignableCandidates(ctx context.Context, tenantId int64, walletType int64, coin string, chainCode int64, reusableBefore int64, limit int64) ([]*TCryptoRechargeAddress, error)
		HasEnabledAddress(ctx context.Context, tenantId int64, walletType int64, coin string, chainCode int64) (bool, error)
	}

	customTCryptoRechargeAddressModel struct {
		*defaultTCryptoRechargeAddressModel
	}
)

// NewTCryptoRechargeAddressModel returns a model for the database table.
func NewTCryptoRechargeAddressModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TCryptoRechargeAddressModel {
	return &customTCryptoRechargeAddressModel{
		defaultTCryptoRechargeAddressModel: newTCryptoRechargeAddressModel(conn, c, opts...),
	}
}

func (m *defaultTCryptoRechargeAddressModel) FindPage(ctx context.Context, filter CryptoRechargeAddressPageFilter, cursor int64, limit int64) ([]*TCryptoRechargeAddress, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqInt64("user_id", filter.UserId)
	builder.EqInt64("wallet_type", filter.WalletType)
	builder.EqString("coin", filter.Coin)
	builder.EqInt64("chain_code", filter.ChainCode)
	builder.EqString("address", filter.Address)
	appendCryptoAddressStatusFilter(builder, filter.Status)
	builder.EqInt64("address_type", filter.AddressType)

	where := builder.Where()
	args := builder.Args()

	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSQL, args...); err != nil {
		return nil, 0, err
	}

	listArgs := append([]any{}, args...)
	listSQL := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tCryptoRechargeAddressRows, m.table, where)
	if cursor > 0 {
		listSQL += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	listSQL += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)

	var list []*TCryptoRechargeAddress
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSQL, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func appendCryptoAddressStatusFilter(builder *sqlutil.PageQueryBuilder, status int64) {
	if status != 0 {
		builder.And("status = ?", status)
	}
}

func (m *defaultTCryptoRechargeAddressModel) FindOneAssignable(ctx context.Context, tenantId int64, walletType int64, coin string, chainCode int64) (*TCryptoRechargeAddress, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM %s
		WHERE tenant_id = ? AND user_id = 0 AND wallet_type = ? AND coin = ? AND chain_code = ? AND status = 2
		ORDER BY address_type DESC, last_used_time ASC, id ASC
		LIMIT 1`,
		tCryptoRechargeAddressRows, m.table,
	)

	var item TCryptoRechargeAddress
	if err := m.QueryRowNoCacheCtx(ctx, &item, query, tenantId, walletType, coin, chainCode); err != nil {
		return nil, err
	}

	return &item, nil
}

func (m *defaultTCryptoRechargeAddressModel) FindAssignableCandidates(ctx context.Context, tenantId int64, walletType int64, coin string, chainCode int64, reusableBefore int64, limit int64) ([]*TCryptoRechargeAddress, error) {
	limit = sqlutil.NormalizeLimit(limit)
	query := fmt.Sprintf(
		`SELECT %s FROM %s
		WHERE tenant_id = ? AND wallet_type = ? AND coin = ? AND chain_code = ? AND status = 2
			AND (user_id = 0 OR last_used_time <= ?)
		ORDER BY address_type DESC, last_used_time ASC, id ASC
		LIMIT ?`,
		tCryptoRechargeAddressRows, m.table,
	)

	var list []*TCryptoRechargeAddress
	if err := m.QueryRowsNoCacheCtx(ctx, &list, query, tenantId, walletType, coin, chainCode, reusableBefore, limit); err != nil {
		return nil, err
	}

	return list, nil
}

func (m *defaultTCryptoRechargeAddressModel) HasEnabledAddress(ctx context.Context, tenantId int64, walletType int64, coin string, chainCode int64) (bool, error) {
	query := fmt.Sprintf(
		`SELECT COUNT(1) FROM %s
		WHERE tenant_id = ? AND wallet_type = ? AND coin = ? AND chain_code = ? AND status = 2`,
		m.table,
	)

	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, query, tenantId, walletType, coin, chainCode); err != nil {
		return false, err
	}

	return total > 0, nil
}
