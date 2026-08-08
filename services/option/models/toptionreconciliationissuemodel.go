package models

import (
	"context"
	"fmt"

	"wklive/common/sqlutil"
	"wklive/proto/option"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionReconciliationIssueModel = (*customTOptionReconciliationIssueModel)(nil)

type (
	OptionReconciliationIssuePageFilter struct {
		TenantId  int64
		BizNo     string
		CheckType int64
		Status    int64
	}

	// TOptionReconciliationIssueModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionReconciliationIssueModel.
	TOptionReconciliationIssueModel interface {
		tOptionReconciliationIssueModel
		Open(ctx context.Context, item *TOptionReconciliationIssue) error
		Resolve(ctx context.Context, tenantId int64, issueKey string, now int64) error
		ResolveOpenByPrefix(ctx context.Context, tenantId int64, issueKeyPrefix string, now int64) error
		ResolveOpenByPrefixForCheckType(ctx context.Context, tenantId, checkType int64, issueKeyPrefix string, now int64) error
		FindPage(ctx context.Context, filter OptionReconciliationIssuePageFilter, cursor, limit int64) ([]*TOptionReconciliationIssue, int64, error)
	}

	customTOptionReconciliationIssueModel struct {
		*defaultTOptionReconciliationIssueModel
	}
)

func (m *customTOptionReconciliationIssueModel) FindPage(
	ctx context.Context, filter OptionReconciliationIssuePageFilter, cursor, limit int64,
) ([]*TOptionReconciliationIssue, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.LikeString("biz_no", filter.BizNo)
	builder.EqInt64("check_type", filter.CheckType)
	builder.EqInt64("status", filter.Status)
	where, args := builder.Where(), builder.Args()
	var total int64
	if err := m.QueryRowNoCacheCtx(
		ctx, &total, fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where), args...,
	); err != nil {
		return nil, 0, err
	}
	listArgs := append([]any{}, args...)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tOptionReconciliationIssueRows, m.table, where)
	if cursor > 0 {
		query += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	query += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)
	var items []*TOptionReconciliationIssue
	if err := m.QueryRowsNoCacheCtx(ctx, &items, query, listArgs...); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// NewTOptionReconciliationIssueModel returns a model for the database table.
func NewTOptionReconciliationIssueModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionReconciliationIssueModel {
	return &customTOptionReconciliationIssueModel{
		defaultTOptionReconciliationIssueModel: newTOptionReconciliationIssueModel(conn, c, opts...),
	}
}

func (m *customTOptionReconciliationIssueModel) Open(ctx context.Context, item *TOptionReconciliationIssue) error {
	_, err := m.ExecNoCacheCtx(ctx, `INSERT INTO t_option_reconciliation_issue
(tenant_id,issue_key,check_type,biz_no,instruction_id,expected_value,actual_value,detail,status,occurrence_count,resolved_at,create_times,update_times)
VALUES(?,?,?,?,?,?,?,?,?,1,0,?,?)
ON DUPLICATE KEY UPDATE
check_type=VALUES(check_type),biz_no=VALUES(biz_no),instruction_id=VALUES(instruction_id),
expected_value=VALUES(expected_value),actual_value=VALUES(actual_value),detail=VALUES(detail),
status=VALUES(status),occurrence_count=occurrence_count+1,resolved_at=0,update_times=VALUES(update_times)`,
		item.TenantId, item.IssueKey, item.CheckType, item.BizNo, item.InstructionId,
		item.ExpectedValue, item.ActualValue, item.Detail, item.Status, item.CreateTimes, item.UpdateTimes)
	return err
}

