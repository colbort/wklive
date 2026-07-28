package priceengine

import (
	"errors"
	"sort"
	"wklive/proto/itick"

	"github.com/shopspring/decimal"
)

type Input struct {
	Price      decimal.Decimal `json:"price"`
	Weight     decimal.Decimal `json:"weight"`
	SnapshotID string          `json:"snapshot_id"`
}

func Calculate(algorithm itick.PriceAlgorithm, inputs []Input) (decimal.Decimal, error) {
	if len(inputs) == 0 {
		return decimal.Zero, errors.New("price formula has no inputs")
	}
	for _, in := range inputs {
		if !in.Price.IsPositive() {
			return decimal.Zero, errors.New("price formula input must be positive")
		}
	}
	switch algorithm {
	case itick.PriceAlgorithm_PRICE_ALGORITHM_WEIGHTED_MEAN:
		total, weights := decimal.Zero, decimal.Zero
		for _, in := range inputs {
			if !in.Weight.IsPositive() {
				return decimal.Zero, errors.New("weight must be positive")
			}
			total = total.Add(in.Price.Mul(in.Weight))
			weights = weights.Add(in.Weight)
		}
		return total.Div(weights), nil
	case itick.PriceAlgorithm_PRICE_ALGORITHM_MEDIAN:
		prices := make([]decimal.Decimal, len(inputs))
		for i, in := range inputs {
			prices[i] = in.Price
		}
		sort.Slice(prices, func(i, j int) bool { return prices[i].LessThan(prices[j]) })
		n := len(prices)
		if n%2 == 1 {
			return prices[n/2], nil
		}
		return prices[n/2-1].Add(prices[n/2]).Div(decimal.NewFromInt(2)), nil
	case itick.PriceAlgorithm_PRICE_ALGORITHM_PREMIUM_RATE:
		if len(inputs) != 2 {
			return decimal.Zero, errors.New("premium rate requires mark and index inputs")
		}
		return inputs[0].Price.Sub(inputs[1].Price).Div(inputs[1].Price), nil
	case itick.PriceAlgorithm_PRICE_ALGORITHM_INDEX_BASIS:
		price, _, _, err := CalculateIndexBasis(inputs, 0)
		return price, err
	default:
		return decimal.Zero, errors.New("unsupported price formula algorithm")
	}
}

// CalculateIndexBasis derives a production mark price from an authoritative
// index and a perpetual-market quote. The raw basis is capped symmetrically so
// a dislocated venue cannot drag the mark beyond the configured risk bound.
func CalculateIndexBasis(inputs []Input, maxBasisBps int64) (price, rawBasis, appliedBasis decimal.Decimal, err error) {
	if len(inputs) != 2 && len(inputs) != 3 {
		return decimal.Zero, decimal.Zero, decimal.Zero, errors.New("index basis requires index, perpetual quote and optional previous mark inputs")
	}
	if !inputs[0].Price.IsPositive() || !inputs[1].Price.IsPositive() ||
		len(inputs) == 3 && !inputs[2].Price.IsPositive() {
		return decimal.Zero, decimal.Zero, decimal.Zero, errors.New("index basis inputs must be positive")
	}
	if maxBasisBps <= 0 || maxBasisBps > 10000 {
		return decimal.Zero, decimal.Zero, decimal.Zero, errors.New("index basis requires max_deviation_bps in (0,10000]")
	}
	indexPrice := inputs[0].Price
	rawBasis = inputs[1].Price.Sub(indexPrice).Div(indexPrice)
	limit := decimal.NewFromInt(maxBasisBps).Div(decimal.NewFromInt(10000))
	appliedBasis = rawBasis
	if appliedBasis.GreaterThan(limit) {
		appliedBasis = limit
	} else if appliedBasis.LessThan(limit.Neg()) {
		appliedBasis = limit.Neg()
	}
	price = indexPrice.Mul(decimal.NewFromInt(1).Add(appliedBasis))
	if len(inputs) == 3 {
		currentWeight, previousWeight := inputs[1].Weight, inputs[2].Weight
		if !currentWeight.IsPositive() || !previousWeight.IsPositive() {
			return decimal.Zero, rawBasis, appliedBasis, errors.New("index basis smoothing weights must be positive")
		}
		price = price.Mul(currentWeight).
			Add(inputs[2].Price.Mul(previousWeight)).
			Div(currentWeight.Add(previousWeight))
	}
	if !price.IsPositive() {
		return decimal.Zero, rawBasis, appliedBasis, errors.New("index basis produced non-positive mark")
	}
	return price, rawBasis, appliedBasis, nil
}
