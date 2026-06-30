package models

import (
	"context"
	"fmt"
	"wklive/common/sqlutil"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TChatSessionModel = (*customTChatSessionModel)(nil)

const (
	chatSessionStatusWaiting       = 1
	chatSessionStatusPendingAgent  = 4
	chatSessionStatusInternetError = 7
)

type (
	ChatSessionPageFilter struct {
		MerchantId int64
		UserId     int64
		AgentId    int64
		GroupId    int64
		Status     int64
		Priority   int64
		Category   string
		Keyword    string
		StartTime  int64
		EndTime    int64
	}

	// TChatSessionModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTChatSessionModel.
	TChatSessionModel interface {
		tChatSessionModel
		FindPage(ctx context.Context, filter ChatSessionPageFilter, cursor int64, limit int64) ([]*TChatSession, int64, error)
		FindByUser(ctx context.Context, merchantId int64, userId int64) (*TChatSession, error)
		FindOpenByUser(ctx context.Context, merchantId int64, userId int64) (*TChatSession, error)
		FindLatestByUser(ctx context.Context, merchantId int64, userId int64) (*TChatSession, error)
		FindLatestByUserSource(ctx context.Context, merchantId int64, userId int64, source int64) (*TChatSession, error)
		FindExpiredInternetError(ctx context.Context, beforeMillis int64, limit int64) ([]*TChatSession, error)
		CountWaitingPosition(ctx context.Context, session *TChatSession) (int64, int64, error)
	}

	customTChatSessionModel struct {
		*defaultTChatSessionModel
	}
)

// NewTChatSessionModel returns a model for the database table.
func NewTChatSessionModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TChatSessionModel {
	return &customTChatSessionModel{
		defaultTChatSessionModel: newTChatSessionModel(conn, c, opts...),
	}
}

func (m *customTChatSessionModel) FindPage(ctx context.Context, filter ChatSessionPageFilter, cursor int64, limit int64) ([]*TChatSession, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("merchant_id", filter.MerchantId)
	builder.EqInt64("user_id", filter.UserId)
	builder.EqInt64("group_id", filter.GroupId)
	builder.EqInt64("status", filter.Status)
	builder.EqInt64("priority", filter.Priority)
	builder.EqString("category", filter.Category)
	builder.GteInt64("create_times", filter.StartTime)
	builder.LteInt64("create_times", filter.EndTime)
	if filter.Keyword != "" {
		builder.Or(
			[]string{"session_no LIKE ?", "title LIKE ?", "last_message LIKE ?"},
			"%"+filter.Keyword+"%",
			"%"+filter.Keyword+"%",
			"%"+filter.Keyword+"%",
		)
	}
	if filter.AgentId > 0 {
		builder.And("(agent_id = ? OR (agent_id = 0 AND status IN (?, ?)))", filter.AgentId, chatSessionStatusWaiting, chatSessionStatusPendingAgent)
	}

	where := builder.Where()
	args := builder.Args()

	var total int64
	countSql := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSql, args...); err != nil {
		return nil, 0, err
	}

	listSql := fmt.Sprintf("SELECT %s FROM %s WHERE %s ORDER BY last_message_time DESC,id DESC LIMIT ?,?", tChatSessionRows, m.table, where)
	listArgs := append(append([]any{}, args...), cursor, limit)

	var list []*TChatSession
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (m *customTChatSessionModel) FindByUser(ctx context.Context, merchantId int64, userId int64) (*TChatSession, error) {
	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE merchant_id = ? AND user_id = ? ORDER BY id DESC LIMIT 1",
		tChatSessionRows,
		m.table,
	)
	var resp TChatSession
	if err := m.QueryRowNoCacheCtx(ctx, &resp, query, merchantId, userId); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (m *customTChatSessionModel) FindOpenByUser(ctx context.Context, merchantId int64, userId int64) (*TChatSession, error) {
	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE merchant_id = ? AND user_id = ? AND status <> ? ORDER BY id DESC LIMIT 1",
		tChatSessionRows,
		m.table,
	)
	var resp TChatSession
	if err := m.QueryRowNoCacheCtx(ctx, &resp, query, merchantId, userId, 5); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (m *customTChatSessionModel) FindLatestByUser(ctx context.Context, merchantId int64, userId int64) (*TChatSession, error) {
	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE merchant_id = ? AND user_id = ? ORDER BY id DESC LIMIT 1",
		tChatSessionRows,
		m.table,
	)
	var resp TChatSession
	if err := m.QueryRowNoCacheCtx(ctx, &resp, query, merchantId, userId); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (m *customTChatSessionModel) FindLatestByUserSource(ctx context.Context, merchantId int64, userId int64, source int64) (*TChatSession, error) {
	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE merchant_id = ? AND user_id = ? AND source = ? ORDER BY id DESC LIMIT 1",
		tChatSessionRows,
		m.table,
	)
	var resp TChatSession
	if err := m.QueryRowNoCacheCtx(ctx, &resp, query, merchantId, userId, source); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (m *customTChatSessionModel) FindExpiredInternetError(ctx context.Context, beforeMillis int64, limit int64) ([]*TChatSession, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE status = ? AND disconnect_time > 0 AND disconnect_time <= ? ORDER BY disconnect_time ASC LIMIT ?",
		tChatSessionRows,
		m.table,
	)
	var list []*TChatSession
	if err := m.QueryRowsNoCacheCtx(ctx, &list, query, chatSessionStatusInternetError, beforeMillis, limit); err != nil {
		return nil, err
	}
	return list, nil
}

func (m *customTChatSessionModel) CountWaitingPosition(ctx context.Context, session *TChatSession) (int64, int64, error) {
	if session == nil || session.MerchantId <= 0 || session.SessionNo == "" {
		return 0, 0, nil
	}

	const waitingWhere = "merchant_id = ? AND group_id = ? AND agent_id = 0 AND status IN (?, ?)"
	args := []any{session.MerchantId, session.GroupId, chatSessionStatusWaiting, chatSessionStatusPendingAgent}

	var total int64
	countSql := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, waitingWhere)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSql, args...); err != nil {
		return 0, 0, err
	}

	if session.AgentId != 0 || (session.Status != chatSessionStatusWaiting && session.Status != chatSessionStatusPendingAgent) {
		return 0, total, nil
	}

	var position int64
	positionSql := fmt.Sprintf(
		"SELECT COUNT(1) FROM %s WHERE %s AND (priority > ? OR (priority = ? AND (create_times < ? OR (create_times = ? AND id <= ?))))",
		m.table,
		waitingWhere,
	)
	positionArgs := append(append([]any{}, args...), session.Priority, session.Priority, session.CreateTimes, session.CreateTimes, session.Id)
	if err := m.QueryRowNoCacheCtx(ctx, &position, positionSql, positionArgs...); err != nil {
		return 0, 0, err
	}

	return position, total, nil
}
