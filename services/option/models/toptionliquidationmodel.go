package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"wklive/common/sqlutil"
	"wklive/proto/option"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func (m *customTOptionLiquidationModel) Insert(
	ctx context.Context, data *TOptionLiquidation,
) (sql.Result, error) {
	if data == nil {
		return nil, errors.New("option liquidation is required")
	}
	if data.LiquidationScope == int64(option.LiquidationScope_LIQUIDATION_SCOPE_UNKNOWN) {
		data.LiquidationScope = int64(option.LiquidationScope_LIQUIDATION_SCOPE_ISOLATED_POSITION)
	}
	if err := validateLiquidationEvidence(data); err != nil {
		return nil, err
	}
	return m.defaultTOptionLiquidationModel.Insert(ctx, data)
}

func (m *customTOptionLiquidationModel) Update(ctx context.Context, data *TOptionLiquidation) error {
	if data == nil {
		return errors.New("option liquidation is required")
	}
	current, err := m.defaultTOptionLiquidationModel.FindOne(ctx, data.Id)
	if err != nil {
		return err
	}
	if current.TenantId != data.TenantId || current.LiquidationNo != data.LiquidationNo ||
		current.UserId != data.UserId || current.AccountId != data.AccountId ||
		current.ContractId != data.ContractId || current.PositionId != data.PositionId ||
		!current.Quantity.Equal(data.Quantity) || !current.MarkPrice.Equal(data.MarkPrice) ||
		current.LiquidationScope != data.LiquidationScope ||
		current.PortfolioRiskConfigId != data.PortfolioRiskConfigId ||
		current.PortfolioRiskConfigVersion != data.PortfolioRiskConfigVersion ||
		!current.PortfolioMaintenanceBefore.Equal(data.PortfolioMaintenanceBefore) ||
		!current.PortfolioMaintenanceAfter.Equal(data.PortfolioMaintenanceAfter) ||
		!current.PortfolioInitialAfter.Equal(data.PortfolioInitialAfter) ||
		!current.PortfolioCollateralBefore.Equal(data.PortfolioCollateralBefore) ||
		!current.PortfolioCollateralAfter.Equal(data.PortfolioCollateralAfter) {
		return errors.New("option liquidation identity and risk evidence are immutable")
	}
	if err := validateLiquidationEvidence(data); err != nil {
		return err
	}
	return m.defaultTOptionLiquidationModel.Update(ctx, data)
}

func validateLiquidationEvidence(data *TOptionLiquidation) error {
	switch option.LiquidationScope(data.LiquidationScope) {
	case option.LiquidationScope_LIQUIDATION_SCOPE_ISOLATED_POSITION:
		if data.PortfolioRiskConfigId != 0 || data.PortfolioRiskConfigVersion != 0 ||
			!data.PortfolioMaintenanceBefore.IsZero() || !data.PortfolioMaintenanceAfter.IsZero() ||
			!data.PortfolioInitialAfter.IsZero() || !data.PortfolioCollateralBefore.IsZero() ||
			!data.PortfolioCollateralAfter.IsZero() {
			return errors.New("isolated liquidation cannot contain portfolio evidence")
		}
	case option.LiquidationScope_LIQUIDATION_SCOPE_PORTFOLIO_WALLET:
		if data.AccountId != 0 || data.PortfolioRiskConfigId <= 0 || data.PortfolioRiskConfigVersion <= 0 ||
			!data.PortfolioMaintenanceBefore.GreaterThan(data.PortfolioMaintenanceAfter) ||
			data.PortfolioMaintenanceAfter.IsNegative() || data.PortfolioInitialAfter.IsNegative() ||
			data.PortfolioCollateralBefore.LessThan(data.PortfolioCollateralAfter) ||
			data.PortfolioCollateralAfter.LessThan(data.PortfolioInitialAfter) {
			return errors.New("invalid portfolio liquidation risk and collateral evidence")
		}
	default:
		return errors.New("invalid option liquidation scope")
	}
	return nil
}

var _ TOptionLiquidationModel = (*customTOptionLiquidationModel)(nil)

type (
	// TOptionLiquidationModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionLiquidationModel.
	TOptionLiquidationModel interface {
		tOptionLiquidationModel
		FindOpenByPosition(ctx context.Context, tenantId, positionId int64) (*TOptionLiquidation, error)
		FindOpenByWallet(ctx context.Context, tenantId, userId int64, settleCoin string) (*TOptionLiquidation, error)
		FindOpenPortfolioByWallet(ctx context.Context, tenantId, userId int64, settleCoin string) (*TOptionLiquidation, error)
		FindRunnable(ctx context.Context, tenantId int64, limit int64) ([]*TOptionLiquidation, error)
		FindOneForUpdate(ctx context.Context, id int64) (*TOptionLiquidation, error)
		Claim(ctx context.Context, id, now int64) (bool, error)
		FindPage(ctx context.Context, filter OptionLiquidationPageFilter, cursor, limit int64) ([]*TOptionLiquidation, int64, error)
		ResetForManualRetry(ctx context.Context, id, now int64) (bool, error)
		CancelStalePortfolio(ctx context.Context, id, now int64, reason string) (bool, error)
	}

	customTOptionLiquidationModel struct {
		*defaultTOptionLiquidationModel
	}
)

