package tasklogic

import (
	"context"
	"time"

	"wklive/common/helper"
	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/provider"
	"wklive/services/liquidity/internal/svc"
	"wklive/services/liquidity/models"
)

func processInternalQuotes(ctx context.Context, svcCtx *svc.ServiceContext, in *liquidity.LiquidityTaskReq, recover bool) (*liquidity.LiquidityTaskResp, error) {
	limit := int64(in.BatchSize)
	if limit <= 0 {
		limit = 100
	}
	status := int64(liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_PENDING_SUBMIT)
	if recover {
		status = int64(liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_UNCERTAIN)
	}
	rows, _, err := svcCtx.QuoteOrderModel.FindPage(ctx, models.LiquidityQuoteOrderPageFilter{
		ConfigId: in.ConfigId, ProviderId: in.ProviderId, Status: status,
	}, 0, limit)
	if err != nil {
		return nil, err
	}
	resp := &liquidity.LiquidityTaskResp{Base: helper.OkResp(), ScannedCount: int64(len(rows))}
	for _, row := range rows {
		provider, findErr := svcCtx.ProviderModel.FindOne(ctx, row.ProviderId)
		if findErr != nil {
			markQuoteFailed(ctx, svcCtx, row, findErr)
			resp.FailedCount++
			continue
		}
		var resultErr error
		if recover {
			result, queryErr := svcCtx.InternalMarketMaker.QueryQuote(ctx, provider, row)
			resultErr = queryErr
			if queryErr == nil {
				applyQuoteResult(row, result)
			}
		} else {
			result, placeErr := svcCtx.InternalMarketMaker.PlaceQuote(ctx, provider, row)
			resultErr = placeErr
			if placeErr == nil {
				applyQuoteResult(row, result)
			}
		}
		if resultErr != nil {
			markQuoteFailed(ctx, svcCtx, row, resultErr)
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

func applyQuoteResult(row *models.TLiquidityQuoteOrder, result *provider.QuoteResult) {
	row.InternalOrderId = result.InternalOrderID
	row.InternalOrderNo = result.OrderNo
	row.Status = result.Status
	row.FilledQty = result.FilledQty
	row.CancelReason = result.Reason
	row.LastErrorMsg = ""
	row.UpdateTimes, row.Version = time.Now().UnixMilli(), row.Version+1
}

func markQuoteFailed(ctx context.Context, svcCtx *svc.ServiceContext, row *models.TLiquidityQuoteOrder, cause error) {
	row.Status = int64(liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_FAILED)
	row.LastErrorMsg = cause.Error()
	row.UpdateTimes, row.Version = time.Now().UnixMilli(), row.Version+1
	_ = svcCtx.QuoteOrderModel.Update(ctx, row)
}
