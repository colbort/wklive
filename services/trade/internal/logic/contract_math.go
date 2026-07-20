package logic

import (
	"fmt"

	"wklive/proto/trade"

	"github.com/shopspring/decimal"
)

const contractAmountScale int32 = 18

type contractTradeValues struct {
	// QuoteNotional is used by order/fill statistics and risk tiers.
	QuoteNotional decimal.Decimal
	// SettlementNotional is denominated in the settlement/margin asset.
	SettlementNotional decimal.Decimal
}

func calculateContractTradeValues(valueType int64, qty, contractSize, price decimal.Decimal) (contractTradeValues, error) {
	if !qty.IsPositive() || !contractSize.IsPositive() || !price.IsPositive() {
		return contractTradeValues{}, fmt.Errorf("contract qty, size and price must be positive")
	}
	contractValue := qty.Mul(contractSize)
	switch trade.ContractValueType(valueType) {
	case trade.ContractValueType_CONTRACT_VALUE_TYPE_LINEAR:
		notional := contractValue.Mul(price)
		return contractTradeValues{QuoteNotional: roundContractDebit(notional), SettlementNotional: roundContractDebit(notional)}, nil
	case trade.ContractValueType_CONTRACT_VALUE_TYPE_INVERSE:
		return contractTradeValues{QuoteNotional: roundContractDebit(contractValue), SettlementNotional: roundContractDebit(contractValue.Div(price))}, nil
	default:
		return contractTradeValues{}, fmt.Errorf("unsupported contract value type: %d", valueType)
	}
}

func calculateContractMargin(values contractTradeValues, leverage int64) (decimal.Decimal, error) {
	if leverage <= 0 {
		return decimal.Zero, fmt.Errorf("leverage must be positive")
	}
	return roundContractDebit(values.SettlementNotional.Div(decimal.NewFromInt(leverage))), nil
}

func calculateContractFee(values contractTradeValues, feeRate decimal.Decimal) decimal.Decimal {
	if !feeRate.IsPositive() {
		return decimal.Zero
	}
	return roundContractDebit(values.SettlementNotional.Mul(feeRate))
}

// Debits are rounded away from zero so a required margin/fee is never
// understated. Credits and PnL use roundContractCredit (towards zero).
func roundContractDebit(value decimal.Decimal) decimal.Decimal {
	if !value.IsPositive() {
		return decimal.Zero
	}
	return value.RoundCeil(contractAmountScale)
}

func roundContractCredit(value decimal.Decimal) decimal.Decimal {
	return value.Truncate(contractAmountScale)
}
