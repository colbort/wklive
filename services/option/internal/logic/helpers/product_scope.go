package helpers

import (
	"wklive/proto/common"
	"wklive/proto/option"
	"wklive/services/option/internal/config"
	"wklive/services/option/models"
)

const (
	ProductScopeSellerTradingDisabled    = "SELLER_TRADING_DISABLED"
	ProductScopePortfolioMarginDisabled  = "PORTFOLIO_MARGIN_DISABLED"
	ProductScopePhysicalDeliveryDisabled = "PHYSICAL_DELIVERY_DISABLED"
	ProductScopeMMPDisabled              = "MMP_DISABLED"
	ProductScopeAmericanExerciseDisabled = "AMERICAN_EXERCISE_DISABLED"
)

// ContractLaunchProductScopeReady prevents optional contract types from being
// published before their production scope has been explicitly approved.
func ContractLaunchProductScopeReady(
	contract *models.TOptionContract, scope config.ProductScope,
) (bool, string) {
	if contract == nil {
		return false, "CONTRACT_MISSING"
	}
	if contract.SettlementType == int64(option.SettlementType_SETTLEMENT_TYPE_PHYSICAL) &&
		!scope.PhysicalDeliveryEnabled {
		return false, ProductScopePhysicalDeliveryDisabled
	}
	if contract.ExerciseStyle == int64(option.ExerciseStyle_EXERCISE_STYLE_AMERICAN) &&
		!scope.AmericanExerciseEnabled {
		return false, ProductScopeAmericanExerciseDisabled
	}
	switch option.SellerMarginMode(contract.SellerMarginMode) {
	case option.SellerMarginMode_SELLER_MARGIN_MODE_DISABLED:
		return true, ""
	case option.SellerMarginMode_SELLER_MARGIN_MODE_PORTFOLIO:
		if !scope.SellerTradingEnabled {
			return false, ProductScopeSellerTradingDisabled
		}
		if !scope.PortfolioMarginEnabled {
			return false, ProductScopePortfolioMarginDisabled
		}
	case option.SellerMarginMode_SELLER_MARGIN_MODE_ISOLATED,
		option.SellerMarginMode_SELLER_MARGIN_MODE_COVERED_DELIVERY:
		if !scope.SellerTradingEnabled {
			return false, ProductScopeSellerTradingDisabled
		}
	default:
		return false, "SELLER_MARGIN_MODE_UNSUPPORTED"
	}
	return true, ""
}

// OrderProductScopeReady blocks only risk-increasing paths when an optional
// capability is disabled. CLOSE orders remain available for risk reduction.
func OrderProductScopeReady(
	contract *models.TOptionContract,
	scope config.ProductScope,
	side common.Side,
	effect option.PositionEffect,
	mmp common.YesNo,
) (bool, string) {
	if contract == nil {
		return false, "CONTRACT_MISSING"
	}
	if mmp == common.YesNo_YES_NO_YES && !scope.MMPEnabled {
		return false, ProductScopeMMPDisabled
	}
	if effect == option.PositionEffect_POSITION_EFFECT_CLOSE {
		return true, ""
	}
	if contract.SettlementType == int64(option.SettlementType_SETTLEMENT_TYPE_PHYSICAL) &&
		!scope.PhysicalDeliveryEnabled {
		return false, ProductScopePhysicalDeliveryDisabled
	}
	if effect == option.PositionEffect_POSITION_EFFECT_OPEN &&
		contract.SellerMarginMode != int64(option.SellerMarginMode_SELLER_MARGIN_MODE_DISABLED) {
		if !scope.SellerTradingEnabled {
			return false, ProductScopeSellerTradingDisabled
		}
		if contract.SellerMarginMode == int64(option.SellerMarginMode_SELLER_MARGIN_MODE_PORTFOLIO) &&
			!scope.PortfolioMarginEnabled {
			return false, ProductScopePortfolioMarginDisabled
		}
	}
	_ = side // retained in the signature for explicit request-context auditing.
	return true, ""
}