// OpenOptionReconciliationIssue writes an issue through the provided connection without
// constructing the cache-backed model. This keeps transactional reconciliation writes
// on the same SQL transaction and avoids making Redis availability part of correctness.
func OpenOptionReconciliationIssue(
	ctx context.Context, conn sqlx.SqlConn, item *TOptionReconciliationIssue,
) error {
	_, err := conn.ExecCtx(ctx, `INSERT INTO t_option_reconciliation_issue
(tenant_id,issue_key,check_type,biz_no,instruction_id,expected_value,actual_value,detail,status,occurrence_count,resolved_at,create_times,update_times)
VALUES(?,?,?,?,?,?,?,?,?,1,0,?,?)
ON DUPLICATE KEY UPDATE
check_type=VALUES(check_type),biz_no=VALUES(biz_no),instruction_id=VALUES(instruction_id),
expected_value=VALUES(expected_value),actual_value=VALUES(actual_value),detail=VALUES(detail),
status=VALUES(status),occurrence_count=occurrence_count+1,resolved_at=0,update_times=VALUES(update_times)`,
		item.TenantId, item.IssueKey, item.CheckType, item.BizNo, item.InstructionId,
		item.ExpectedValue, item.ActualValue, item.Detail, item.Status, item.CreateTimes, item.UpdateTimes)
	return err
}

// ResolveOpenOptionReconciliationIssuesByPrefix resolves the previous result set in the
// caller's transaction before current mismatches are reopened.
func ResolveOpenOptionReconciliationIssuesByPrefix(
	ctx context.Context, conn sqlx.SqlConn, tenantId, checkType int64, issueKeyPrefix string, now int64,
) error {
	_, err := conn.ExecCtx(ctx, `UPDATE t_option_reconciliation_issue
SET status=?,resolved_at=?,update_times=?
WHERE tenant_id=? AND check_type=? AND issue_key LIKE ? AND status=?`,
		int64(option.ReconciliationIssueStatus_RECONCILIATION_ISSUE_STATUS_RESOLVED),
		now, now, tenantId, checkType, issueKeyPrefix+"%",
		int64(option.ReconciliationIssueStatus_RECONCILIATION_ISSUE_STATUS_OPEN))
	return err
}

func (m *customTOptionReconciliationIssueModel) Resolve(ctx context.Context, tenantId int64, issueKey string, now int64) error {
	_, err := m.ExecNoCacheCtx(ctx, `UPDATE t_option_reconciliation_issue
SET status=?,resolved_at=?,update_times=?
WHERE tenant_id=? AND issue_key=? AND status=?`,
		int64(option.ReconciliationIssueStatus_RECONCILIATION_ISSUE_STATUS_RESOLVED),
		now, now, tenantId, issueKey,
		int64(option.ReconciliationIssueStatus_RECONCILIATION_ISSUE_STATUS_OPEN))
	return err
}

func (m *customTOptionReconciliationIssueModel) ResolveOpenByPrefix(
	ctx context.Context, tenantId int64, issueKeyPrefix string, now int64,
) error {
	_, err := m.ExecNoCacheCtx(ctx, `UPDATE t_option_reconciliation_issue
SET status=?,resolved_at=?,update_times=?
WHERE tenant_id=? AND check_type=? AND issue_key LIKE ? AND status=?`,
		int64(option.ReconciliationIssueStatus_RECONCILIATION_ISSUE_STATUS_RESOLVED),
		now, now, tenantId,
		int64(option.ReconciliationCheckType_RECONCILIATION_CHECK_TYPE_BALANCE_MIRROR),
		issueKeyPrefix+"%",
		int64(option.ReconciliationIssueStatus_RECONCILIATION_ISSUE_STATUS_OPEN))
	return err
}

func (m *customTOptionReconciliationIssueModel) ResolveOpenByPrefixForCheckType(
	ctx context.Context, tenantId, checkType int64, issueKeyPrefix string, now int64,
) error {
	_, err := m.ExecNoCacheCtx(ctx, `UPDATE t_option_reconciliation_issue
SET status=?,resolved_at=?,update_times=?
WHERE tenant_id=? AND check_type=? AND issue_key LIKE ? AND status=?`,
		int64(option.ReconciliationIssueStatus_RECONCILIATION_ISSUE_STATUS_RESOLVED),
		now, now, tenantId, checkType, issueKeyPrefix+"%",
		int64(option.ReconciliationIssueStatus_RECONCILIATION_ISSUE_STATUS_OPEN))
	return err
}
