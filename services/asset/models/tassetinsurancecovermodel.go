package models

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TAssetInsuranceCoverModel = (*customTAssetInsuranceCoverModel)(nil)

type (
	// TAssetInsuranceCoverModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTAssetInsuranceCoverModel.
	TAssetInsuranceCoverModel interface {
		tAssetInsuranceCoverModel
		FindOneByTenantLiquidationNo(context.Context, int64, string) (*TAssetInsuranceCover, error)
		FindOneForUpdate(context.Context, int64, string) (*TAssetInsuranceCover, error)
		MarkReversed(context.Context, int64, int64) error
	}

	customTAssetInsuranceCoverModel struct {
		*defaultTAssetInsuranceCoverModel
	}
)

// NewTAssetInsuranceCoverModel returns a model for the database table.
func NewTAssetInsuranceCoverModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TAssetInsuranceCoverModel {
	return &customTAssetInsuranceCoverModel{
		defaultTAssetInsuranceCoverModel: newTAssetInsuranceCoverModel(conn, c, opts...),
	}
}

func (m *defaultTAssetInsuranceCoverModel) FindOneForUpdate(ctx context.Context, t int64, n string) (*TAssetInsuranceCover, error) {
	var r TAssetInsuranceCover
	err := m.QueryRowNoCacheCtx(ctx, &r, "SELECT "+tAssetInsuranceCoverRows+" FROM t_asset_insurance_cover WHERE tenant_id=? AND liquidation_no=? FOR UPDATE", t, n)
	if err != nil {
		return nil, err
	}
	return &r, nil
}
func (m *defaultTAssetInsuranceCoverModel) MarkReversed(ctx context.Context, id, now int64) error {
	var current TAssetInsuranceCover
	if err := m.QueryRowNoCacheCtx(ctx, &current, "SELECT "+tAssetInsuranceCoverRows+" FROM t_asset_insurance_cover WHERE id=? FOR UPDATE", id); err != nil {
		return err
	}
	idKey := fmt.Sprintf("%s%v", cacheTAssetInsuranceCoverIdPrefix, id)
	uniqueKey := fmt.Sprintf("%s%v:%v", cacheTAssetInsuranceCoverTenantIdLiquidationNoPrefix, current.TenantId, current.LiquidationNo)
	res, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		return conn.ExecCtx(ctx, "UPDATE t_asset_insurance_cover SET status=2,update_times=? WHERE id=? AND status=1", now, id)
	}, idKey, uniqueKey)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		return sql.ErrNoRows
	}
	return nil
}

func (m *defaultTAssetInsuranceCoverModel) FindOneByTenantLiquidationNo(ctx context.Context, t int64, n string) (*TAssetInsuranceCover, error) {
	var r TAssetInsuranceCover
	err := m.QueryRowNoCacheCtx(ctx, &r, "SELECT "+tAssetInsuranceCoverRows+" FROM t_asset_insurance_cover WHERE tenant_id=? AND liquidation_no=?", t, n)
	if err != nil {
		return nil, err
	}
	return &r, nil
}
