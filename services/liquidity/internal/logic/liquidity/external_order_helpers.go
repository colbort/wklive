package liquiditylogic

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/svc"
	"wklive/services/liquidity/models"
)

func parsePositive(name, value string) (float64, error) {
	number, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil || number <= 0 {
		return 0, fmt.Errorf("%s must be a positive number", name)
	}
	return number, nil
}

func loadExternalRoute(ctx context.Context, svcCtx *svc.ServiceContext, symbolID int64) (*models.TLiquiditySymbolConfig, *models.TLiquidityProvider, error) {
	config, err := svcCtx.SymbolConfigModel.FindActiveExternalBySymbol(ctx, symbolID)
	if err != nil {
		return nil, nil, fmt.Errorf("active external liquidity config: %w", err)
	}
	if config.ExternalProviderId <= 0 {
		return nil, nil, fmt.Errorf("external provider is not configured")
	}
	provider, err := svcCtx.ProviderModel.FindOne(ctx, config.ExternalProviderId)
	if err != nil {
		return nil, nil, err
	}
	if provider.ProviderType != int64(liquidity.ProviderType_PROVIDER_TYPE_EXTERNAL) {
		return nil, nil, fmt.Errorf("external provider not found")
	}
	if provider.Status != int64(liquidity.ProviderStatus_PROVIDER_STATUS_ENABLED) {
		return nil, nil, fmt.Errorf("external provider is disabled")
	}
	return config, provider, nil
}

func applyExternalResult(row *models.TLiquidityExternalOrder, resultStatus int64, externalOrderID string, filledQty, avgPrice, feeAmount float64, feeAsset, raw string) {
	now := time.Now().UnixMilli()
	row.ExternalOrderId = sql.NullString{String: externalOrderID, Valid: strings.TrimSpace(externalOrderID) != ""}
	row.FilledQty, row.AvgPrice, row.FeeAmount, row.FeeAsset = filledQty, avgPrice, feeAmount, feeAsset
	if resultStatus == 0 {
		resultStatus = int64(liquidity.ExternalOrderStatus_EXTERNAL_ORDER_STATUS_SUBMITTED)
	}
	row.Status, row.RawResponse = resultStatus, sql.NullString{String: raw, Valid: raw != ""}
	row.SubmittedAt, row.UpdateTimes = now, now
	if resultStatus == int64(liquidity.ExternalOrderStatus_EXTERNAL_ORDER_STATUS_FILLED) ||
		resultStatus == int64(liquidity.ExternalOrderStatus_EXTERNAL_ORDER_STATUS_CANCELED) ||
		resultStatus == int64(liquidity.ExternalOrderStatus_EXTERNAL_ORDER_STATUS_REJECTED) ||
		resultStatus == int64(liquidity.ExternalOrderStatus_EXTERNAL_ORDER_STATUS_FAILED) {
		row.FinishedAt = now
	}
	row.Version++
}
