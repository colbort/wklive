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
	default:
		return decimal.Zero, errors.New("unsupported price formula algorithm")
	}
}
