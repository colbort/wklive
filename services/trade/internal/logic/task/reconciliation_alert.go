package tasklogic

import (
	"context"
	"fmt"
	"time"

	"wklive/common/alert"
	"wklive/common/notify"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

const reconciliationAlertReminderInterval = 30 * time.Minute

type reconciliationAlertReservationReleaser interface {
	ReleaseAlertReservation(context.Context, int64, string, int64) error
}

func (l *ReconcileContractAssetFlowsLogic) recordContractReconciliationFinding(
	issue *models.TContractReconciliationIssue,
) error {
	shouldAlert, err := l.svcCtx.ContractReconcileIssueModel.RecordFindingWithAlert(
		l.ctx, issue, reconciliationAlertReminderInterval.Milliseconds(),
	)
	if err != nil {
		return err
	}
	if shouldAlert {
		logx.WithContext(l.ctx).Errorf(
			"contract reconciliation issue key=%s expected=%s actual=%s detail=%s",
			issue.IssueKey, issue.ExpectedValue, issue.ActualValue, issue.Detail,
		)
		if publishErr := deliverReservedReconciliationAlert(
			l.ctx,
			l.svcCtx.OperationalAlertNotifier,
			l.svcCtx.ContractReconcileIssueModel,
			issue,
		); publishErr != nil {
			return publishErr
		}
	}
	return nil
}

func deliverReservedReconciliationAlert(
	ctx context.Context,
	notifier alert.Notifier,
	releaser reconciliationAlertReservationReleaser,
	issue *models.TContractReconciliationIssue,
) error {
	if issue == nil {
		return fmt.Errorf("publish reconciliation alert: issue is nil")
	}
	reservedAt := issue.LastSeenAt
	if reservedAt <= 0 {
		reservedAt = issue.UpdateTimes
	}
	at := alert.New(
		alert.TypeContractReconciliation,
		alert.StateFiring,
		notify.EventLevelError,
		"trade",
		issue.IssueKey,
		"合约对账差异",
		fmt.Sprintf("expected=%s actual=%s detail=%s", issue.ExpectedValue, issue.ActualValue, issue.Detail),
		reservedAt,
	)
	at.TenantID = issue.TenantId
	at.Data = map[string]any{
		"checkType":     issue.CheckType,
		"bizType":       issue.BizType,
		"bizNo":         issue.BizNo,
		"instructionId": issue.InstructionId,
		"expected":      issue.ExpectedValue,
		"actual":        issue.ActualValue,
	}
	if publishErr := alert.Notify(ctx, notifier, at); publishErr != nil {
		releaseErr := releaser.ReleaseAlertReservation(
			ctx,
			issue.TenantId,
			issue.IssueKey,
			reservedAt,
		)
		if releaseErr != nil {
			return fmt.Errorf(
				"publish reconciliation alert: %v; release alert reservation: %w",
				publishErr,
				releaseErr,
			)
		}
		return fmt.Errorf("publish reconciliation alert: %w", publishErr)
	}
	return nil
}