func (m *defaultTOptionLiquidationModel) CancelStalePortfolio(
	ctx context.Context, id, now int64, reason string,
) (bool, error) {
	current, err := m.FindOne(ctx, id)
	if err != nil {
		return false, err
	}
	idKey := fmt.Sprintf("%s%v", cacheTOptionLiquidationIdPrefix, id)
	identityKey := fmt.Sprintf("%s%v:%v",
		cacheTOptionLiquidationTenantIdLiquidationNoPrefix,
		current.TenantId, current.LiquidationNo,
	)
	result, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		return conn.ExecCtx(ctx, `UPDATE t_option_liquidation
SET status = ?, last_error_msg = ?, update_times = ?
WHERE id = ? AND liquidation_scope = ? AND status IN (?, ?, ?)`,
			int64(option.LiquidationStatus_LIQUIDATION_STATUS_CANCELED), reason, now, id,
			int64(option.LiquidationScope_LIQUIDATION_SCOPE_PORTFOLIO_WALLET),
			int64(option.LiquidationStatus_LIQUIDATION_STATUS_PENDING),
			int64(option.LiquidationStatus_LIQUIDATION_STATUS_EXECUTING),
			int64(option.LiquidationStatus_LIQUIDATION_STATUS_FAILED),
		)
	}, idKey, identityKey)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	if rows != 1 {
		return false, fmt.Errorf(
			"cancel stale portfolio liquidation %d affected %d rows (status=%d scope=%d)",
			id, rows, current.Status, current.LiquidationScope,
		)
	}
	return true, nil
}

type OptionLiquidationPageFilter struct {
	TenantId   int64
	UserId     int64
	AccountId  int64
	ContractId int64
	PositionId int64
	Status     int64
}

func (m *defaultTOptionLiquidationModel) ResetForManualRetry(ctx context.Context, id, now int64) (bool, error) {
	result, err := m.ExecNoCacheCtx(ctx, `UPDATE t_option_liquidation
SET status = ?, retry_count = 0, insurance_attempt = insurance_attempt + 1,
    remaining_deficit = 0, last_error_msg = '', update_times = ?
WHERE id = ? AND status IN (?, ?)`,
		int64(option.LiquidationStatus_LIQUIDATION_STATUS_PENDING), now, id,
		int64(option.LiquidationStatus_LIQUIDATION_STATUS_FAILED),
		int64(option.LiquidationStatus_LIQUIDATION_STATUS_MANUAL_REVIEW),
	)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	return rows == 1, err
}

func (m *defaultTOptionLiquidationModel) FindPage(ctx context.Context, filter OptionLiquidationPageFilter, cursor, limit int64) ([]*TOptionLiquidation, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqInt64("user_id", filter.UserId)
	builder.EqInt64("account_id", filter.AccountId)
	builder.EqInt64("contract_id", filter.ContractId)
	builder.EqInt64("position_id", filter.PositionId)
	builder.EqInt64("status", filter.Status)
	where, args := builder.Where(), builder.Args()
	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where), args...); err != nil {
		return nil, 0, err
	}
	listArgs := append([]any{}, args...)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tOptionLiquidationRows, m.table, where)
	if cursor > 0 {
		query += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	query += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)
	var list []*TOptionLiquidation
	err := m.QueryRowsNoCacheCtx(ctx, &list, query, listArgs...)
	return list, total, err
}

func (m *defaultTOptionLiquidationModel) FindRunnable(ctx context.Context, tenantId int64, limit int64) ([]*TOptionLiquidation, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	query := fmt.Sprintf("SELECT %s FROM %s WHERE status IN (?, ?, ?)", tOptionLiquidationRows, m.table)
	args := []any{
		int64(option.LiquidationStatus_LIQUIDATION_STATUS_PENDING),
		int64(option.LiquidationStatus_LIQUIDATION_STATUS_EXECUTING),
		int64(option.LiquidationStatus_LIQUIDATION_STATUS_FAILED),
	}
	if tenantId > 0 {
		query += " AND tenant_id = ?"
		args = append(args, tenantId)
	}
	query += " ORDER BY id LIMIT ?"
	args = append(args, limit)
	var list []*TOptionLiquidation
	err := m.QueryRowsNoCacheCtx(ctx, &list, query, args...)
	return list, err
}

