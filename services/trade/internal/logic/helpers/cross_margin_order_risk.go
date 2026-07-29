package helpers

import (
	"context"
	"errors"
	"fmt"

	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/proto/common"
	"wklive/services/trade/internal/svc"

	"github.com/shopspring/decimal"
)

// ValidateCrossMarginOpeningCapacity performs a synchronous account-level
// check immediately before an exposure-increasing CROSS order is persisted.
// The caller must serialize this check and the following Asset freeze by
// tenant/user/margin-asset so concurrent orders cannot spend the same capacity.
func ValidateCrossMarginOpeningCapacity(
	ctx context.Context,
	svcCtx *svc.ServiceContext,
	tenantID, userID int64,
	marginAsset string,
	orderMargin, fee, orderMaintenance decimal.Decimal,
) error {
	if svcCtx == nil || tenantID <= 0 || userID <= 0 || marginAsset == "" {
		return errors.New("invalid cross margin account identity")
	}

	unfinishedLiquidations, err := svcCtx.ContractAccountLiqModel.CountUnfinishedByRiskUnit(
		ctx, tenantID, userID, marginAsset,
	)
	if err != nil {
		return fmt.Errorf("query cross margin liquidation state: %w", err)
	}
	if unfinishedLiquidations > 0 {
		return errors.New("cross margin account has an unfinished liquidation")
	}

	freezingOpenings, err := svcCtx.TradeOrderModel.CountFreezingCrossMarginOpenings(
		ctx, tenantID, userID, marginAsset,
	)
	if err != nil {
		return fmt.Errorf("query cross margin freezing orders: %w", err)
	}
	if freezingOpenings > 0 {
		return errors.New("cross margin account has an unfinished opening reservation")
	}

	aggregate, err := svcCtx.ContractPositionModel.FindCrossMarginOpeningAggregate(
		ctx, tenantID, userID, marginAsset, utils.NowMillis()-30_000,
	)
	if err != nil {
		return fmt.Errorf("query cross margin positions: %w", err)
	}
	if aggregate.StaleMarkCount > 0 {
		return errors.New("cross margin mark-price projection is missing or stale")
	}

	resp, err := svcCtx.AssetAdminClient.GetUserAssetDetail(ctx, &asset.GetUserAssetDetailReq{
		TenantId: tenantID, UserId: userID,
		WalletType: common.WalletType_WALLET_TYPE_CONTRACT,
		Coin:       marginAsset,
	})
	if err != nil {
		return fmt.Errorf("query cross margin wallet: %w", err)
	}
	if resp == nil || resp.GetBase() == nil || resp.GetBase().GetCode() != 200 || resp.GetData() == nil {
		return errors.New("cross margin wallet query rejected")
	}
	walletTotal, err := decimal.NewFromString(resp.GetData().GetTotalAmount())
	if err != nil {
		return fmt.Errorf("parse cross margin wallet total: %w", err)
	}
	walletAvailable, err := decimal.NewFromString(resp.GetData().GetAvailableAmount())
	if err != nil {
		return fmt.Errorf("parse cross margin wallet available: %w", err)
	}
	return validateCrossMarginOpeningRisk(
		walletTotal, walletAvailable,
		aggregate.PositionMargin, aggregate.MaintenanceMargin, aggregate.UnrealizedPnl,
		orderMargin, fee, orderMaintenance,
	)
}

func validateCrossMarginOpeningRisk(
	walletTotal, walletAvailable,
	positionMargin, maintenanceMargin, unrealizedPnl,
	orderMargin, fee, orderMaintenance decimal.Decimal,
) error {
	if walletTotal.IsNegative() || walletAvailable.IsNegative() ||
		positionMargin.IsNegative() || maintenanceMargin.IsNegative() ||
		!orderMargin.IsPositive() || fee.IsNegative() || orderMaintenance.IsNegative() {
		return errors.New("invalid cross margin risk input")
	}
	if orderMargin.LessThan(orderMaintenance) {
		return errors.New("cross margin order margin is below maintenance margin")
	}
	requiredAvailable := orderMargin.Add(fee)
	availableMargin := walletAvailable.Add(unrealizedPnl)
	if availableMargin.LessThan(requiredAvailable) {
		return fmt.Errorf(
			"insufficient cross margin: available=%s required=%s",
			availableMargin, requiredAvailable,
		)
	}
	equityAfterFee := walletTotal.Add(positionMargin).Add(unrealizedPnl).Sub(fee)
	requiredMaintenance := maintenanceMargin.Add(orderMaintenance)
	if !equityAfterFee.GreaterThan(requiredMaintenance) {
		return fmt.Errorf(
			"cross margin risk limit exceeded: equity_after_fee=%s maintenance=%s",
			equityAfterFee, requiredMaintenance,
		)
	}
	return nil
}
