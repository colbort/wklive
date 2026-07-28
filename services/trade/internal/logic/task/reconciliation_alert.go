package tasklogic

import (
	"time"

	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

const reconciliationAlertReminderInterval = 30 * time.Minute

func (l *ReconcileContractAssetFlowsLogic) recordContractReconciliationFinding(
	issue *models.TContractReconciliationIssue,
) error {
	shouldAlert, err := l.svcCtx.ContractReconcileIssueModel.RecordFindingWithAlert(
		l.ctx, issue, reconciliationAlertReminderInterval.Milliseconds(),
	)
	if shouldAlert {
		logx.WithContext(l.ctx).Errorf(
			"contract reconciliation issue key=%s expected=%s actual=%s detail=%s",
			issue.IssueKey, issue.ExpectedValue, issue.ActualValue, issue.Detail,
		)
	}
	return err
}
