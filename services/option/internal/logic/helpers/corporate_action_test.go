package helpers

import (
	"testing"

	"wklive/services/option/models"

	"github.com/shopspring/decimal"
)

func TestConvertCorporateActionPositionPreservesCostBasis(t *testing.T) {
	position := &models.TOptionPosition{
		PositionQty:  decimal.RequireFromString("2"),
		AvailableQty: decimal.RequireFromString("2"),
		OpenAvgPrice: decimal.RequireFromString("10"),
	}
	source := &models.TOptionContract{Multiplier: decimal.RequireFromString("1")}
	successor := &models.TOptionContract{Multiplier: decimal.RequireFromString("0.5")}
	result, err := ConvertCorporateActionPosition(
		position, source, successor, decimal.RequireFromString("2"), decimal.RequireFromString("1"),
	)
	if err != nil {
		t.Fatal(err)
	}
	if !result.SuccessorQuantity.Equal(decimal.RequireFromString("4")) ||
		!result.SuccessorOpenAvgPrice.Equal(decimal.RequireFromString("10")) ||
		!result.CostBasisBefore.Equal(result.CostBasisAfter) {
		t.Fatalf("unexpected conversion: %+v", result)
	}
}

func TestConvertCorporateActionPositionRejectsInexactQuantity(t *testing.T) {
	position := &models.TOptionPosition{
		PositionQty:  decimal.RequireFromString("0.0000000000000001"),
		AvailableQty: decimal.RequireFromString("0.0000000000000001"),
		OpenAvgPrice: decimal.RequireFromString("1"),
	}
	_, err := ConvertCorporateActionPosition(
		position,
		&models.TOptionContract{Multiplier: decimal.NewFromInt(1)},
		&models.TOptionContract{Multiplier: decimal.NewFromInt(1)},
		decimal.NewFromInt(1), decimal.NewFromInt(3),
	)
	if !errorsIsCorporateActionInexact(err) {
		t.Fatalf("expected inexact error, got %v", err)
	}
}

func TestConvertCorporateActionPositionRejectsSuccessorStepRemainder(t *testing.T) {
	position := &models.TOptionPosition{
		PositionQty:  decimal.RequireFromString("1"),
		AvailableQty: decimal.RequireFromString("1"),
		OpenAvgPrice: decimal.RequireFromString("1"),
	}
	_, err := ConvertCorporateActionPosition(
		position,
		&models.TOptionContract{Multiplier: decimal.NewFromInt(1)},
		&models.TOptionContract{
			Multiplier: decimal.NewFromInt(1),
			QtyStep:    decimal.RequireFromString("0.3"),
		},
		decimal.NewFromInt(1), decimal.NewFromInt(1),
	)
	if !errorsIsCorporateActionInexact(err) {
		t.Fatalf("expected quantity step error, got %v", err)
	}
}

func TestParsePositiveCorporateActionInteger(t *testing.T) {
	if _, err := ParsePositiveCorporateActionInteger("100"); err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{"", "0", "-1", "1.5"} {
		if _, err := ParsePositiveCorporateActionInteger(value); err == nil {
			t.Fatalf("expected %q to fail", value)
		}
	}
}

func errorsIsCorporateActionInexact(err error) bool {
	return err == ErrCorporateActionInexact
}
