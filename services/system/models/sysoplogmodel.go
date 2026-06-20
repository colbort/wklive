package models

import (
	"context"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"wklive/common/sqlutil"
)

var _ SysOpLogModel = (*customSysOpLogModel)(nil)

type (
	OpLogPageFilter struct {
		Username string
		Method   string
		Path     string
	}

	// SysOpLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customSysOpLogModel.
	SysOpLogModel interface {
		sysOpLogModel
		FindPage(ctx context.Context, filter OpLogPageFilter, cursor int64, limit int64) ([]*SysOpLog, int64, error)
	}

	customSysOpLogModel struct {
		*defaultSysOpLogModel
	}
)

// NewSysOpLogModel returns a model for the database table.
func NewSysOpLogModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) SysOpLogModel {
	return &customSysOpLogModel{
		defaultSysOpLogModel: newSysOpLogModel(conn, c, opts...),
	}
}

func (m *defaultSysOpLogModel) FindPage(
	ctx context.Context,
	filter OpLogPageFilter,
	cursor int64,
	limit int64,
) ([]*SysOpLog, int64, error) {

	limit = sqlutil.NormalizeLimit(limit)

	// ---- WHERE 条件 ----
	builder := sqlutil.NewPageQueryBuilder()
	builder.LikeString("username", filter.Username)
	builder.LikeString("method", filter.Method)
	builder.LikeString("path", filter.Path)

	where := builder.Where()
	args := builder.Args()

	// ---- total ----
	var total int64
	countSql := "SELECT COUNT(1) FROM " + m.table + " WHERE " + where
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSql, args...); err != nil {
		return nil, 0, err
	}

	// ---- list ----
	listArgs := append([]any{}, args...)
	var listSql string

	if cursor <= 0 {
		// 第一页
		listSql = "SELECT " + sysOpLogRows +
			" FROM " + m.table +
			" WHERE " + where +
			" ORDER BY id DESC LIMIT ?"
		listArgs = append(listArgs, limit)
	} else {
		// 后续页
		listSql = "SELECT " + sysOpLogRows +
			" FROM " + m.table +
			" WHERE " + where +
			" AND id < ? ORDER BY id DESC LIMIT ?"
		listArgs = append(listArgs, cursor, limit)
	}

	var list []*SysOpLog
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
