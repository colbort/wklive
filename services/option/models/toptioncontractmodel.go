package models

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"wklive/common/sqlutil"
	"wklive/proto/common"
	"wklive/proto/option"
)

var _ TOptionContractModel = (*customTOptionContractModel)(nil)

type (
	OptionContractPageFilter struct {
		TenantId         int64
		ContractCode     string
		UnderlyingSymbol string
		OptionType       int64
		Status           int64
		ListTimeStart    int64
		ListTimeEnd      int64
		ExpireTimeStart  int64
		ExpireTimeEnd    int64
	}

	// TOptionContractModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionContractModel.
	TOptionContractModel interface {
		tOptionContractModel
		FindPage(ctx context.Context, filter OptionContractPageFilter, cursor int64, limit int64) ([]*TOptionContract, int64, error)
		FindOneForUpdate(ctx context.Context, id int64) (*TOptionContract, error)
		FindOneForPublicMarket(ctx context.Context, tenantId, contractId int64) (*TOptionContract, error)
		FindOptionChain(ctx context.Context, tenantId int64, underlyingSymbol string, expireTime, status, limit int64) ([]*TOptionContract, error)
	}

	customTOptionContractModel struct {
		*defaultTOptionContractModel
	}
)

func (m *defaultTOptionContractModel) FindOneForPublicMarket(
	ctx context.Context, tenantId, contractId int64,
) (*TOptionContract, error) {
	query := fmt.Sprintf(`SELECT %s FROM %s
WHERE tenant_id=? AND id=? AND is_deleted=? AND status IN (?,?) LIMIT 1`,
		tOptionContractRows, m.table)
	var item TOptionContract
	err := m.QueryRowNoCacheCtx(
		ctx, &item, query,
		tenantId, contractId, int64(common.YesNo_YES_NO_NO),
		int64(option.ContractStatus_CONTRACT_STATUS_TRADING),
		int64(option.ContractStatus_CONTRACT_STATUS_PAUSED),
	)
	return &item, err
}

func (m *defaultTOptionContractModel) FindOptionChain(
	ctx context.Context,
	tenantId int64,
	underlyingSymbol string,
	expireTime, status, limit int64,
) ([]*TOptionContract, error) {
	if limit <= 0 {
		limit = 501
	}
	query := fmt.Sprintf(`SELECT %s FROM %s
WHERE tenant_id=? AND underlying_symbol=? AND expire_time=? AND status=?
  AND is_deleted=? AND option_type IN (?,?)
ORDER BY strike_price ASC, option_type ASC, id ASC LIMIT ?`,
		tOptionContractRows, m.table)
	var items []*TOptionContract
	err := m.QueryRowsNoCacheCtx(
		ctx, &items, query,
		tenantId, strings.TrimSpace(underlyingSymbol), expireTime, status,
		int64(common.YesNo_YES_NO_NO),
		int64(option.OptionType_OPTION_TYPE_CALL),
		int64(option.OptionType_OPTION_TYPE_PUT),
		limit,
	)
	return items, err
}

func (m *defaultTOptionContractModel) FindOneForUpdate(ctx context.Context, id int64) (*TOptionContract, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE id = ? LIMIT 1 FOR UPDATE", tOptionContractRows, m.table)
	var item TOptionContract
	if err := m.QueryRowNoCacheCtx(ctx, &item, query, id); err != nil {
		return nil, err
	}
	return &item, nil
}

// NewTOptionContractModel returns a model for the database table.
func NewTOptionContractModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionContractModel {
	return &customTOptionContractModel{
		defaultTOptionContractModel: newTOptionContractModel(conn, c, opts...),
	}
}

func (m *defaultTOptionContractModel) FindPage(ctx context.Context, filter OptionContractPageFilter, cursor int64, limit int64) ([]*TOptionContract, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqString("contract_code", filter.ContractCode)
	builder.EqString("underlying_symbol", filter.UnderlyingSymbol)
	builder.EqInt64("option_type", filter.OptionType)
	builder.EqInt64("status", filter.Status)
	builder.GteInt64("list_time", filter.ListTimeStart)
	builder.LteInt64("list_time", filter.ListTimeEnd)
	builder.GteInt64("expire_time", filter.ExpireTimeStart)
	builder.LteInt64("expire_time", filter.ExpireTimeEnd)

	where := builder.Where()
	args := builder.Args()

	var total int64
	countSql := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSql, args...); err != nil {
		return nil, 0, err
	}

	listArgs := append([]any{}, args...)
	listSql := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tOptionContractRows, m.table, where)
	if cursor > 0 {
		listSql += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	listSql += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)

	var list []*TOptionContract
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
