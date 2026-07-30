package adminlogic

import (
	"fmt"
	"strings"

	"wklive/proto/common"
	"wklive/proto/option"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
)

func parseOptionalOptionRate(value string) (decimal.Decimal, error) {
	if strings.TrimSpace(value) == "" {
		return decimal.Zero, nil
	}
	rate, err := decimal.NewFromString(value)
	if err != nil || rate.IsNegative() || rate.GreaterThan(decimal.NewFromInt(1)) {
		return decimal.Zero, fmt.Errorf("invalid option fee rate")
	}
	return rate, nil
}

// validateSupportedContract exposes only combinations whose clearing and risk
// paths are implemented. Cash supports European and American exercise; physical
// delivery is restricted to European, auto-exercised, fully covered contracts.
func validateSupportedContract(item *models.TOptionContract) bool {
	if item == nil || item.TenantId <= 0 ||
		strings.TrimSpace(item.ContractCode) == "" ||
		strings.TrimSpace(item.UnderlyingSymbol) == "" ||
		strings.TrimSpace(item.SettleCoin) == "" ||
		strings.TrimSpace(item.QuoteCoin) == "" {
		return false
	}
	if item.OptionType != int64(option.OptionType_OPTION_TYPE_CALL) &&
		item.OptionType != int64(option.OptionType_OPTION_TYPE_PUT) {
		return false
	}
	if (item.ExerciseStyle != int64(option.ExerciseStyle_EXERCISE_STYLE_EUROPEAN) &&
		item.ExerciseStyle != int64(option.ExerciseStyle_EXERCISE_STYLE_AMERICAN)) ||
		item.IsAutoExercise != int64(common.YesNo_YES_NO_YES) {
		return false
	}
	switch option.SettlementType(item.SettlementType) {
	case option.SettlementType_SETTLEMENT_TYPE_CASH:
		if item.PhysicalDeliveryPolicy != int64(option.PhysicalDeliveryPolicy_PHYSICAL_DELIVERY_POLICY_UNKNOWN) {
			return false
		}
	case option.SettlementType_SETTLEMENT_TYPE_PHYSICAL:
		if item.ExerciseStyle != int64(option.ExerciseStyle_EXERCISE_STYLE_EUROPEAN) ||
			strings.TrimSpace(item.UnderlyingCoin) == "" ||
			strings.EqualFold(item.UnderlyingCoin, item.SettleCoin) ||
			item.PhysicalDeliveryPolicy != int64(option.PhysicalDeliveryPolicy_PHYSICAL_DELIVERY_POLICY_STRICT) ||
			item.ExerciseFeeRate.IsPositive() {
			return false
		}
	default:
		return false
	}
	if !item.StrikePrice.IsPositive() || !item.ContractUnit.IsPositive() ||
		!item.MinOrderQty.IsPositive() || !item.PriceTick.IsPositive() ||
		!item.QtyStep.IsPositive() || !item.Multiplier.IsPositive() {
		return false
	}
	if item.MaxOrderQty.IsPositive() && item.MaxOrderQty.LessThan(item.MinOrderQty) {
		return false
	}
	if item.MakerFeeRate.IsNegative() || item.MakerFeeRate.GreaterThan(decimal.NewFromInt(1)) ||
		item.TakerFeeRate.IsNegative() || item.TakerFeeRate.GreaterThan(decimal.NewFromInt(1)) ||
		item.ExerciseFeeRate.IsNegative() || item.ExerciseFeeRate.GreaterThan(decimal.NewFromInt(1)) {
		return false
	}
	if (item.MakerFeeRate.IsPositive() || item.TakerFeeRate.IsPositive() || item.ExerciseFeeRate.IsPositive()) &&
		(item.FeeUserId <= 0 || item.FeeAccountId <= 0) {
		return false
	}
	switch option.SellerMarginMode(item.SellerMarginMode) {
	case option.SellerMarginMode_SELLER_MARGIN_MODE_DISABLED:
		if item.LiquidationDeficitPolicy !=
			int64(option.LiquidationDeficitPolicy_LIQUIDATION_DEFICIT_POLICY_MANUAL_REVIEW) {
			return false
		}
	case option.SellerMarginMode_SELLER_MARGIN_MODE_ISOLATED,
		option.SellerMarginMode_SELLER_MARGIN_MODE_PORTFOLIO:
		if item.SettlementType != int64(option.SettlementType_SETTLEMENT_TYPE_CASH) {
			return false
		}
		if !item.InitialMarginRate.IsPositive() || !item.MaintenanceMarginRate.IsPositive() ||
			!item.MinMarginRate.IsPositive() ||
			item.InitialMarginRate.LessThan(item.MaintenanceMarginRate) ||
			item.MaintenanceMarginRate.LessThan(item.MinMarginRate) ||
			item.LiquidationFeeRate.IsNegative() ||
			item.InsuranceUserId <= 0 || item.InsuranceAccountId <= 0 {
			return false
		}
		switch option.LiquidationDeficitPolicy(item.LiquidationDeficitPolicy) {
		case option.LiquidationDeficitPolicy_LIQUIDATION_DEFICIT_POLICY_MANUAL_REVIEW,
			option.LiquidationDeficitPolicy_LIQUIDATION_DEFICIT_POLICY_PLATFORM_BACKSTOP:
		default:
			return false
		}
		if item.SellerMarginMode == int64(option.SellerMarginMode_SELLER_MARGIN_MODE_PORTFOLIO) &&
			item.ExerciseStyle != int64(option.ExerciseStyle_EXERCISE_STYLE_EUROPEAN) {
			// Removing a long protection leg through early exercise needs a
			// separate pre-exercise margin admission flow. V1 stays European.
			return false
		}
	case option.SellerMarginMode_SELLER_MARGIN_MODE_COVERED_DELIVERY:
		if item.SettlementType != int64(option.SettlementType_SETTLEMENT_TYPE_PHYSICAL) ||
			item.LiquidationDeficitPolicy != int64(option.LiquidationDeficitPolicy_LIQUIDATION_DEFICIT_POLICY_MANUAL_REVIEW) {
			return false
		}
	default:
		return false
	}
	if item.ListTime <= 0 || item.ExpireTime <= item.ListTime ||
		item.DeliverTime < item.ExpireTime {
		return false
	}
	switch option.ContractStatus(item.Status) {
	case option.ContractStatus_CONTRACT_STATUS_PENDING,
		option.ContractStatus_CONTRACT_STATUS_TRADING,
		option.ContractStatus_CONTRACT_STATUS_PAUSED,
		option.ContractStatus_CONTRACT_STATUS_OFFLINE:
		return true
	default:
		return false
	}
}

