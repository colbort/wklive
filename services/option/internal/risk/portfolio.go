package risk

import (
	"errors"
	"fmt"
	"sort"
	"strings"

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
	Requirement          decimal.Decimal
	ScenarioLoss         decimal.Decimal
	ShortFloor           decimal.Decimal
	ConcentrationAddon   decimal.Decimal
	LiquidityAddon       decimal.Decimal
	GrossShortRiskAmount decimal.Decimal
}

// PortfolioConfig is an approved model parameter snapshot. Production callers
// must resolve it from t_option_portfolio_risk_config for the calculation time.
type PortfolioConfig struct {
	InitialShockRate       decimal.Decimal
	MaintenanceShockRate   decimal.Decimal
	ScenarioShocks         []decimal.Decimal
	ConcentrationThreshold decimal.Decimal
	ConcentrationAddonRate decimal.Decimal
	LiquidityAddonRate     decimal.Decimal
}

type portfolioGroupKey struct {
	underlying string
	settleCoin string
	expireTime int64
}

// EvaluatePortfolio implements PORTFOLIO_RISK_METHOD_EXPIRY_SCENARIO_V1 with
// an explicit, approved parameter snapshot.
//
// Offsets are deliberately limited to contracts with the same underlying,
// settlement coin and expiry. Each group is valued at zero spot, configured
// down/up shocks, current spot, every strike, and 2x spot. The largest loss is
// bounded below by a minimum requirement for net naked shorts. This keeps the
// method deterministic and prevents unrelated expiries or assets from
// receiving an implicit correlation credit.
func EvaluatePortfolio(
	legs []PortfolioLeg,
	maintenance bool,
	config PortfolioConfig,
) (PortfolioResult, error) {
	if err := ValidatePortfolioConfig(config); err != nil {
		return PortfolioResult{}, err
	}
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
		groupResult, err := evaluatePortfolioGroup(group, maintenance, config)
		if err != nil {
			return PortfolioResult{}, err
		}
		result.Requirement = result.Requirement.Add(groupResult.Requirement)
		result.ScenarioLoss = result.ScenarioLoss.Add(groupResult.ScenarioLoss)
		result.ShortFloor = result.ShortFloor.Add(groupResult.ShortFloor)
		result.ConcentrationAddon = result.ConcentrationAddon.Add(groupResult.ConcentrationAddon)
		result.LiquidityAddon = result.LiquidityAddon.Add(groupResult.LiquidityAddon)
		result.GrossShortRiskAmount = result.GrossShortRiskAmount.Add(groupResult.GrossShortRiskAmount)
	}
	result.Requirement = result.Requirement.Round(16)
	result.ScenarioLoss = result.ScenarioLoss.Round(16)
	result.ShortFloor = result.ShortFloor.Round(16)
	result.ConcentrationAddon = result.ConcentrationAddon.Round(16)
	result.LiquidityAddon = result.LiquidityAddon.Round(16)
	result.GrossShortRiskAmount = result.GrossShortRiskAmount.Round(16)
	return result, nil
}

