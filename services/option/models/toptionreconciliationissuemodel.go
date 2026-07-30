package models

import (
	"context"

	"wklive/proto/option"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionReconciliationIssueModel = (*customTOptionReconciliationIssueModel)(nil)

type (
	// TOptionReconciliationIssueModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionReconciliationIssueModel.
	TOptionReconciliationIssueModel interface {
		tOptionReconciliationIssueModel
		Open(ctx context.Context, item *TOptionReconciliationIssue) error
		Resolve(ctx context.Context, tenantId int64, issueKey string, now int64) error
	}

	customTOptionReconciliationIssueModel struct {
		*defaultTOptionReconciliationIssueModel
	}
)

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

func (m *customTOptionReconciliationIssueModel) Resolve(ctx context.Context, tenantId int64, issueKey string, now int64) error {
	_, err := m.ExecNoCacheCtx(ctx, `UPDATE t_option_reconciliation_issue
SET status=?,resolved_at=?,update_times=?
WHERE tenant_id=? AND issue_key=? AND status=?`,
		int64(option.ReconciliationIssueStatus_RECONCILIATION_ISSUE_STATUS_RESOLVED),
		now, now, tenantId, issueKey,
		int64(option.ReconciliationIssueStatus_RECONCILIATION_ISSUE_STATUS_OPEN))
	return err
}
