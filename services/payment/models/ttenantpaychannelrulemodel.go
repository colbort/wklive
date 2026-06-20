package models

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"wklive/common/sqlutil"
)

var _ TTenantPayChannelRuleModel = (*customTTenantPayChannelRuleModel)(nil)

type (
	// TTenantPayChannelRuleModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTTenantPayChannelRuleModel.
	TTenantPayChannelRuleModel interface {
		tTenantPayChannelRuleModel
		FindPage(ctx context.Context, channelId int64, cursor int64, limit int64) ([]*TTenantPayChannelRule, int64, error)
	}

	customTTenantPayChannelRuleModel struct {
		*defaultTTenantPayChannelRuleModel
	}
)

// NewTTenantPayChannelRuleModel returns a model for the database table.
func NewTTenantPayChannelRuleModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TTenantPayChannelRuleModel {
	return &customTTenantPayChannelRuleModel{
		defaultTTenantPayChannelRuleModel: newTTenantPayChannelRuleModel(conn, c, opts...),
	}
}

func (m *defaultTTenantPayChannelRuleModel) FindPage(ctx context.Context, channelId int64, cursor int64, limit int64) ([]*TTenantPayChannelRule, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("channel_id", channelId)

	where := builder.Where()
	args := builder.Args()

	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSQL, args...); err != nil {
		return nil, 0, err
	}

	listArgs := append([]any{}, args...)
	listSQL := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tTenantPayChannelRuleRows, m.table, where)
	if cursor > 0 {
		listSQL += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	listSQL += " ORDER BY priority ASC, id DESC LIMIT ?"
	listArgs = append(listArgs, limit)

	var list []*TTenantPayChannelRule
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSQL, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
