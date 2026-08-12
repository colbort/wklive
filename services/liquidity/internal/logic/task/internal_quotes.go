package tasklogic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/provider"
	"wklive/services/liquidity/internal/svc"
	"wklive/services/liquidity/models"

	"github.com/zeromicro/go-zero/core/logx"
)

func processInternalQuotes(ctx context.Context, svcCtx *svc.ServiceContext, in *liquidity.LiquidityTaskReq, recover bool) (*liquidity.LiquidityTaskResp, error) {
	limit := int64(in.BatchSize)
	if limit <= 0 {
		limit = 100
	}
	status := int64(liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_PENDING_SUBMIT)
	var rows []*models.TLiquidityQuoteOrder
	var err error
	if recover {
		rows, err = svcCtx.QuoteOrderModel.FindRecoveryCandidates(ctx, models.LiquidityQuoteOrderPageFilter{
			ConfigId: in.ConfigId, ProviderId: in.ProviderId,
		}, limit)
	} else {
		rows, _, err = svcCtx.QuoteOrderModel.FindPage(ctx, models.LiquidityQuoteOrderPageFilter{
			ConfigId: in.ConfigId, ProviderId: in.ProviderId, Status: status,
		}, 0, limit)
	}
	if err != nil {
		return nil, err
	}
	resp := &liquidity.LiquidityTaskResp{Base: helper.OkResp(), ScannedCount: int64(len(rows))}
	blockedConfigs := make(map[int64]error)
	type tradingCheck struct {
		open bool
		err  error
	}
	tradingChecks := make(map[int64]tradingCheck)
	for _, row := range rows {
		if cause, blocked := blockedConfigs[row.ConfigId]; blocked {
			markQuoteFailed(ctx, svcCtx, row, cause)
			resp.FailedCount++
			continue
		}
		if !recover {
			check, ok := tradingChecks[row.ConfigId]
			if !ok {
				config, findErr := svcCtx.SymbolConfigModel.FindOne(ctx, row.ConfigId)
				if findErr != nil {
					check.err = findErr
				} else {
					check.open, check.err = ensureMarketOpen(ctx, svcCtx, config, time.Now().UnixMilli())
				}
				tradingChecks[row.ConfigId] = check
			}
			if check.err != nil {
				resp.FailedCount++
				continue
			}
			if !check.open {
				continue
			}
		}
		providerRow, findErr := svcCtx.ProviderModel.FindOne(ctx, row.ProviderId)
		if findErr != nil {
			markQuoteFailed(ctx, svcCtx, row, findErr)
			resp.FailedCount++
			continue
		}
		var resultErr error
		var result *provider.QuoteResult
		if recover {
			var queryErr error
			result, queryErr = svcCtx.InternalMarketMaker.QueryQuote(ctx, providerRow, row)
			resultErr = queryErr
			if queryErr == nil {
				applyQuoteResult(row, result)
			}
		} else {
			var placeErr error
			result, placeErr = svcCtx.InternalMarketMaker.PlaceQuote(ctx, providerRow, row)
			resultErr = placeErr
			if placeErr == nil {
				applyQuoteResult(row, result)
			}
		}
		if resultErr != nil {
			if recover {
				markQuoteUncertain(ctx, svcCtx, row, resultErr)
			} else {
				markQuoteFailed(ctx, svcCtx, row, resultErr)
				if i18n.IsStatusError(resultErr, i18n.InsufficientAvailableBalance) {
					cause := fmt.Errorf("liquidity quoting stopped: internal provider balance is insufficient: %w", resultErr)
					blockedConfigs[row.ConfigId] = cause
					tripErr := tripLiquidityConfigCircuitBreaker(ctx, svcCtx, row.ConfigId, "internal provider balance insufficient")
					cancelErr := cancelActiveQuotes(ctx, svcCtx, row.ConfigId, "internal provider balance insufficient")
					if tripErr != nil || cancelErr != nil {
						logTaskError(ctx, "trip insufficient-balance circuit breaker", row, errors.Join(tripErr, cancelErr))
					}
				}
			}
			resp.FailedCount++
			continue
		}
		if !recover && result != nil && result.Status == int64(liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_FAILED) {
			if err := svcCtx.QuoteOrderModel.Update(ctx, row); err != nil {
				resp.FailedCount++
				continue
			}
			reason := "internal quote was rejected by trade"
			if result.Reason != "" {
				reason = fmt.Sprintf("%s: %s", reason, result.Reason)
			}
			cause := errors.New(reason)
			blockedConfigs[row.ConfigId] = cause
			tripErr := tripLiquidityConfigCircuitBreaker(ctx, svcCtx, row.ConfigId, "internal quote rejected by trade")
			cancelErr := cancelActiveQuotes(ctx, svcCtx, row.ConfigId, "internal quote rejected by trade")
			if tripErr != nil || cancelErr != nil {
				logTaskError(ctx, "trip rejected-quote circuit breaker", row, errors.Join(tripErr, cancelErr))
			}
			resp.FailedCount++
			continue
		}
		if err := svcCtx.QuoteOrderModel.Update(ctx, row); err != nil {
			resp.FailedCount++
			continue
		}
		resp.SuccessCount++
	}
	return resp, nil
}

func tripLiquidityConfigCircuitBreaker(ctx context.Context, svcCtx *svc.ServiceContext, configID int64, reason string) error {
	config, err := svcCtx.SymbolConfigModel.FindOne(ctx, configID)
	if err != nil {
		return err
	}
	if config.Status != int64(liquidity.SymbolLiquidityStatus_SYMBOL_LIQUIDITY_STATUS_RUNNING) {
		return nil
	}
	config.Status = int64(liquidity.SymbolLiquidityStatus_SYMBOL_LIQUIDITY_STATUS_CIRCUIT_BREAKER)
	config.PauseReason = reason
	config.UpdateTimes, config.Version = time.Now().UnixMilli(), config.Version+1
	return svcCtx.SymbolConfigModel.Update(ctx, config)
}

func logTaskError(ctx context.Context, action string, row *models.TLiquidityQuoteOrder, err error) {
	if err == nil {
		return
	}
	logx.WithContext(ctx).Errorf("%s failed: config_id=%d quote_no=%s err=%v", action, row.ConfigId, row.QuoteNo, err)
}

func applyQuoteResult(row *models.TLiquidityQuoteOrder, result *provider.QuoteResult) {
	row.InternalOrderId = result.InternalOrderID
	row.InternalOrderNo = result.OrderNo
	row.Status = result.Status
	row.FilledQty = result.FilledQty
	row.CancelReason = result.Reason
	row.LastErrorMsg = result.LastErrorMsg
	row.UpdateTimes, row.Version = time.Now().UnixMilli(), row.Version+1
}

func markQuoteUncertain(ctx context.Context, svcCtx *svc.ServiceContext, row *models.TLiquidityQuoteOrder, cause error) {
	row.Status = int64(liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_UNCERTAIN)
	row.LastErrorMsg = cause.Error()
	row.UpdateTimes, row.Version = time.Now().UnixMilli(), row.Version+1
	_ = svcCtx.QuoteOrderModel.Update(ctx, row)
}

func markQuoteFailed(ctx context.Context, svcCtx *svc.ServiceContext, row *models.TLiquidityQuoteOrder, cause error) {
	row.Status = int64(liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_FAILED)
	row.LastErrorMsg = cause.Error()
	row.UpdateTimes, row.Version = time.Now().UnixMilli(), row.Version+1
	_ = svcCtx.QuoteOrderModel.Update(ctx, row)
}
