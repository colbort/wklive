package adminlogic

import (
	"context"
	"time"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOperationsOverviewLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetOperationsOverviewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOperationsOverviewLogic {
	return &GetOperationsOverviewLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询运营异常和任务水位汇总
func (l *GetOperationsOverviewLogic) GetOperationsOverview(in *option.GetOperationsOverviewReq) (*option.GetOperationsOverviewResp, error) {
	tenantId, allowed, forbidden, err := utils.ResolveAdminTenantReadScopeFromMd(l.ctx, in.TenantId)
	if err != nil {
		return nil, err
	}
	if forbidden || !allowed {
		return &option.GetOperationsOverviewResp{
			Base: helper.ErrResp(i18n.PermissionDenied, i18n.Translate(i18n.PermissionDenied, l.ctx)),
		}, nil
	}
	staleSeconds, valid := normalizeOperationsStaleSeconds(in.RiskStaleSeconds)
	if !valid {
		return &option.GetOperationsOverviewResp{
			Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx)),
		}, nil
	}
	comboStaleSeconds, valid := normalizeOperationsStaleSeconds(in.ComboStaleSeconds)
	if !valid {
		return &option.GetOperationsOverviewResp{
			Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx)),
		}, nil
	}
	now := time.Now().Unix()
	summary, err := models.QueryOptionOperationsOverview(
		l.ctx, l.svcCtx.DB, tenantId, now-staleSeconds, now-comboStaleSeconds,
	)
	if err != nil {
		return nil, err
	}
	data := &option.OptionOperationsOverview{
		GeneratedAt: now, RiskStaleSeconds: staleSeconds, ComboStaleSeconds: comboStaleSeconds,
		AssetPendingCount:              summary.AssetPendingCount,
		AssetFailedCount:               summary.AssetFailedCount,
		AssetManualReviewCount:         summary.AssetManualReviewCount,
		OldestAssetInstructionTime:     summary.OldestAssetInstructionTime,
		OpenReconciliationCount:        summary.OpenReconciliationCount,
		OldestReconciliationTime:       summary.OldestReconciliationTime,
		PendingSettlementPriceCount:    summary.PendingSettlementPriceCount,
		StaleRiskAccountCount:          summary.StaleRiskAccountCount,
		OldestRiskCalcTime:             summary.OldestRiskCalcTime,
		PendingExerciseCount:           summary.PendingExerciseCount,
		OldestExerciseTime:             summary.OldestExerciseTime,
		PendingSettlementCount:         summary.PendingSettlementCount,
		FailedSettlementCount:          summary.FailedSettlementCount,
		OldestSettlementTime:           summary.OldestSettlementTime,
		PendingLiquidationCount:        summary.PendingLiquidationCount,
		ExceptionLiquidationCount:      summary.ExceptionLiquidationCount,
		OldestLiquidationTime:          summary.OldestLiquidationTime,
		PendingOutboxCount:             summary.PendingOutboxCount,
		OldestOutboxTime:               summary.OldestOutboxTime,
		PendingInboxCount:              summary.PendingInboxCount,
		OldestInboxTime:                summary.OldestInboxTime,
		PhysicalExceptionCount:         summary.PhysicalExceptionCount,
		ComboStaleCount:                summary.ComboStaleCount,
		ComboManualReviewCount:         summary.ComboManualReviewCount,
		OldestComboExceptionTime:       summary.OldestComboExceptionTime,
		ComboInvariantIssueCount:       summary.ComboInvariantIssueCount,
		ComboIncompleteMatchGroupCount: summary.ComboIncompleteMatchCount,
	}
	data.InsuranceLedger = toOperationsCoinAmounts(summary.InsuranceLedger)
	data.BackstopLiability = toOperationsCoinAmounts(summary.BackstopLiability)
	data.UnresolvedDeficit = toOperationsCoinAmounts(summary.UnresolvedDeficit)
	return &option.GetOperationsOverviewResp{Base: helper.OkResp(), Data: data}, nil
}

func normalizeOperationsStaleSeconds(value int64) (int64, bool) {
	if value == 0 {
		return 60, true
	}
	if value < 10 || value > 300 {
		return 0, false
	}
	return value, true
}

func toOperationsCoinAmounts(items []*models.OptionCoinAmount) []*option.OptionCoinAmount {
	result := make([]*option.OptionCoinAmount, 0, len(items))
	for _, item := range items {
		result = append(result, &option.OptionCoinAmount{
			Coin: item.Coin, Amount: item.Amount.String(),
		})
	}
	return result
}
