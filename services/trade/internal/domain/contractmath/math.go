package contractmath

import (
	"fmt"

	"wklive/proto/trade"

	"github.com/shopspring/decimal"
)

const contractAmountScale int32 = 18

type TradeValues struct {
	// QuoteNotional is used by order/fill statistics and risk tiers.
	QuoteNotional decimal.Decimal
	// SettlementNotional is denominated in the settlement/margin asset.
	SettlementNotional decimal.Decimal
}

func CalculateTradeValues(valueType int64, qty, contractSize, price decimal.Decimal) (TradeValues, error) {
	if !qty.IsPositive() || !contractSize.IsPositive() || !price.IsPositive() {
		return TradeValues{}, fmt.Errorf("contract qty, size and price must be positive")
	}
	contractValue := qty.Mul(contractSize)
	switch trade.ContractValueType(valueType) {
	case trade.ContractValueType_CONTRACT_VALUE_TYPE_LINEAR:
		notional := contractValue.Mul(price)
		return TradeValues{QuoteNotional: RoundDebit(notional), SettlementNotional: RoundDebit(notional)}, nil
	case trade.ContractValueType_CONTRACT_VALUE_TYPE_INVERSE:
		return TradeValues{QuoteNotional: RoundDebit(contractValue), SettlementNotional: RoundDebit(contractValue.Div(price))}, nil
	default:
		return TradeValues{}, fmt.Errorf("unsupported contract value type: %d", valueType)
	}
}

func CalculateMargin(values TradeValues, leverage int64) (decimal.Decimal, error) {
	if leverage <= 0 {
		return decimal.Zero, fmt.Errorf("leverage must be positive")
	}
	return RoundDebit(values.SettlementNotional.Div(decimal.NewFromInt(leverage))), nil
}

func CalculateFee(values TradeValues, feeRate decimal.Decimal) decimal.Decimal {
	if !feeRate.IsPositive() {
		return decimal.Zero
	}
	return RoundDebit(values.SettlementNotional.Mul(feeRate))
}

// Debits are rounded away from zero so a required margin/fee is never
// understated. Credits and PnL use RoundCredit (towards zero).
func RoundDebit(value decimal.Decimal) decimal.Decimal {
	if !value.IsPositive() {
		return decimal.Zero
	}
	return value.RoundCeil(contractAmountScale)
}

func RoundCredit(value decimal.Decimal) decimal.Decimal {
	return value.Truncate(contractAmountScale)
}
