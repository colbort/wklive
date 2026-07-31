package models

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionCorporateActionContractModel = (*customTOptionCorporateActionContractModel)(nil)

type (
	// TOptionCorporateActionContractModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionCorporateActionContractModel.
	TOptionCorporateActionContractModel interface {
		tOptionCorporateActionContractModel
		FindByAction(ctx context.Context, tenantId, actionId int64) ([]*TOptionCorporateActionContract, error)
		FindOneForUpdate(ctx context.Context, id int64) (*TOptionCorporateActionContract, error)
		IsSuccessorBlocked(ctx context.Context, tenantId, contractId int64) (bool, error)
		IsContractMigrationActive(ctx context.Context, tenantId, contractId int64) (bool, error)
	}

	customTOptionCorporateActionContractModel struct {
		*defaultTOptionCorporateActionContractModel
	}
)

func (m *defaultTOptionCorporateActionContractModel) IsContractMigrationActive(
	ctx context.Context, tenantId, contractId int64,
) (bool, error) {
	query := fmt.Sprintf(`SELECT COUNT(1) FROM %s m
JOIN t_option_corporate_action a ON a.id=m.action_id AND a.tenant_id=m.tenant_id
WHERE m.tenant_id=? AND (m.source_contract_id=? OR m.successor_contract_id=?)
  AND a.status IN (1,2,4,6,7)`, m.table)
	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, query, tenantId, contractId, contractId); err != nil {
		return false, err
	}
	return total > 0, nil
}

func (m *defaultTOptionCorporateActionContractModel) IsSuccessorBlocked(
	ctx context.Context, tenantId, contractId int64,
) (bool, error) {
	query := fmt.Sprintf(`SELECT COUNT(1) FROM %s m
JOIN t_option_corporate_action a ON a.id=m.action_id AND a.tenant_id=m.tenant_id
WHERE m.tenant_id=? AND m.successor_contract_id=? AND a.status IN (1,2,4,6,7)`, m.table)
	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, query, tenantId, contractId); err != nil {
		return false, err
	}
	return total > 0, nil
}

func (m *defaultTOptionCorporateActionContractModel) FindByAction(
	ctx context.Context, tenantId, actionId int64,
) ([]*TOptionCorporateActionContract, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE tenant_id=? AND action_id=? ORDER BY id",
		tOptionCorporateActionContractRows, m.table)
	var items []*TOptionCorporateActionContract
	err := m.QueryRowsNoCacheCtx(ctx, &items, query, tenantId, actionId)
	return items, err
}

func (m *defaultTOptionCorporateActionContractModel) FindOneForUpdate(
	ctx context.Context, id int64,
) (*TOptionCorporateActionContract, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE id=? LIMIT 1 FOR UPDATE",
		tOptionCorporateActionContractRows, m.table)
	var item TOptionCorporateActionContract
	if err := m.QueryRowNoCacheCtx(ctx, &item, query, id); err != nil {
		return nil, err
	}
	return &item, nil
}

// NewTOptionCorporateActionContractModel returns a model for the database table.
func NewTOptionCorporateActionContractModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionCorporateActionContractModel {
	return &customTOptionCorporateActionContractModel{
		defaultTOptionCorporateActionContractModel: newTOptionCorporateActionContractModel(conn, c, opts...),
	}
}
