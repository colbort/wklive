package adminlogic

import (
	"testing"

	"wklive/proto/option"
	"wklive/services/option/models"
)

func tradeCorrectionFixture() (*models.TOptionTrade, *models.TOptionContract) {
	return &models.TOptionTrade{
			Id: 100, TenantId: 9, ContractId: 88,
			BuyUserId: 11, BuyAccountId: 111,
			SellUserId: 22, SellAccountId: 222,
			FeeCoin: "USDT",
		}, &models.TOptionContract{
			Id: 88, TenantId: 9,
			FeeUserId: 33, FeeAccountId: 333,
			InsuranceUserId: 44, InsuranceAccountId: 444,
		}
}

func TestValidateTradeCorrectionLegsRequiresExactBalance(t *testing.T) {
	trade, contract := tradeCorrectionFixture()
	legs, err := validateTradeCorrectionLegs([]*option.TradeCorrectionLegInput{
		{
			UserId: 11, AccountId: 111, Coin: "USDT",
			Direction: option.TradeCorrectionLegDirection_TRADE_CORRECTION_LEG_DIRECTION_DEBIT,
			Amount:    "10.25",
		},
		{
			UserId: 22, AccountId: 222, Coin: "USDT",
			Direction: option.TradeCorrectionLegDirection_TRADE_CORRECTION_LEG_DIRECTION_CREDIT,
			Amount:    "10.25",
		},
	}, trade, contract)
	if err != nil {
		t.Fatalf("balanced correction rejected: %v", err)
	}
	if len(legs) != 2 || legs[0].Amount.String() != "10.25" {
		t.Fatalf("unexpected validated legs: %+v", legs)
	}

	_, err = validateTradeCorrectionLegs([]*option.TradeCorrectionLegInput{
		{
			UserId: 11, AccountId: 111, Coin: "USDT",
			Direction: option.TradeCorrectionLegDirection_TRADE_CORRECTION_LEG_DIRECTION_DEBIT,
			Amount:    "10.25",
		},
		{
			UserId: 22, AccountId: 222, Coin: "USDT",
			Direction: option.TradeCorrectionLegDirection_TRADE_CORRECTION_LEG_DIRECTION_CREDIT,
			Amount:    "10.24",
		},
	}, trade, contract)
	if err == nil {
		t.Fatal("unbalanced correction must be rejected")
	}
}

func TestValidateTradeCorrectionLegsRestrictsParticipantsAndCoin(t *testing.T) {
	trade, contract := tradeCorrectionFixture()
	base := []*option.TradeCorrectionLegInput{
		{
			UserId: 11, AccountId: 111, Coin: "USDT",
			Direction: option.TradeCorrectionLegDirection_TRADE_CORRECTION_LEG_DIRECTION_DEBIT,
			Amount:    "1",
		},
		{
			UserId: 44, AccountId: 444, Coin: "USDT",
			Direction: option.TradeCorrectionLegDirection_TRADE_CORRECTION_LEG_DIRECTION_CREDIT,
			Amount:    "1",
		},
	}
	if _, err := validateTradeCorrectionLegs(base, trade, contract); err != nil {
		t.Fatalf("approved platform participant rejected: %v", err)
	}

	unauthorized := []*option.TradeCorrectionLegInput{base[0], {
		UserId: 99, AccountId: 999, Coin: "USDT",
		Direction: option.TradeCorrectionLegDirection_TRADE_CORRECTION_LEG_DIRECTION_CREDIT,
		Amount:    "1",
	}}
	if _, err := validateTradeCorrectionLegs(unauthorized, trade, contract); err == nil {
		t.Fatal("unrelated account must be rejected")
	}

	wrongCoin := []*option.TradeCorrectionLegInput{base[0], {
		UserId: 22, AccountId: 222, Coin: "BTC",
		Direction: option.TradeCorrectionLegDirection_TRADE_CORRECTION_LEG_DIRECTION_CREDIT,
		Amount:    "1",
	}}
	if _, err := validateTradeCorrectionLegs(wrongCoin, trade, contract); err == nil {
		t.Fatal("coin different from original trade must be rejected")
	}
}

func TestValidateTradeCorrectionLegsRejectsExcessPrecision(t *testing.T) {
	trade, contract := tradeCorrectionFixture()
	_, err := validateTradeCorrectionLegs([]*option.TradeCorrectionLegInput{
		{
			UserId: 11, AccountId: 111, Coin: "USDT",
			Direction: option.TradeCorrectionLegDirection_TRADE_CORRECTION_LEG_DIRECTION_DEBIT,
			Amount:    "0.00000000000000001",
		},
		{
			UserId: 22, AccountId: 222, Coin: "USDT",
			Direction: option.TradeCorrectionLegDirection_TRADE_CORRECTION_LEG_DIRECTION_CREDIT,
			Amount:    "0.00000000000000001",
		},
	}, trade, contract)
	if err == nil {
		t.Fatal("more than 16 decimal places must be rejected")
	}
}
