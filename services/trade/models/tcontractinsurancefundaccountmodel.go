package models

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type TContractInsuranceFundAccount struct {
	Id          int64  `db:"id"`
	TenantId    int64  `db:"tenant_id"`
	SymbolId    int64  `db:"symbol_id"`
	SettleAsset string `db:"settle_asset"`
	FundUserId  int64  `db:"fund_user_id"`
	WalletType  int64  `db:"wallet_type"`
	AdlEnabled  int64  `db:"adl_enabled"`
	Status      int64  `db:"status"`
	Version     int64  `db:"version"`
	CreateTimes int64  `db:"create_times"`
	UpdateTimes int64  `db:"update_times"`
}

type TContractInsuranceFundAccountModel interface {
	FindOne(ctx context.Context, id int64) (*TContractInsuranceFundAccount, error)
	FindEnabled(ctx context.Context, tenantID, symbolID int64, asset string) (*TContractInsuranceFundAccount, error)
	FindPage(ctx context.Context, tenantID, symbolID, status, cursor, limit int64, asset string) ([]*TContractInsuranceFundAccount, int64, error)
	Insert(ctx context.Context, row *TContractInsuranceFundAccount) (sql.Result, error)
	Update(ctx context.Context, row *TContractInsuranceFundAccount) error
}

func (m *contractInsuranceFundAccountModel) FindOne(ctx context.Context, id int64) (*TContractInsuranceFundAccount, error) {
	var r TContractInsuranceFundAccount
	err := m.conn.QueryRowCtx(ctx, &r, "SELECT id,tenant_id,symbol_id,settle_asset,fund_user_id,wallet_type,adl_enabled,status,version,create_times,update_times FROM t_contract_insurance_fund_account WHERE id=?", id)
	if err != nil {
		return nil, err
	}
	return &r, nil
}
func (m *contractInsuranceFundAccountModel) FindPage(ctx context.Context, t, sym, status, cursor, limit int64, asset string) ([]*TContractInsuranceFundAccount, int64, error) {
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
	if err := m.conn.QueryRowCtx(ctx, &total, "SELECT COUNT(1) FROM t_contract_insurance_fund_account WHERE "+where, args...); err != nil {
		return nil, 0, err
	}
	if cursor > 0 {
		where += " AND id<?"
		args = append(args, cursor)
	}
	args = append(args, limit)
	var rows []*TContractInsuranceFundAccount
	err := m.conn.QueryRowsCtx(ctx, &rows, "SELECT id,tenant_id,symbol_id,settle_asset,fund_user_id,wallet_type,adl_enabled,status,version,create_times,update_times FROM t_contract_insurance_fund_account WHERE "+where+" ORDER BY id DESC LIMIT ?", args...)
	return rows, total, err
}

type contractInsuranceFundAccountModel struct{ conn sqlx.SqlConn }

func NewTContractInsuranceFundAccountModel(conn sqlx.SqlConn) TContractInsuranceFundAccountModel {
	return &contractInsuranceFundAccountModel{conn: conn}
}
func (m *contractInsuranceFundAccountModel) FindEnabled(ctx context.Context, tenantID, symbolID int64, asset string) (*TContractInsuranceFundAccount, error) {
	var row TContractInsuranceFundAccount
	q := "SELECT id,tenant_id,symbol_id,settle_asset,fund_user_id,wallet_type,adl_enabled,status,version,create_times,update_times FROM t_contract_insurance_fund_account WHERE tenant_id=? AND settle_asset=? AND status=1 AND symbol_id IN (?,0) ORDER BY symbol_id DESC LIMIT 1"
	if err := m.conn.QueryRowCtx(ctx, &row, q, tenantID, asset, symbolID); err != nil {
		return nil, err
	}
	return &row, nil
}
func (m *contractInsuranceFundAccountModel) Insert(ctx context.Context, row *TContractInsuranceFundAccount) (sql.Result, error) {
	return m.conn.ExecCtx(ctx, "INSERT INTO t_contract_insurance_fund_account(tenant_id,symbol_id,settle_asset,fund_user_id,wallet_type,adl_enabled,status,version,create_times,update_times) VALUES(?,?,?,?,?,?,?,?,?,?)", row.TenantId, row.SymbolId, row.SettleAsset, row.FundUserId, row.WalletType, row.AdlEnabled, row.Status, row.Version, row.CreateTimes, row.UpdateTimes)
}
func (m *contractInsuranceFundAccountModel) Update(ctx context.Context, row *TContractInsuranceFundAccount) error {
	res, err := m.conn.ExecCtx(ctx, "UPDATE t_contract_insurance_fund_account SET symbol_id=?,settle_asset=?,fund_user_id=?,wallet_type=?,adl_enabled=?,status=?,version=version+1,update_times=? WHERE id=? AND version=?", row.SymbolId, row.SettleAsset, row.FundUserId, row.WalletType, row.AdlEnabled, row.Status, row.UpdateTimes, row.Id, row.Version)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		return fmt.Errorf("insurance fund account concurrent update")
	}
	return nil
}
