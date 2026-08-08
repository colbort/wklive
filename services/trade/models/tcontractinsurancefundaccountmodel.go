package models

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TContractInsuranceFundAccountModel = (*customTContractInsuranceFundAccountModel)(nil)

type (
	// TContractInsuranceFundAccountModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTContractInsuranceFundAccountModel.
	TContractInsuranceFundAccountModel interface {
		tContractInsuranceFundAccountModel
		FindEnabled(ctx context.Context, tenantID, symbolID int64, asset string) (*TContractInsuranceFundAccount, error)
		FindPage(ctx context.Context, tenantID, symbolID, status, cursor, limit int64, asset string) ([]*TContractInsuranceFundAccount, int64, error)
	}

	customTContractInsuranceFundAccountModel struct {
		*defaultTContractInsuranceFundAccountModel
	}
)

// NewTContractInsuranceFundAccountModel returns a model for the database table.
func NewTContractInsuranceFundAccountModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TContractInsuranceFundAccountModel {
	return &customTContractInsuranceFundAccountModel{
		defaultTContractInsuranceFundAccountModel: newTContractInsuranceFundAccountModel(conn, c, opts...),
	}
}

func (m *customTContractInsuranceFundAccountModel) FindPage(ctx context.Context, t, sym, status, cursor, limit int64, asset string) ([]*TContractInsuranceFundAccount, int64, error) {
	where := "tenant_id=?"
	args := []any{t}
	if sym > 0 {
		where += " AND symbol_id=?"
		args = append(args, sym)
	}
	if status > 0 {
		where += " AND status=?"
		args = append(args, status)
	}
	if asset != "" {
		where += " AND settle_asset=?"
		args = append(args, asset)
	}
	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, "SELECT COUNT(1) FROM t_contract_insurance_fund_account WHERE "+where, args...); err != nil {
		return nil, 0, err
	}
	if cursor > 0 {
		where += " AND id<?"
		args = append(args, cursor)
	}
	args = append(args, limit)
	var rows []*TContractInsuranceFundAccount
	err := m.QueryRowsNoCacheCtx(ctx, &rows, "SELECT "+tContractInsuranceFundAccountRows+" FROM t_contract_insurance_fund_account WHERE "+where+" ORDER BY id DESC LIMIT ?", args...)
	return rows, total, err
}

func (m *customTContractInsuranceFundAccountModel) FindEnabled(ctx context.Context, tenantID, symbolID int64, asset string) (*TContractInsuranceFundAccount, error) {
	var row TContractInsuranceFundAccount
	q := "SELECT " + tContractInsuranceFundAccountRows + " FROM t_contract_insurance_fund_account WHERE tenant_id=? AND settle_asset=? AND status=1 AND symbol_id IN (?,0) ORDER BY symbol_id DESC LIMIT 1"
	if err := m.QueryRowNoCacheCtx(ctx, &row, q, tenantID, asset, symbolID); err != nil {
		return nil, err
	}
	return &row, nil
}