func evaluatePortfolioGroup(
	legs []PortfolioLeg,
	maintenance bool,
	config PortfolioConfig,
) (PortfolioResult, error) {
	if len(legs) == 0 {
		return PortfolioResult{}, nil
	}
	spot := legs[0].Market.UnderlyingPrice
	shockRate := config.InitialShockRate
	if maintenance {
		shockRate = config.MaintenanceShockRate
	}
	scenarios := []decimal.Decimal{decimal.Zero, spot, spot.Mul(decimal.NewFromInt(2))}
	for _, relativeShock := range config.ScenarioShocks {
		scenarios = append(scenarios,
			decimal.Max(spot.Mul(decimal.NewFromInt(1).Add(relativeShock)), decimal.Zero))
	}
	currentValue := decimal.Zero
	shortFloor := decimal.Zero
	grossShortRiskAmount := decimal.Zero

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
			riskAmount := floorBase.Mul(netShort).Mul(multiplier)
			grossShortRiskAmount = grossShortRiskAmount.Add(riskAmount)
			shortFloor = shortFloor.Add(
				riskAmount.Mul(leg.Contract.MinMarginRate),
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
	grossShortRiskAmount = grossShortRiskAmount.Round(16)
	concentrationAddon := decimal.Max(
		grossShortRiskAmount.Sub(config.ConcentrationThreshold),
		decimal.Zero,
	).Mul(config.ConcentrationAddonRate).Round(16)
	liquidityAddon := grossShortRiskAmount.Mul(config.LiquidityAddonRate).Round(16)
	return PortfolioResult{
		Requirement: decimal.Max(maxLoss, shortFloor).
			Add(concentrationAddon).Add(liquidityAddon).Round(16),
		ScenarioLoss:         maxLoss,
		ShortFloor:           shortFloor,
		ConcentrationAddon:   concentrationAddon,
		LiquidityAddon:       liquidityAddon,
		GrossShortRiskAmount: grossShortRiskAmount,
	}, nil
}

// ParseScenarioShocks validates and canonicalizes a comma-separated relative
// shock set. A governed V1 set must cover total underlying loss (-100%) and at
// least a five-times spot scenario (+400%).
func ParseScenarioShocks(value string) ([]decimal.Decimal, string, error) {
	parts := strings.Split(value, ",")
	unique := make(map[string]decimal.Decimal)
	hasTotalLoss := false
	hasFiveTimes := false
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		shock, err := decimal.NewFromString(part)
		if err != nil || shock.LessThan(decimal.NewFromInt(-1)) || shock.GreaterThan(decimal.NewFromInt(10)) {
			return nil, "", fmt.Errorf("invalid portfolio scenario shock %q", part)
		}
		shock = shock.Round(10)
		unique[shock.String()] = shock
		hasTotalLoss = hasTotalLoss || shock.Equal(decimal.NewFromInt(-1))
		hasFiveTimes = hasFiveTimes || shock.GreaterThanOrEqual(decimal.NewFromInt(4))
	}
	if !hasTotalLoss || !hasFiveTimes {
		return nil, "", errors.New("portfolio scenarios must include -1 and a shock >= 4")
	}
	result := make([]decimal.Decimal, 0, len(unique))
	for _, shock := range unique {
		result = append(result, shock)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].LessThan(result[j]) })
	canonical := make([]string, 0, len(result))
	for _, shock := range result {
		canonical = append(canonical, shock.String())
	}
	return result, strings.Join(canonical, ","), nil
}

func ValidatePortfolioConfig(config PortfolioConfig) error {
	if !config.InitialShockRate.IsPositive() ||
		config.InitialShockRate.GreaterThan(decimal.NewFromInt(10)) ||
		!config.MaintenanceShockRate.IsPositive() ||
		config.MaintenanceShockRate.GreaterThan(config.InitialShockRate) {
		return errors.New("invalid portfolio initial or maintenance shock rate")
	}
	if config.ConcentrationThreshold.IsNegative() ||
		config.ConcentrationAddonRate.IsNegative() ||
		config.ConcentrationAddonRate.GreaterThan(decimal.NewFromInt(1)) ||
		config.LiquidityAddonRate.IsNegative() ||
		config.LiquidityAddonRate.GreaterThan(decimal.NewFromInt(1)) {
		return errors.New("invalid portfolio concentration or liquidity parameter")
	}
	_, _, err := ParseScenarioShocks(decimalSliceString(config.ScenarioShocks))
	return err
}

func PortfolioConfigFromModel(item *models.TOptionPortfolioRiskConfig) (PortfolioConfig, error) {
	if item == nil {
		return PortfolioConfig{}, errors.New("portfolio risk config is required")
	}
	shocks, _, err := ParseScenarioShocks(item.ScenarioShocks)
	if err != nil {
		return PortfolioConfig{}, err
	}
	config := PortfolioConfig{
		InitialShockRate:       item.InitialShockRate,
		MaintenanceShockRate:   item.MaintenanceShockRate,
		ScenarioShocks:         shocks,
		ConcentrationThreshold: item.ConcentrationThreshold,
		ConcentrationAddonRate: item.ConcentrationAddonRate,
		LiquidityAddonRate:     item.LiquidityAddonRate,
	}
	if err := ValidatePortfolioConfig(config); err != nil {
		return PortfolioConfig{}, err
	}
	return config, nil
}

func decimalSliceString(items []decimal.Decimal) string {
	values := make([]string, 0, len(items))
	for _, item := range items {
		values = append(values, item.String())
	}
	return strings.Join(values, ",")
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
