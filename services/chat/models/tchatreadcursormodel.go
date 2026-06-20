package models

import (
	"context"
	"fmt"
	"wklive/common/sqlutil"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TChatReadCursorModel = (*customTChatReadCursorModel)(nil)

type (
	// TChatReadCursorModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTChatReadCursorModel.
	TChatReadCursorModel interface {
		tChatReadCursorModel
		ListBySessionNo(ctx context.Context, merchantId int64, sessionNo string) ([]*TChatReadCursor, error)
	}

	customTChatReadCursorModel struct {
		*defaultTChatReadCursorModel
	}
)

// NewTChatReadCursorModel returns a model for the database table.
func NewTChatReadCursorModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TChatReadCursorModel {
	return &customTChatReadCursorModel{
		defaultTChatReadCursorModel: newTChatReadCursorModel(conn, c, opts...),
	}
}

func (m *customTChatReadCursorModel) ListBySessionNo(ctx context.Context, merchantId int64, sessionNo string) ([]*TChatReadCursor, error) {
	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("merchant_id", merchantId)
	builder.EqString("session_no", sessionNo)

	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s ORDER BY reader_type ASC,reader_id ASC,device_id ASC", tChatReadCursorRows, m.table, builder.Where())
	var list []*TChatReadCursor
	if err := m.QueryRowsNoCacheCtx(ctx, &list, query, builder.Args()...); err != nil {
		return nil, err
	}
	return list, nil
}