func economicContractFieldsEqual(left, right *models.TOptionContract) bool {
	return left.TenantId == right.TenantId &&
		left.ContractCode == right.ContractCode &&
		left.UnderlyingSymbol == right.UnderlyingSymbol &&
		left.UnderlyingCoin == right.UnderlyingCoin &&
		left.SettleCoin == right.SettleCoin &&
		left.QuoteCoin == right.QuoteCoin &&
		left.OptionType == right.OptionType &&
		left.ExerciseStyle == right.ExerciseStyle &&
		left.SettlementType == right.SettlementType &&
		left.StrikePrice.Equal(right.StrikePrice) &&
		left.ContractUnit.Equal(right.ContractUnit) &&
		left.MinOrderQty.Equal(right.MinOrderQty) &&
		left.MaxOrderQty.Equal(right.MaxOrderQty) &&
		left.PriceTick.Equal(right.PriceTick) &&
		left.QtyStep.Equal(right.QtyStep) &&
		left.Multiplier.Equal(right.Multiplier) &&
		left.ListTime == right.ListTime &&
		left.ExpireTime == right.ExpireTime &&
		left.DeliverTime == right.DeliverTime &&
		left.IsAutoExercise == right.IsAutoExercise &&
		left.MakerFeeRate.Equal(right.MakerFeeRate) &&
		left.TakerFeeRate.Equal(right.TakerFeeRate) &&
		left.ExerciseFeeRate.Equal(right.ExerciseFeeRate) &&
		left.FeeUserId == right.FeeUserId &&
		left.FeeAccountId == right.FeeAccountId &&
		left.SellerMarginMode == right.SellerMarginMode &&
		left.InitialMarginRate.Equal(right.InitialMarginRate) &&
		left.MaintenanceMarginRate.Equal(right.MaintenanceMarginRate) &&
		left.MinMarginRate.Equal(right.MinMarginRate) &&
		left.LiquidationFeeRate.Equal(right.LiquidationFeeRate) &&
		left.InsuranceUserId == right.InsuranceUserId &&
		left.InsuranceAccountId == right.InsuranceAccountId &&
		left.LiquidationDeficitPolicy == right.LiquidationDeficitPolicy &&
		left.PhysicalDeliveryPolicy == right.PhysicalDeliveryPolicy
}
