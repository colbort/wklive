package models

import (
	"context"
	"fmt"
	"wklive/common/sqlutil"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TChatMerchantModel = (*customTChatMerchantModel)(nil)

type (
	ChatMerchantPageFilter struct {
		Keyword      string
		Enabled      int64
		MerchantCode string
		MerchantName string
		ContactName  string
		ContactPhone string
		ContactEmail string
	}

	// TChatMerchantModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTChatMerchantModel.
	TChatMerchantModel interface {
		tChatMerchantModel
		FindPage(ctx context.Context, filter ChatMerchantPageFilter, offset, limit int64) ([]*TChatMerchant, int64, error)
		UpdateWithUniqueCache(ctx context.Context, data *TChatMerchant) error
	}

	customTChatMerchantModel struct {
		*defaultTChatMerchantModel
	}
)

// NewTChatMerchantModel returns a model for the database table.
func NewTChatMerchantModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TChatMerchantModel {
	return &customTChatMerchantModel{
		defaultTChatMerchantModel: newTChatMerchantModel(conn, c, opts...),
	}
}

// UpdateWithUniqueCache also clears the cache key derived from the new merchant
// code. The generated Update only knows and clears the previous unique key.
func (m *customTChatMerchantModel) UpdateWithUniqueCache(ctx context.Context, data *TChatMerchant) error {
	if err := m.Update(ctx, data); err != nil {
		return err
	}
	return m.DelCacheCtx(ctx, fmt.Sprintf("%s%v", cacheTChatMerchantMerchantCodePrefix, data.MerchantCode))
}

func (m *customTChatMerchantModel) FindPage(ctx context.Context, filter ChatMerchantPageFilter, offset, limit int64) ([]*TChatMerchant, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	builder := sqlutil.NewPageQueryBuilder()
	if filter.Keyword != "" {
		like := "%" + filter.Keyword + "%"
		builder.And("(merchant_name LIKE ? OR merchant_code LIKE ? OR contact_name LIKE ? OR contact_phone LIKE ? OR contact_email LIKE ?)",
			like, like, like, like, like)
	}
	builder.EqInt64("enabled", filter.Enabled)
	builder.EqString("merchant_code", filter.MerchantCode)
	builder.EqString("merchant_name", filter.MerchantName)
	builder.EqString("contact_name", filter.ContactName)
	builder.EqString("contact_phone", filter.ContactPhone)
	builder.EqString("contact_email", filter.ContactEmail)
	where, args := builder.Where(), builder.Args()
	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where), args...); err != nil {
		return nil, 0, err
	}
	listArgs := append(append([]any{}, args...), offset, limit)
	var rows []*TChatMerchant
	if err := m.QueryRowsNoCacheCtx(ctx, &rows,
		fmt.Sprintf("SELECT %s FROM %s WHERE %s ORDER BY id DESC LIMIT ?,?", tChatMerchantRows, m.table, where),
		listArgs...); err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}
