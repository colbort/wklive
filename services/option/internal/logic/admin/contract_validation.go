package adminlogic

import (
	"fmt"
	"strings"

	"wklive/proto/common"
	"wklive/proto/option"
	logichelpers "wklive/services/option/internal/logic/helpers"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
)

const (
	defaultSettlementPriceSource   = "authoritative-market"
	defaultSettlementPriceMethod   = "MEDIAN"
	defaultSettlementWindowSeconds = int64(60)
	defaultSettlementMinSamples    = int64(3)
	minPhysicalCureSeconds         = int64(300)
	maxPhysicalCureSeconds         = int64(7 * 24 * 60 * 60)
)

func normalizeSettlementPriceRule(item *models.TOptionContract) {
	if item == nil {
		return
	}
	if strings.TrimSpace(item.SettlementPriceSource) == "" {
		item.SettlementPriceSource = defaultSettlementPriceSource
	}
	item.SettlementPriceSource = strings.ToLower(strings.TrimSpace(item.SettlementPriceSource))
	if strings.TrimSpace(item.SettlementPriceMethod) == "" {
		item.SettlementPriceMethod = defaultSettlementPriceMethod
	}
	item.SettlementPriceMethod = strings.ToUpper(strings.TrimSpace(item.SettlementPriceMethod))
	if item.SettlementWindowSeconds == 0 {
		item.SettlementWindowSeconds = defaultSettlementWindowSeconds
	}
	if item.SettlementMinSamples == 0 {
		item.SettlementMinSamples = defaultSettlementMinSamples
	}
	if item.ExerciseCutoffTime == 0 {
		item.ExerciseCutoffTime = item.ExpireTime
	}
}

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
	normalizeSettlementPriceRule(item)
	if item == nil {
		return false
	}
	if strings.TrimSpace(item.TradingCalendarCode) == "" {
		item.TradingCalendarCode = logichelpers.DefaultTradingCalendarCode
	}
	calendarCode, validCalendarCode := logichelpers.NormalizeTradingCalendarCode(item.TradingCalendarCode)
	if !validCalendarCode {
		return false
	}
	item.TradingCalendarCode = calendarCode
	if item.TenantId <= 0 ||
		strings.TrimSpace(item.ContractCode) == "" ||
		strings.TrimSpace(item.UnderlyingSymbol) == "" ||
		strings.TrimSpace(item.SettleCoin) == "" ||
		strings.TrimSpace(item.QuoteCoin) == "" {
		return false
	}
	if item.SettlementPriceSource != defaultSettlementPriceSource ||
		item.SettlementPriceMethod != defaultSettlementPriceMethod ||
		item.SettlementWindowSeconds < 1 || item.SettlementWindowSeconds > 3600 ||
		item.SettlementMinSamples < 1 || item.SettlementMinSamples > 1000 {
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
		if item.PhysicalDeliveryPolicy != int64(option.PhysicalDeliveryPolicy_PHYSICAL_DELIVERY_POLICY_UNKNOWN) ||
			item.PhysicalDeliveryCureSeconds != 0 {
			return false
		}
	case option.SettlementType_SETTLEMENT_TYPE_PHYSICAL:
		if item.ExerciseStyle != int64(option.ExerciseStyle_EXERCISE_STYLE_EUROPEAN) ||
			strings.TrimSpace(item.UnderlyingCoin) == "" ||
			strings.EqualFold(item.UnderlyingCoin, item.SettleCoin) ||
			item.PhysicalDeliveryPolicy != int64(option.PhysicalDeliveryPolicy_PHYSICAL_DELIVERY_POLICY_STRICT) ||
			item.ExerciseFeeRate.IsPositive() || item.AutoExerciseThreshold.IsPositive() ||
			item.PhysicalDeliveryCureSeconds < minPhysicalCureSeconds ||
			item.PhysicalDeliveryCureSeconds > maxPhysicalCureSeconds {
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
	if item.AutoExerciseThreshold.IsNegative() {
		return false
	}
	if item.MaxUserLongQty.IsNegative() || item.MaxUserShortQty.IsNegative() ||
		item.MaxOpenInterest.IsNegative() || item.OrderPriceBandRatio.IsNegative() ||
		item.OrderPriceBandRatio.GreaterThan(decimal.NewFromInt(1)) ||
		item.CircuitBreakerRatio.IsNegative() ||
		item.CircuitBreakerRatio.GreaterThan(decimal.NewFromInt(1)) {
		return false
	}
	if item.Status == int64(option.ContractStatus_CONTRACT_STATUS_TRADING) &&
		(!item.MaxUserLongQty.IsPositive() || !item.MaxUserShortQty.IsPositive() ||
			!item.MaxOpenInterest.IsPositive() || !item.OrderPriceBandRatio.IsPositive() ||
			!item.CircuitBreakerRatio.IsPositive()) {
		// A live contract must never interpret zero as an unlimited risk control.
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
		item.DeliverTime < item.ExpireTime ||
		item.ExerciseCutoffTime <= item.ListTime ||
		item.ExerciseCutoffTime > item.ExpireTime {
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
		left.TradingCalendarCode == right.TradingCalendarCode &&
		left.ExerciseCutoffTime == right.ExerciseCutoffTime &&
		left.AutoExerciseThreshold.Equal(right.AutoExerciseThreshold) &&
		left.SettlementPriceSource == right.SettlementPriceSource &&
		left.SettlementPriceMethod == right.SettlementPriceMethod &&
		left.SettlementWindowSeconds == right.SettlementWindowSeconds &&
		left.SettlementMinSamples == right.SettlementMinSamples &&
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
		left.PhysicalDeliveryPolicy == right.PhysicalDeliveryPolicy &&
		left.PhysicalDeliveryCureSeconds == right.PhysicalDeliveryCureSeconds
}
