package tasklogic

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"wklive/proto/option"
	"wklive/services/option/internal/logic/helpers"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const accountMirrorIssuePrefix = "ACCOUNT_MIRROR:"
const dailyConservationIssuePrefix = "DAILY:"

type ProcessDailyReconciliationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewProcessDailyReconciliationLogic(
	ctx context.Context, svcCtx *svc.ServiceContext,
) *ProcessDailyReconciliationLogic {
	return &ProcessDailyReconciliationLogic{
		ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx),
	}
}

func (l *ProcessDailyReconciliationLogic) ProcessDailyReconciliation(
	in *option.OptionTaskReq,
) (*option.OptionTaskResp, error) {
	return helpers.RunTaskWithLock(l.ctx, l.svcCtx, "process_daily_reconciliation", func() (*option.OptionTaskResp, error) {
		now := time.Now().UTC()
		// The hourly cron avoids depending on the host timezone. Automatic runs
		// only execute in the first UTC hour and close the previous UTC business day.
		// An explicit tenant is a controlled manual rerun and may run outside the window.
		if in.TenantId <= 0 && now.Hour() != 0 {
			return helpers.OkTaskResp(), nil
		}
		tenantIDs, err := models.ListOptionReconciliationTenantIDs(l.ctx, l.svcCtx.DB, in.TenantId)
		if err != nil {
			return nil, err
		}
		var result error
		for _, tenantID := range tenantIDs {
			if err := l.reconcileAccountMirror(tenantID, now, in.TenantId > 0); err != nil {
				result = errors.Join(result, fmt.Errorf("tenant %d account mirror: %w", tenantID, err))
			}
			if err := l.reconcileFullFunds(tenantID, now, in.TenantId > 0); err != nil {
				result = errors.Join(result, fmt.Errorf("tenant %d full funds: %w", tenantID, err))
			}
		}
		if result != nil {
			return nil, result
		}
		return helpers.OkTaskResp(), nil
	})
}

func (l *ProcessDailyReconciliationLogic) reconcileAccountMirror(tenantID int64, now time.Time, force bool) error {
	businessDate := reconciliationBusinessDate(now)
	snapshotTime := now.Unix()
	if !force {
		succeeded, err := models.HasSuccessfulOptionReconciliationRun(
			l.ctx, l.svcCtx.DB, tenantID, businessDate,
			models.OptionReconciliationScopeAccountMirror,
		)
		if err != nil {
			return err
		}
		if succeeded {
			return nil
		}
	}
	attempt, err := models.NextOptionReconciliationAttempt(
		l.ctx, l.svcCtx.DB, tenantID, businessDate,
		models.OptionReconciliationScopeAccountMirror,
	)
	if err != nil {
		return err
	}
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		summaries, queryErr := models.QueryOptionAccountMirrorSummaries(ctx, conn, tenantID)
		if queryErr != nil {
			return queryErr
		}
		issueModel := models.NewTOptionReconciliationIssueModel(conn, l.svcCtx.Config.CacheRedis)
		if resolveErr := issueModel.ResolveOpenByPrefix(
			ctx, tenantID, accountMirrorIssuePrefix, snapshotTime,
		); resolveErr != nil {
			return resolveErr
		}
		run := &models.OptionReconciliationRun{
			TenantID: tenantID, BusinessDate: businessDate,
			Scope:     models.OptionReconciliationScopeAccountMirror,
			AttemptNo: attempt, Status: models.OptionReconciliationRunSucceeded,
			SnapshotTime: snapshotTime,
			SnapshotRef:  fmt.Sprintf("mysql-single-statement:%d", snapshotTime),
			CompletedAt:  snapshotTime,
		}
		for _, summary := range summaries {
			if summary == nil {
				continue
			}
			run.CoinCount++
			run.AccountCount += summary.AccountCount
			run.MismatchCount += summary.MismatchCount
			if summary.MismatchCount == 0 {
				continue
			}
			run.Status = models.OptionReconciliationRunMismatch
			issue := &models.TOptionReconciliationIssue{
				TenantId:  tenantID,
				IssueKey:  accountMirrorIssuePrefix + summary.Coin,
				CheckType: int64(option.ReconciliationCheckType_RECONCILIATION_CHECK_TYPE_BALANCE_MIRROR),
				ExpectedValue: fmt.Sprintf(
					"option total=%s available=%s frozen=%s",
					summary.OptionTotal, summary.OptionAvailable, summary.OptionFrozen,
				),
				ActualValue: fmt.Sprintf(
					"asset total=%s available=%s frozen=%s locked=%s",
					summary.AssetTotal, summary.AssetAvailable,
					summary.AssetFrozen, summary.AssetLocked,
				),
				Detail: fmt.Sprintf(
					"UTC snapshot=%d coin=%s wallets=%d mismatches=%d",
					snapshotTime, summary.Coin, summary.AccountCount, summary.MismatchCount,
				),
				Status:          int64(option.ReconciliationIssueStatus_RECONCILIATION_ISSUE_STATUS_OPEN),
				OccurrenceCount: 1, CreateTimes: snapshotTime, UpdateTimes: snapshotTime,
			}
			if openErr := issueModel.Open(ctx, issue); openErr != nil {
				return openErr
			}
		}
		run.Detail = fmt.Sprintf(
			"account mirror snapshot: coins=%d wallets=%d mismatches=%d",
			run.CoinCount, run.AccountCount, run.MismatchCount,
		)
		return models.InsertOptionReconciliationRun(ctx, conn, run)
	})
	if err == nil {
		return nil
	}
	failureRun := &models.OptionReconciliationRun{
		TenantID: tenantID, BusinessDate: businessDate,
		Scope:     models.OptionReconciliationScopeAccountMirror,
		AttemptNo: attempt, Status: models.OptionReconciliationRunFailed,
		SnapshotTime: snapshotTime,
		SnapshotRef:  fmt.Sprintf("mysql-single-statement:%d", snapshotTime),
		Detail:       truncateReconciliationDetail(err.Error()), CompletedAt: time.Now().UTC().Unix(),
	}
	if recordErr := models.InsertOptionReconciliationRun(l.ctx, l.svcCtx.DB, failureRun); recordErr != nil {
		return errors.Join(err, fmt.Errorf("record failed reconciliation run: %w", recordErr))
	}
	return err
}