func (m *defaultTOptionLiquidationModel) FindOneForUpdate(ctx context.Context, id int64) (*TOptionLiquidation, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE id = ? LIMIT 1 FOR UPDATE", tOptionLiquidationRows, m.table)
	var item TOptionLiquidation
	if err := m.QueryRowNoCacheCtx(ctx, &item, query, id); err != nil {
		return nil, err
	}
	return &item, nil
}

func (m *defaultTOptionLiquidationModel) Claim(ctx context.Context, id, now int64) (bool, error) {
	result, err := m.ExecNoCacheCtx(ctx, `UPDATE t_option_liquidation
SET status = ?, update_times = ?
WHERE id = ? AND status IN (?, ?)`,
		int64(option.LiquidationStatus_LIQUIDATION_STATUS_EXECUTING), now, id,
		int64(option.LiquidationStatus_LIQUIDATION_STATUS_PENDING),
		int64(option.LiquidationStatus_LIQUIDATION_STATUS_FAILED),
	)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	return rows == 1, err
}

func (m *defaultTOptionLiquidationModel) FindOpenByPosition(ctx context.Context, tenantId, positionId int64) (*TOptionLiquidation, error) {
	query := fmt.Sprintf(`SELECT %s FROM %s
WHERE tenant_id = ? AND position_id = ? AND status IN (?, ?, ?, ?)
ORDER BY id DESC LIMIT 1`, tOptionLiquidationRows, m.table)
	var item TOptionLiquidation
	err := m.QueryRowNoCacheCtx(ctx, &item, query, tenantId, positionId,
		int64(option.LiquidationStatus_LIQUIDATION_STATUS_PENDING),
		int64(option.LiquidationStatus_LIQUIDATION_STATUS_EXECUTING),
		int64(option.LiquidationStatus_LIQUIDATION_STATUS_FAILED),
		int64(option.LiquidationStatus_LIQUIDATION_STATUS_MANUAL_REVIEW),
	)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (m *defaultTOptionLiquidationModel) FindOpenByWallet(
	ctx context.Context, tenantId, userId int64, settleCoin string,
) (*TOptionLiquidation, error) {
	query := fmt.Sprintf(`SELECT l.* FROM %s l
JOIN t_option_contract c ON c.tenant_id = l.tenant_id AND c.id = l.contract_id
WHERE l.tenant_id = ? AND l.user_id = ? AND c.settle_coin = ?
  AND l.status IN (?, ?, ?, ?)
ORDER BY l.id DESC LIMIT 1`, m.table)
	var item TOptionLiquidation
	err := m.QueryRowNoCacheCtx(ctx, &item, query,
		tenantId, userId, settleCoin,
		int64(option.LiquidationStatus_LIQUIDATION_STATUS_PENDING),
		int64(option.LiquidationStatus_LIQUIDATION_STATUS_EXECUTING),
		int64(option.LiquidationStatus_LIQUIDATION_STATUS_FAILED),
		int64(option.LiquidationStatus_LIQUIDATION_STATUS_MANUAL_REVIEW),
	)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (m *defaultTOptionLiquidationModel) FindOpenPortfolioByWallet(
	ctx context.Context, tenantId, userId int64, settleCoin string,
) (*TOptionLiquidation, error) {
	query := fmt.Sprintf(`SELECT l.* FROM %s l
JOIN t_option_contract c ON c.tenant_id = l.tenant_id AND c.id = l.contract_id
WHERE l.tenant_id = ? AND l.user_id = ? AND c.settle_coin = ?
  AND l.liquidation_scope = ? AND l.status IN (?, ?, ?, ?)
ORDER BY l.id DESC LIMIT 1`, m.table)
	var item TOptionLiquidation
	err := m.QueryRowNoCacheCtx(ctx, &item, query,
		tenantId, userId, settleCoin,
		int64(option.LiquidationScope_LIQUIDATION_SCOPE_PORTFOLIO_WALLET),
		int64(option.LiquidationStatus_LIQUIDATION_STATUS_PENDING),
		int64(option.LiquidationStatus_LIQUIDATION_STATUS_EXECUTING),
		int64(option.LiquidationStatus_LIQUIDATION_STATUS_FAILED),
		int64(option.LiquidationStatus_LIQUIDATION_STATUS_MANUAL_REVIEW),
	)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// NewTOptionLiquidationModel returns a model for the database table.
func NewTOptionLiquidationModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionLiquidationModel {
	return &customTOptionLiquidationModel{
		defaultTOptionLiquidationModel: newTOptionLiquidationModel(conn, c, opts...),
	}
}
