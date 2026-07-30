package tasklogic

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestValidateEarlyExerciseBalance(t *testing.T) {
	if err := validateEarlyExerciseBalance(
		decimal.NewFromInt(100),
		decimal.NewFromInt(100),
		decimal.NewFromInt(2),
		decimal.NewFromInt(98),
	); err != nil {
		t.Fatalf("balanced early exercise rejected: %v", err)
	}
	if err := validateEarlyExerciseBalance(
		decimal.NewFromInt(99),
		decimal.NewFromInt(100),
		decimal.NewFromInt(2),
		decimal.NewFromInt(98),
	); err == nil {
		t.Fatal("unbalanced early exercise accepted")
	}
}
