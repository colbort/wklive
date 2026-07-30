package tasklogic

import (
	"testing"

	"wklive/proto/option"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
)

func TestValidateOptionSettlementBalance(t *testing.T) {
	if err := validateOptionSettlementBalance(1, optionSettlementSummary{
		totalCredit: decimal.NewFromInt(10),
		totalDebit:  decimal.NewFromInt(10),
	}); err != nil {
		t.Fatalf("balanced settlement rejected: %v", err)
	}

	if err := validateOptionSettlementBalance(1, optionSettlementSummary{
		totalCredit: decimal.NewFromInt(10),
		totalDebit:  decimal.NewFromInt(9),
	}); err == nil {
		t.Fatal("unbalanced settlement must be rejected")
	}
}

func TestPhysicalLongAssetLegs(t *testing.T) {
	contract := &models.TOptionContract{
		OptionType:     int64(option.OptionType_OPTION_TYPE_CALL),
		UnderlyingCoin: "BTC", SettleCoin: "USDT",
	}
	debitCoin, debit, creditCoin, credit := physicalLongAssetLegs(
		contract, decimal.NewFromInt(2), decimal.NewFromInt(200),
	)
	if debitCoin != "USDT" || !debit.Equal(decimal.NewFromInt(200)) ||
		creditCoin != "BTC" || !credit.Equal(decimal.NewFromInt(2)) {
		t.Fatalf("unexpected call delivery legs: %s %s -> %s %s", debitCoin, debit, creditCoin, credit)
	}
	contract.OptionType = int64(option.OptionType_OPTION_TYPE_PUT)
	debitCoin, debit, creditCoin, credit = physicalLongAssetLegs(
		contract, decimal.NewFromInt(2), decimal.NewFromInt(200),
	)
	if debitCoin != "BTC" || !debit.Equal(decimal.NewFromInt(2)) ||
		creditCoin != "USDT" || !credit.Equal(decimal.NewFromInt(200)) {
		t.Fatalf("unexpected put delivery legs: %s %s -> %s %s", debitCoin, debit, creditCoin, credit)
	}
}
