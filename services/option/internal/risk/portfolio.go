package risk

import (
	"errors"
	"fmt"
	"sort"

	"wklive/proto/option"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
)

// PortfolioLeg is the net position used by the deterministic portfolio risk
// engine. Pending orders are represented by applying their full-fill quantity
// to LongQuantity or ShortQuantity before EvaluatePortfolio is called.
type PortfolioLeg struct {
	Contract      *models.TOptionContract
	Market        *models.TOptionMarket
	LongQuantity  decimal.Decimal
	ShortQuantity decimal.Decimal
}

type PortfolioResult struct {
	Requirement  decimal.Decimal
	ScenarioLoss decimal.Decimal
	ShortFloor   decimal.Decimal
}

type portfolioGroupKey struct {
	underlying string
	settleCoin string
	expireTime int64
}

// EvaluatePortfolio implements PORTFOLIO_RISK_METHOD_EXPIRY_SCENARIO_V1.
//
// Offsets are deliberately limited to contracts with the same underlying,
// settlement coin and expiry. Each group is valued at zero spot, configured
// down/up shocks, current spot, every strike, and 2x spot. The largest loss is
// bounded below by a minimum requirement for net naked shorts. This keeps the
// method deterministic and prevents unrelated expiries or assets from
// receiving an implicit correlation credit.
func EvaluatePortfolio(legs []PortfolioLeg, maintenance bool) (PortfolioResult, error) {
	grouped := make(map[portfolioGroupKey][]PortfolioLeg)
	for _, leg := range legs {
		if leg.Contract == nil || leg.Market == nil {
			return PortfolioResult{}, errors.New("portfolio leg contract and market are required")
		}
		if leg.LongQuantity.IsNegative() || leg.ShortQuantity.IsNegative() {
			return PortfolioResult{}, errors.New("portfolio quantities cannot be negative")
		}
		if leg.LongQuantity.IsZero() && leg.ShortQuantity.IsZero() {
			continue
		}
		if !leg.Market.UnderlyingPrice.IsPositive() || !leg.Market.MarkPrice.IsPositive() {
			return PortfolioResult{}, fmt.Errorf("invalid portfolio market for contract %d", leg.Contract.Id)
		}
		key := portfolioGroupKey{
			underlying: leg.Contract.UnderlyingSymbol,
			settleCoin: leg.Contract.SettleCoin,
			expireTime: leg.Contract.ExpireTime,
		}
		grouped[key] = append(grouped[key], leg)
	}

	result := PortfolioResult{}
	for _, group := range grouped {
		groupResult, err := evaluatePortfolioGroup(group, maintenance)
		if err != nil {
			return PortfolioResult{}, err
		}
		result.Requirement = result.Requirement.Add(groupResult.Requirement)
		result.ScenarioLoss = result.ScenarioLoss.Add(groupResult.ScenarioLoss)
		result.ShortFloor = result.ShortFloor.Add(groupResult.ShortFloor)
	}
	result.Requirement = result.Requirement.Round(16)
	result.ScenarioLoss = result.ScenarioLoss.Round(16)
	result.ShortFloor = result.ShortFloor.Round(16)
	return result, nil
}

func evaluatePortfolioGroup(legs []PortfolioLeg, maintenance bool) (PortfolioResult, error) {
	if len(legs) == 0 {
		return PortfolioResult{}, nil
	}
	spot := legs[0].Market.UnderlyingPrice
	shockRate := decimal.Zero
	scenarios := []decimal.Decimal{decimal.Zero, spot, spot.Mul(decimal.NewFromInt(2))}
	currentValue := decimal.Zero
	shortFloor := decimal.Zero

	for _, leg := range legs {
		if !samePortfolioSpot(spot, leg.Market.UnderlyingPrice) {
			return PortfolioResult{}, fmt.Errorf("inconsistent underlying price in portfolio group")
		}
		rate := leg.Contract.InitialMarginRate
		if maintenance {
			rate = leg.Contract.MaintenanceMarginRate
		}
		shockRate = decimal.Max(shockRate, rate)
		scenarios = append(scenarios, leg.Contract.StrikePrice)

		multiplier := portfolioMultiplier(leg.Contract)
		netQuantity := leg.LongQuantity.Sub(leg.ShortQuantity)
		currentValue = currentValue.Add(leg.Market.MarkPrice.Mul(netQuantity).Mul(multiplier))

		netShort := decimal.Max(leg.ShortQuantity.Sub(leg.LongQuantity), decimal.Zero)
		if netShort.IsPositive() {
			floorBase := spot
			if leg.Contract.OptionType == int64(option.OptionType_OPTION_TYPE_PUT) {
				floorBase = leg.Contract.StrikePrice
			}
			shortFloor = shortFloor.Add(
				floorBase.Mul(netShort).Mul(multiplier).Mul(leg.Contract.MinMarginRate),
			)
		}
	}
	if !shockRate.IsPositive() {
		return PortfolioResult{}, errors.New("portfolio shock rate must be positive")
	}
	scenarios = append(
		scenarios,
		decimal.Max(spot.Mul(decimal.NewFromInt(1).Sub(shockRate)), decimal.Zero),
		spot.Mul(decimal.NewFromInt(1).Add(shockRate)),
	)
	sort.Slice(scenarios, func(i, j int) bool { return scenarios[i].LessThan(scenarios[j]) })

	maxLoss := decimal.Zero
	var previous *decimal.Decimal
	for _, scenario := range scenarios {
		if previous != nil && scenario.Equal(*previous) {
			continue
		}
		value := decimal.Zero
		for _, leg := range legs {
			netQuantity := leg.LongQuantity.Sub(leg.ShortQuantity)
			intrinsic := portfolioIntrinsic(leg.Contract, scenario)
			value = value.Add(intrinsic.Mul(netQuantity).Mul(portfolioMultiplier(leg.Contract)))
		}
		maxLoss = decimal.Max(maxLoss, currentValue.Sub(value))
		current := scenario
		previous = &current
	}
	maxLoss = decimal.Max(maxLoss, decimal.Zero).Round(16)
	shortFloor = shortFloor.Round(16)
	return PortfolioResult{
		Requirement:  decimal.Max(maxLoss, shortFloor).Round(16),
		ScenarioLoss: maxLoss,
		ShortFloor:   shortFloor,
	}, nil
}

func portfolioIntrinsic(contract *models.TOptionContract, spot decimal.Decimal) decimal.Decimal {
	if contract.OptionType == int64(option.OptionType_OPTION_TYPE_CALL) {
		return decimal.Max(spot.Sub(contract.StrikePrice), decimal.Zero)
	}
	return decimal.Max(contract.StrikePrice.Sub(spot), decimal.Zero)
}

func portfolioMultiplier(contract *models.TOptionContract) decimal.Decimal {
	if contract.Multiplier.IsPositive() {
		return contract.Multiplier
	}
	if contract.ContractUnit.IsPositive() {
		return contract.ContractUnit
	}
	return decimal.NewFromInt(1)
}

func samePortfolioSpot(left, right decimal.Decimal) bool {
	if left.Equal(right) {
		return true
	}
	base := decimal.Max(left.Abs(), right.Abs())
	if base.IsZero() {
		return true
	}
	return left.Sub(right).Abs().Div(base).LessThanOrEqual(decimal.RequireFromString("0.0001"))
}
