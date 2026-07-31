package adminlogic

import (
	"testing"

	"wklive/proto/option"
	"wklive/services/option/models"
)

func TestValidateSettlementPriceReviewEnforcesFourEyes(t *testing.T) {
	item := &models.TOptionSettlementPrice{
		Id: 7, Status: int64(option.SettlementPriceStatus_SETTLEMENT_PRICE_STATUS_PENDING),
		CreatedBy: 100,
	}
	if err := validateSettlementPriceReview(item, item, 101); err != nil {
		t.Fatalf("independent reviewer rejected: %v", err)
	}
	if err := validateSettlementPriceReview(item, item, 100); err == nil {
		t.Fatal("creator was allowed to approve own correction")
	}
	older := *item
	older.Id = 6
	if err := validateSettlementPriceReview(&older, item, 101); err == nil {
		t.Fatal("non-latest version was allowed to be reviewed")
	}
	item.Status = int64(option.SettlementPriceStatus_SETTLEMENT_PRICE_STATUS_CONFIRMED)
	if err := validateSettlementPriceReview(item, item, 101); err == nil {
		t.Fatal("confirmed version was allowed to be reviewed again")
	}
}
