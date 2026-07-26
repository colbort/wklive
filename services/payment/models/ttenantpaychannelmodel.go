package models

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"wklive/common/sqlutil"
)

var _ TTenantPayChannelModel = (*customTTenantPayChannelModel)(nil)

type (
	TenantPayChannelPageFilter struct {
		TenantId   int64
		PlatformId int64
		ProductId  int64
		AccountId  int64
		Keyword    string
		Enabled    int64
	}

	// TTenantPayChannelModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTTenantPayChannelModel.
	TTenantPayChannelModel interface {
		tTenantPayChannelModel
		FindPage(ctx context.Context, filter TenantPayChannelPageFilter, cursor int64, limit int64) ([]*TTenantPayChannel, int64, error)
		FindByIDs(ctx context.Context, ids []int64) ([]*TTenantPayChannel, error)
	}

	customTTenantPayChannelModel struct {
		*defaultTTenantPayChannelModel
	}
)

// NewTTenantPayChannelModel returns a model for the database table.
func NewTTenantPayChannelModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TTenantPayChannelModel {
	return &customTTenantPayChannelModel{
		defaultTTenantPayChannelModel: newTTenantPayChannelModel(conn, c, opts...),
	}
}

func (m *defaultTTenantPayChannelModel) FindByIDs(ctx context.Context, ids []int64) ([]*TTenantPayChannel, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	builder := sqlutil.NewPageQueryBuilder()
	builder.InInt64("id", ids)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tTenantPayChannelRows, m.table, builder.Where())
	var list []*TTenantPayChannel
	if err := m.QueryRowsNoCacheCtx(ctx, &list, query, builder.Args()...); err != nil {
		return nil, err
	}
	return list, nil
}

func (m *defaultTTenantPayChannelModel) FindPage(ctx context.Context, filter TenantPayChannelPageFilter, cursor int64, limit int64) ([]*TTenantPayChannel, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqInt64("platform_id", filter.PlatformId)
	builder.EqInt64("product_id", filter.ProductId)
	builder.EqInt64("account_id", filter.AccountId)
	if filter.Keyword != "" {
		builder.And("(channel_code LIKE ? OR channel_name LIKE ? OR display_name LIKE ?)", "%"+filter.Keyword+"%", "%"+filter.Keyword+"%", "%"+filter.Keyword+"%")
	}
	builder.EqInt64("enabled", filter.Enabled)

	where := builder.Where()
	args := builder.Args()

	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSQL, args...); err != nil {
		return nil, 0, err
	}

	listArgs := append([]any{}, args...)
	listSQL := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tTenantPayChannelRows, m.table, where)
	if cursor > 0 {
		listSQL += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	listSQL += " ORDER BY sort DESC, id DESC LIMIT ?"
	listArgs = append(listArgs, limit)

	var list []*TTenantPayChannel
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSQL, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
