package tasklogic

import (
	"errors"
	"testing"
)

func TestValidateFinalDeliveryPriceFact(t *testing.T) {
	quote := &marketQuoteSnapshot{
		SnapshotID:     "delivery-snapshot-1",
		FormulaVersion: "delivery-median-v2",
		Confirmed:      true,
	}
	algorithm, version, err := validateFinalDeliveryPriceFact("delivery-median-v2", quote, []*marketQuoteSnapshot{quote})
	if err != nil {
		t.Fatal(err)
	}
	if algorithm != "delivery-median-v2" || version != "delivery-median-v2" {
		t.Fatalf("unexpected audit facts: algorithm=%s version=%s", algorithm, version)
	}
}

func TestValidateFinalDeliveryPriceFactRejectsTradeSideAggregation(t *testing.T) {
	quote := &marketQuoteSnapshot{SnapshotID: "a", FormulaVersion: "v1", Confirmed: true}
	if _, _, err := validateFinalDeliveryPriceFact("median-v1", quote, []*marketQuoteSnapshot{
		quote,
		{SnapshotID: "b", FormulaVersion: "v1", Confirmed: true},
	}); err == nil {
		t.Fatal("Trade must not aggregate multiple final DELIVERY snapshots")
	}
}

func TestValidateFinalDeliveryPriceFactRequiresAuditMetadata(t *testing.T) {
	tests := []struct {
		name       string
		configured string
		quote      *marketQuoteSnapshot
	}{
		{name: "missing snapshot", configured: "median-v1"},
		{name: "unconfirmed", configured: "median-v1", quote: &marketQuoteSnapshot{SnapshotID: "a", FormulaVersion: "v1"}},
		{name: "missing formula version", configured: "median-v1", quote: &marketQuoteSnapshot{SnapshotID: "a", Confirmed: true}},
		{name: "missing configured algorithm", quote: &marketQuoteSnapshot{SnapshotID: "a", FormulaVersion: "v1", Confirmed: true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidates := []*marketQuoteSnapshot{tt.quote}
			if tt.quote == nil {
				candidates = nil
			}
			if _, _, err := validateFinalDeliveryPriceFact(tt.configured, tt.quote, candidates); err == nil {
				t.Fatal("invalid final delivery price fact was accepted")
			}
		})
	}
}

func TestValidateFinalDeliveryPriceFactRequiresConfiguredVersionMatch(t *testing.T) {
	quote := &marketQuoteSnapshot{
		SnapshotID:     "delivery-snapshot-1",
		FormulaVersion: "delivery-v2",
		Confirmed:      true,
	}
	if _, _, err := validateFinalDeliveryPriceFact(
		"delivery-v1", quote, []*marketQuoteSnapshot{quote},
	); err == nil {
		t.Fatal("mismatched delivery formula version was accepted")
	}
}

func TestDeliveryPriceUnavailableClassification(t *testing.T) {
	cause := errors.New("missing final quote")
	err := deliveryPriceUnavailable(cause)
	if !errors.Is(err, errDeliveryPriceUnavailable) {
		t.Fatalf("delivery input error lost classification: %v", err)
	}
	if errors.Is(errors.New("database unavailable"), errDeliveryPriceUnavailable) {
		t.Fatal("unrelated infrastructure error was classified as delivery input")
	}
}
