package models

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"strings"
)

var _ TContractMarginSnapshotModel = (*customTContractMarginSnapshotModel)(nil)

type (
	// TContractMarginSnapshotModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTContractMarginSnapshotModel.
	TContractMarginSnapshotModel interface {
		tContractMarginSnapshotModel
		FindPage(ctx context.Context, tenantId, userId int64, marginAsset string, cursor, limit int64) ([]*TContractMarginSnapshot, int64, error)
	}

	customTContractMarginSnapshotModel struct {
		*defaultTContractMarginSnapshotModel
	}
)

// NewTContractMarginSnapshotModel returns a model for the database table.
func NewTContractMarginSnapshotModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TContractMarginSnapshotModel {
	return &customTContractMarginSnapshotModel{
		defaultTContractMarginSnapshotModel: newTContractMarginSnapshotModel(conn, c, opts...),
	}
}

func (m *defaultTContractMarginSnapshotModel) FindPage(ctx context.Context, tenantId, userId int64, marginAsset string, cursor, limit int64) ([]*TContractMarginSnapshot, int64, error) {
	where, args := []string{"1=1"}, []any{}
	if tenantId > 0 {
		where = append(where, "tenant_id = ?")
		args = append(args, tenantId)
	}
	if userId > 0 {
		where = append(where, "user_id = ?")
		args = append(args, userId)
	}
	if marginAsset != "" {
		where = append(where, "margin_asset = ?")
		args = append(args, marginAsset)
	}
	clause := strings.Join(where, " and ")
	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, fmt.Sprintf("select count(*) from %s where %s", m.table, clause), args...); err != nil {
		return nil, 0, err
	}
	queryArgs := append(append([]any{}, args...), cursor, limit)
	var rows []*TContractMarginSnapshot
	err := m.QueryRowsNoCacheCtx(ctx, &rows, fmt.Sprintf("select %s from %s where %s and id > ? order by id asc limit ?", tContractMarginSnapshotRows, m.table, clause), queryArgs...)
	return rows, total, err
}