func (l *ProcessDailyReconciliationLogic) reconcileFullFunds(tenantID int64, now time.Time, force bool) error {
	businessDate := reconciliationBusinessDate(now)
	snapshotTime, snapshotMillis := now.Unix(), now.UnixMilli()
	if !force {
		succeeded, err := models.HasSuccessfulOptionReconciliationRun(
			l.ctx, l.svcCtx.DB, tenantID, businessDate, models.OptionReconciliationScopeFullFunds,
		)
		if err != nil || succeeded {
			return err
		}
	}
	attempt, err := models.NextOptionReconciliationAttempt(
		l.ctx, l.svcCtx.DB, tenantID, businessDate, models.OptionReconciliationScopeFullFunds,
	)
	if err != nil {
		return err
	}
	startMillis, endMillis := reconciliationDayWindow(now)
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		var isolation string
		if queryErr := conn.QueryRowCtx(ctx, &isolation, "SELECT @@transaction_isolation"); queryErr != nil {
			return queryErr
		}
		if isolation != "REPEATABLE-READ" && isolation != "SERIALIZABLE" {
			return fmt.Errorf("scope-2 requires repeatable-read or serializable, got %s", isolation)
		}
		wallets, queryErr := models.QueryOptionUserWalletConservationSummaries(
			ctx, conn, tenantID, startMillis, endMillis, snapshotMillis,
		)
		if queryErr != nil {
			return queryErr
		}
		platforms, queryErr := models.QueryOptionPlatformAccountConservationSummaries(
			ctx, conn, tenantID, startMillis, endMillis, snapshotMillis,
		)
		if queryErr != nil {
			return queryErr
		}
		subledgers, queryErr := models.QueryOptionSubledgerConservationSummaries(
			ctx, conn, tenantID, startMillis, endMillis, snapshotMillis,
		)
		if queryErr != nil {
			return queryErr
		}

		details := make([]*models.OptionReconciliationRunDetail, 0, len(wallets)+len(platforms)+len(subledgers))
		coins := make(map[string]struct{})
		maxUserFlowID, maxPlatformFlowID, maxInstructionID, maxBillID := int64(0), int64(0), int64(0), int64(0)
		for _, item := range wallets {
			if item == nil {
				continue
			}
			coins[item.Coin] = struct{}{}
			maxUserFlowID = max(maxUserFlowID, item.MaxFlowID)
			status := conservationDetailStatus(item.DifferenceAmount,
				item.IntegrityErrorCount > 0 || item.UnclassifiedFlowCount > 0)
			details = append(details, &models.OptionReconciliationRunDetail{
				TenantID: tenantID, BusinessDate: businessDate, Scope: models.OptionReconciliationScopeFullFunds,
				DimensionType: models.OptionReconciliationDimensionUserWallet, DimensionKey: item.Coin,
				OpeningAmount: item.OpeningAmount, ExternalNet: item.ExternalNet, OptionNet: item.OptionNet,
				ManualNet: item.ManualNet, ExpectedClosing: item.ExpectedClosing, ActualClosing: item.ActualClosing,
				DifferenceAmount: item.DifferenceAmount, FlowCount: item.FlowCount,
				MismatchCount: item.MismatchWalletCount, Status: status,
				EvidenceRef: fmt.Sprintf("asset_flow<=%d", item.MaxFlowID),
				Detail: fmt.Sprintf("wallets=%d integrity_errors=%d unclassified=%d",
					item.WalletCount, item.IntegrityErrorCount, item.UnclassifiedFlowCount), CreateTimes: snapshotTime,
			})
		}
		for _, item := range platforms {
			if item == nil {
				continue
			}
			coins[item.Coin] = struct{}{}
			maxPlatformFlowID = max(maxPlatformFlowID, item.MaxPlatformFlowID)
			status := conservationDetailStatus(item.DifferenceAmount, item.IntegrityErrorCount > 0)
			details = append(details, &models.OptionReconciliationRunDetail{
				TenantID: tenantID, BusinessDate: businessDate, Scope: models.OptionReconciliationScopeFullFunds,
				DimensionType: models.OptionReconciliationDimensionPlatformAccount,
				DimensionKey:  item.AccountType + ":" + item.Coin, OpeningAmount: item.OpeningAmount,
				ExternalNet: decimal.Zero, OptionNet: item.OptionNet, ManualNet: item.ManualNet,
				ExpectedClosing: item.ExpectedClosing, ActualClosing: item.ActualClosing,
				DifferenceAmount: item.DifferenceAmount, FlowCount: item.FlowCount,
				MismatchCount: item.MismatchAccountCount, Status: status,
				EvidenceRef: fmt.Sprintf("asset_platform_flow<=%d", item.MaxPlatformFlowID),
				Detail:      fmt.Sprintf("accounts=%d integrity_errors=%d", item.AccountCount, item.IntegrityErrorCount),
				CreateTimes: snapshotTime,
			})
		}
		for _, item := range subledgers {
			if item == nil {
				continue
			}
			coins[item.Coin] = struct{}{}
			maxUserFlowID = max(maxUserFlowID, item.MaxAssetFlowID)
			maxInstructionID = max(maxInstructionID, item.MaxInstructionID)
			maxBillID = max(maxBillID, item.MaxBillID)
			status := conservationDetailStatus(item.DifferenceAmount, item.MismatchCount > 0)
			details = append(details, &models.OptionReconciliationRunDetail{
				TenantID: tenantID, BusinessDate: businessDate, Scope: models.OptionReconciliationScopeFullFunds,
				DimensionType: models.OptionReconciliationDimensionOptionSubledger, DimensionKey: item.Coin,
				OpeningAmount: decimal.Zero, ExternalNet: decimal.Zero, OptionNet: item.AssetNet,
				ManualNet: decimal.Zero, ExpectedClosing: item.AssetNet, ActualClosing: item.BillNet,
				DifferenceAmount: item.DifferenceAmount,
				FlowCount:        item.AssetFlowCount + item.InstructionCount + item.BillCount,
				MismatchCount:    item.MismatchCount, Status: status,
				EvidenceRef: fmt.Sprintf("asset_flow<=%d;instruction<=%d;bill<=%d",
					item.MaxAssetFlowID, item.MaxInstructionID, item.MaxBillID),
				Detail: fmt.Sprintf("asset_flows=%d instructions=%d bills=%d",
					item.AssetFlowCount, item.InstructionCount, item.BillCount), CreateTimes: snapshotTime,
			})
		}

		run := &models.OptionReconciliationRun{
			TenantID: tenantID, BusinessDate: businessDate, Scope: models.OptionReconciliationScopeFullFunds,
			AttemptNo: attempt, Status: models.OptionReconciliationRunSucceeded,
			SnapshotTime: snapshotTime,
			SnapshotRef: fmt.Sprintf("mysql-repeatable-read:%d;asset_flow<=%d;platform_flow<=%d;instruction<=%d;bill<=%d",
				snapshotMillis, maxUserFlowID, maxPlatformFlowID, maxInstructionID, maxBillID),
			CoinCount: int64(len(coins)), AccountCount: int64(len(details)), CompletedAt: snapshotTime,
		}
		for _, detail := range details {
			if detail.Status != models.OptionReconciliationDetailMatched {
				run.Status = models.OptionReconciliationRunMismatch
				run.MismatchCount++
			}
		}
		run.Detail = fmt.Sprintf("full conservation: dimensions=%d mismatches=%d", len(details), run.MismatchCount)
		runID, insertErr := models.InsertOptionReconciliationRunWithID(ctx, conn, run)
		if insertErr != nil {
			return insertErr
		}
		issuePrefix := dailyConservationIssuePrefix + businessDate + ":"
		if resolveErr := models.ResolveOpenOptionReconciliationIssuesByPrefix(ctx, conn, tenantID,
			int64(option.ReconciliationCheckType_RECONCILIATION_CHECK_TYPE_SETTLEMENT_BALANCE),
			issuePrefix, snapshotTime); resolveErr != nil {
			return resolveErr
		}
		for _, detail := range details {
			detail.RunID = runID
			if insertErr := models.InsertOptionReconciliationRunDetail(ctx, conn, detail); insertErr != nil {
				return insertErr
			}
			if detail.Status == models.OptionReconciliationDetailMatched {
				continue
			}
			issueKey := fmt.Sprintf("%s%d:%s:2", issuePrefix, detail.DimensionType, detail.DimensionKey)
			if openErr := models.OpenOptionReconciliationIssue(ctx, conn, &models.TOptionReconciliationIssue{
				TenantId: tenantID, IssueKey: issueKey,
				CheckType: int64(option.ReconciliationCheckType_RECONCILIATION_CHECK_TYPE_SETTLEMENT_BALANCE),
				BizNo:     businessDate, ExpectedValue: detail.ExpectedClosing.String(),
				ActualValue: detail.ActualClosing.String(), Detail: detail.Detail,
				Status:          int64(option.ReconciliationIssueStatus_RECONCILIATION_ISSUE_STATUS_OPEN),
				OccurrenceCount: 1, CreateTimes: snapshotTime, UpdateTimes: snapshotTime,
			}); openErr != nil {
				return openErr
			}
		}
		return nil
	})
	if err == nil {
		return nil
	}
	failureRun := &models.OptionReconciliationRun{
		TenantID: tenantID, BusinessDate: businessDate, Scope: models.OptionReconciliationScopeFullFunds,
		AttemptNo: attempt, Status: models.OptionReconciliationRunFailed, SnapshotTime: snapshotTime,
		SnapshotRef: fmt.Sprintf("mysql-repeatable-read:%d", snapshotMillis),
		Detail:      truncateReconciliationDetail(err.Error()), CompletedAt: time.Now().UTC().Unix(),
	}
	if recordErr := models.InsertOptionReconciliationRun(l.ctx, l.svcCtx.DB, failureRun); recordErr != nil {
		return errors.Join(err, fmt.Errorf("record failed full-funds reconciliation run: %w", recordErr))
	}
	return err
}

func conservationDetailStatus(difference decimal.Decimal, incomplete bool) int64 {
	if incomplete {
		return models.OptionReconciliationDetailIncomplete
	}
	if !difference.IsZero() {
		return models.OptionReconciliationDetailMismatch
	}
	return models.OptionReconciliationDetailMatched
}

func reconciliationDayWindow(now time.Time) (int64, int64) {
	utc := now.UTC()
	end := time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
	return end.AddDate(0, 0, -1).UnixMilli(), end.UnixMilli()
}

func reconciliationBusinessDate(now time.Time) string {
	return now.UTC().AddDate(0, 0, -1).Format("2006-01-02")
}

func truncateReconciliationDetail(value string) string {
	value = strings.TrimSpace(value)
	runes := []rune(value)
	if len(runes) <= 1000 {
		return value
	}
	return string(runes[:1000])
}
