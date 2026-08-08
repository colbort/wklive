package models

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TContractAdlExecutionModel = (*customTContractAdlExecutionModel)(nil)

type (
	// TContractAdlExecutionModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTContractAdlExecutionModel.
	TContractAdlExecutionModel interface {
		tContractAdlExecutionModel
		FindByLiquidation(context.Context, int64, int64) ([]*TContractAdlExecution, error)
		FindOneByExecutionNo(context.Context, int64, string) (*TContractAdlExecution, error)
		FindOneForUpdate(context.Context, int64) (*TContractAdlExecution, error)
		FindRecoverable(context.Context, int64) ([]*TContractAdlExecution, error)
	}

	customTContractAdlExecutionModel struct {
		*defaultTContractAdlExecutionModel
	}
)

// NewTContractAdlExecutionModel returns a model for the database table.
func NewTContractAdlExecutionModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TContractAdlExecutionModel {
	return &customTContractAdlExecutionModel{
		defaultTContractAdlExecutionModel: newTContractAdlExecutionModel(conn, c, opts...),
	}
}

func (m *customTContractAdlExecutionModel) FindByLiquidation(ctx context.Context, t, l int64) ([]*TContractAdlExecution, error) {
	var rows []*TContractAdlExecution
	err := m.QueryRowsNoCacheCtx(ctx, &rows, "SELECT "+tContractAdlExecutionRows+" FROM t_contract_adl_execution WHERE tenant_id=? AND liquidation_id=? ORDER BY id", t, l)
	return rows, err
}
func (m *customTContractAdlExecutionModel) FindOneByExecutionNo(ctx context.Context, t int64, n string) (*TContractAdlExecution, error) {
	var r TContractAdlExecution
	err := m.QueryRowNoCacheCtx(ctx, &r, "SELECT "+tContractAdlExecutionRows+" FROM t_contract_adl_execution WHERE tenant_id=? AND execution_no=?", t, n)
	return &r, err
}
func (m *customTContractAdlExecutionModel) FindOneForUpdate(ctx context.Context, id int64) (*TContractAdlExecution, error) {
	var r TContractAdlExecution
	err := m.QueryRowNoCacheCtx(ctx, &r, "SELECT "+tContractAdlExecutionRows+" FROM t_contract_adl_execution WHERE id=? FOR UPDATE", id)
	return &r, err
}

func (m *customTContractAdlExecutionModel) FindRecoverable(ctx context.Context, limit int64) ([]*TContractAdlExecution, error) {
	var rows []*TContractAdlExecution
	err := m.QueryRowsNoCacheCtx(ctx, &rows, "SELECT "+tContractAdlExecutionRows+" FROM t_contract_adl_execution WHERE status IN (1,2,4) ORDER BY update_times,id LIMIT ?", limit)
	return rows, err
}
