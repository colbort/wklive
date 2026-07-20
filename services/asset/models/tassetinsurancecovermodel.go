package models

import (
	"context"
	"database/sql"
	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type TAssetInsuranceCover struct {
	Id              int64           `db:"id"`
	TenantId        int64           `db:"tenant_id"`
	FundUserId      int64           `db:"fund_user_id"`
	WalletType      int64           `db:"wallet_type"`
	Coin            string          `db:"coin"`
	LiquidationId   int64           `db:"liquidation_id"`
	LiquidationNo   string          `db:"liquidation_no"`
	RequestedAmount decimal.Decimal `db:"requested_amount"`
	CoveredAmount   decimal.Decimal `db:"covered_amount"`
	RemainingAmount decimal.Decimal `db:"remaining_amount"`
	Status          int64           `db:"status"`
	CreateTimes     int64           `db:"create_times"`
	UpdateTimes     int64           `db:"update_times"`
}
type TAssetInsuranceCoverModel interface {
	Insert(context.Context, *TAssetInsuranceCover) (sql.Result, error)
	FindOneByTenantLiquidationNo(context.Context, int64, string) (*TAssetInsuranceCover, error)
	FindOneForUpdate(context.Context, int64, string) (*TAssetInsuranceCover, error)
	MarkReversed(context.Context, int64, int64) error
}

func (m *assetInsuranceCoverModel) FindOneForUpdate(ctx context.Context, t int64, n string) (*TAssetInsuranceCover, error) {
	var r TAssetInsuranceCover
	err := m.conn.QueryRowCtx(ctx, &r, "SELECT id,tenant_id,fund_user_id,wallet_type,coin,liquidation_id,liquidation_no,requested_amount,covered_amount,remaining_amount,status,create_times,update_times FROM t_asset_insurance_cover WHERE tenant_id=? AND liquidation_no=? FOR UPDATE", t, n)
	if err != nil {
		return nil, err
	}
	return &r, nil
}
func (m *assetInsuranceCoverModel) MarkReversed(ctx context.Context, id, now int64) error {
	res, err := m.conn.ExecCtx(ctx, "UPDATE t_asset_insurance_cover SET status=2,update_times=? WHERE id=? AND status=1", now, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		return sql.ErrNoRows
	}
	return nil
}

type assetInsuranceCoverModel struct{ conn sqlx.SqlConn }

func NewTAssetInsuranceCoverModel(c sqlx.SqlConn) TAssetInsuranceCoverModel {
	return &assetInsuranceCoverModel{c}
}
func (m *assetInsuranceCoverModel) Insert(ctx context.Context, r *TAssetInsuranceCover) (sql.Result, error) {
	return m.conn.ExecCtx(ctx, "INSERT INTO t_asset_insurance_cover(tenant_id,fund_user_id,wallet_type,coin,liquidation_id,liquidation_no,requested_amount,covered_amount,remaining_amount,status,create_times,update_times) VALUES(?,?,?,?,?,?,?,?,?,?,?,?)", r.TenantId, r.FundUserId, r.WalletType, r.Coin, r.LiquidationId, r.LiquidationNo, r.RequestedAmount, r.CoveredAmount, r.RemainingAmount, r.Status, r.CreateTimes, r.UpdateTimes)
}
func (m *assetInsuranceCoverModel) FindOneByTenantLiquidationNo(ctx context.Context, t int64, n string) (*TAssetInsuranceCover, error) {
	var r TAssetInsuranceCover
	err := m.conn.QueryRowCtx(ctx, &r, "SELECT id,tenant_id,fund_user_id,wallet_type,coin,liquidation_id,liquidation_no,requested_amount,covered_amount,remaining_amount,status,create_times,update_times FROM t_asset_insurance_cover WHERE tenant_id=? AND liquidation_no=?", t, n)
	if err != nil {
		return nil, err
	}
	return &r, nil
}
