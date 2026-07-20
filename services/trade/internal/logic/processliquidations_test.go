package logic

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestADLTakeoverQtyCapsAtBankruptRemainingQty(t *testing.T) {
	got := adlTakeoverQty(decimal.NewFromInt(100), decimal.NewFromInt(3), decimal.NewFromInt(50), decimal.NewFromInt(2))
	if !got.Equal(decimal.NewFromInt(3)) {
		t.Fatalf("takeover qty = %s, want 3", got)
	}
}

func TestADLTakeoverQtyRejectsInvalidInputs(t *testing.T) {
	if got := adlTakeoverQty(decimal.NewFromInt(10), decimal.NewFromInt(10), decimal.NewFromInt(5), decimal.Zero); !got.IsZero() {
		t.Fatalf("takeover qty = %s, want 0", got)
	}
}

func TestADLMarginReleaseKeepsMarginBucketsSeparate(t *testing.T) {
	position, isolated := adlMarginRelease(decimal.NewFromInt(80), decimal.NewFromInt(20), decimal.NewFromInt(2), decimal.NewFromInt(10))
	if !position.Equal(decimal.NewFromInt(16)) || !isolated.Equal(decimal.NewFromInt(4)) {
		t.Fatalf("released position=%s isolated=%s, want 16 and 4", position, isolated)
	}
}
