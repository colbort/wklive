package models

import (
	"context"
	"fmt"
	"wklive/common/sqlutil"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TChatUserModel = (*customTChatUserModel)(nil)

type (
	ChatUserPageFilter struct {
		Keyword    string
		MerchantId int64
		UserType   int64
		IsOwner    int64
		Username   string
		Nickname   string
		Mobile     string
		Email      string
		Enabled    int64
	}

	// TChatUserModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTChatUserModel.
	TChatUserModel interface {
		tChatUserModel
		FindPage(ctx context.Context, filter ChatUserPageFilter, cursor int64, limit int64) ([]*TChatUser, int64, error)
		FindOneByUsername(ctx context.Context, username string) (*TChatUser, error)
	}

	customTChatUserModel struct {
		*defaultTChatUserModel
	}
)

// NewTChatUserModel returns a model for the database table.
func NewTChatUserModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TChatUserModel {
	return &customTChatUserModel{
		defaultTChatUserModel: newTChatUserModel(conn, c, opts...),
	}
}

func (m *customTChatUserModel) FindPage(ctx context.Context, filter ChatUserPageFilter, cursor int64, limit int64) ([]*TChatUser, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	if filter.Keyword != "" {
		like := "%" + filter.Keyword + "%"
		builder.And("(username LIKE ? OR nickname LIKE ? OR mobile LIKE ? OR email LIKE ?)", like, like, like, like)
	}
	builder.EqInt64("merchant_id", filter.MerchantId)
	builder.EqInt64("user_type", filter.UserType)
	builder.EqInt64("is_owner", filter.IsOwner)
	builder.LikeString("username", filter.Username)
	builder.LikeString("nickname", filter.Nickname)
	builder.EqString("mobile", filter.Mobile)
	builder.EqString("email", filter.Email)
	builder.EqInt64("enabled", filter.Enabled)

	where := builder.Where()
	args := builder.Args()

	var total int64
	countSql := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSql, args...); err != nil {
		return nil, 0, err
	}

	listSql := fmt.Sprintf("SELECT %s FROM %s WHERE %s ORDER BY id DESC LIMIT ?,?", tChatUserRows, m.table, where)
	listArgs := append(append([]any{}, args...), cursor, limit)

	var list []*TChatUser
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (m *customTChatUserModel) FindOneByUsername(ctx context.Context, username string) (*TChatUser, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE username = ? LIMIT 2", tChatUserRows, m.table)
	var list []*TChatUser
	if err := m.QueryRowsNoCacheCtx(ctx, &list, query, username); err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, ErrNotFound
	}
	if len(list) > 1 {
		return nil, fmt.Errorf("multiple chat users found for username: %s", username)
	}
	return list[0], nil
}
