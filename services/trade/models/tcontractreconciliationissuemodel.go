package models

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"wklive/common/sqlutil"
)

var _ TContractReconciliationIssueModel = (*customTContractReconciliationIssueModel)(nil)

type (
	ContractReconciliationIssuePageFilter struct {
		TenantId  int64
		Status    int64
		CheckType string
		BizNo     string
	}

	// TContractReconciliationIssueModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTContractReconciliationIssueModel.
	TContractReconciliationIssueModel interface {
		tContractReconciliationIssueModel
		RecordFinding(ctx context.Context, issue *TContractReconciliationIssue) error
		ResolveByKey(ctx context.Context, tenantID int64, issueKey, reason string, now int64) error
		FindPage(ctx context.Context, filter ContractReconciliationIssuePageFilter, cursor, limit int64) ([]*TContractReconciliationIssue, int64, error)
		FindOneForUpdate(ctx context.Context, id int64) (*TContractReconciliationIssue, error)
	}

	customTContractReconciliationIssueModel struct {
		*defaultTContractReconciliationIssueModel
	}
)

// NewTContractReconciliationIssueModel returns a model for the database table.
func NewTContractReconciliationIssueModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TContractReconciliationIssueModel {
	return &customTContractReconciliationIssueModel{
		defaultTContractReconciliationIssueModel: newTContractReconciliationIssueModel(conn, c, opts...),
	}
}

func (m *defaultTContractReconciliationIssueModel) FindOneForUpdate(ctx context.Context, id int64) (*TContractReconciliationIssue, error) {
	var row TContractReconciliationIssue
	query := fmt.Sprintf("SELECT %s FROM %s WHERE id=? LIMIT 1 FOR UPDATE", tContractReconciliationIssueRows, m.table)
	if err := m.QueryRowNoCacheCtx(ctx, &row, query, id); err != nil {
		return nil, err
	}
	return &row, nil
}

func (m *defaultTContractReconciliationIssueModel) FindPage(ctx context.Context, filter ContractReconciliationIssuePageFilter, cursor, limit int64) ([]*TContractReconciliationIssue, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqInt64("status", filter.Status)
	builder.EqString("check_type", filter.CheckType)
	builder.EqString("biz_no", filter.BizNo)
	where, args := builder.Where(), builder.Args()
	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where), args...); err != nil {
		return nil, 0, err
	}
	listArgs := append([]any{}, args...)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tContractReconciliationIssueRows, m.table, where)
	if cursor > 0 {
		query += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	query += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)
	var rows []*TContractReconciliationIssue
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, listArgs...); err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (m *defaultTContractReconciliationIssueModel) RecordFinding(ctx context.Context, issue *TContractReconciliationIssue) error {
	query := `INSERT INTO t_contract_reconciliation_issue
(tenant_id,issue_key,check_type,biz_type,biz_no,instruction_id,expected_value,actual_value,detail,status,occurrence_count,first_seen_at,last_seen_at,resolved_at,operator_id,resolution_reason,create_times,update_times)
VALUES(?,?,?,?,?,?,?,?,?,1,1,?,?,0,0,'',?,?)
ON DUPLICATE KEY UPDATE
actual_value=VALUES(actual_value),
detail=VALUES(detail),
last_seen_at=VALUES(last_seen_at),
occurrence_count=occurrence_count+1,
status=IF(status=3,3,1),
resolved_at=IF(status=3,resolved_at,0),
resolution_reason=IF(status=3,resolution_reason,''),
update_times=VALUES(update_times)`
	_, err := m.ExecNoCacheCtx(ctx, query,
		issue.TenantId, issue.IssueKey, issue.CheckType, issue.BizType, issue.BizNo,
		issue.InstructionId, issue.ExpectedValue, issue.ActualValue, issue.Detail,
		issue.FirstSeenAt, issue.LastSeenAt, issue.CreateTimes, issue.UpdateTimes)
	return err
}

func (m *defaultTContractReconciliationIssueModel) ResolveByKey(ctx context.Context, tenantID int64, issueKey, reason string, now int64) error {
	_, err := m.ExecNoCacheCtx(ctx,
		"UPDATE t_contract_reconciliation_issue SET status=2,resolved_at=?,resolution_reason=?,update_times=? WHERE tenant_id=? AND issue_key=? AND status=1",
		now, reason, now, tenantID, issueKey)
	return err
}
