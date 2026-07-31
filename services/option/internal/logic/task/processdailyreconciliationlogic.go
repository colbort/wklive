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

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const accountMirrorIssuePrefix = "ACCOUNT_MIRROR:"

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
			if err := l.reconcileTenant(tenantID, now, in.TenantId > 0); err != nil {
				result = errors.Join(result, fmt.Errorf("tenant %d: %w", tenantID, err))
			}
		}
		if result != nil {
			return nil, result
		}
		return helpers.OkTaskResp(), nil
	})
}

func (l *ProcessDailyReconciliationLogic) reconcileTenant(tenantID int64, now time.Time, force bool) error {
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
