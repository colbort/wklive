package helpers

import (
	"strings"
	"testing"

	"wklive/proto/option"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
)

func settlementEvidenceFixture() (*models.TOptionContract, *models.TOptionSettlementPrice) {
	contract := &models.TOptionContract{
		Id: 10, TenantId: 20, ExpireTime: 1000,
		SettlementPriceSource:   SettlementPriceSourceAutomatic,
		SettlementPriceMethod:   SettlementPriceMethodAutomatic,
		SettlementWindowSeconds: 60, SettlementMinSamples: 3,
	}
	price := &models.TOptionSettlementPrice{
		TenantId: 20, ContractId: 10,
		PriceSource: SettlementPriceSourceAutomatic, CalculationMethod: SettlementPriceMethodAutomatic,
		WindowStart: 940, WindowEnd: 1000, SampleCount: 3,
		DeliveryPrice: decimal.NewFromInt(101), SourceSnapshotIds: `["a","b","c"]`,
		Status: int64(option.SettlementPriceStatus_SETTLEMENT_PRICE_STATUS_PENDING),
	}
	return contract, price
}

func TestValidateSettlementPriceEvidenceAutomaticAndManual(t *testing.T) {
	contract, price := settlementEvidenceFixture()
	if err := ValidateSettlementPriceEvidence(contract, price, false); err != nil {
		t.Fatalf("valid automatic pending evidence rejected: %v", err)
	}
	price.Status = int64(option.SettlementPriceStatus_SETTLEMENT_PRICE_STATUS_CONFIRMED)
	price.ConfirmedBy, price.ConfirmedAt = 7, 1001
	if err := ValidateSettlementPriceEvidence(contract, price, true); err != nil {
		t.Fatalf("valid automatic confirmed evidence rejected: %v", err)
	}

	manual := *price
	manual.PriceSource = SettlementPriceSourceManual
	manual.CalculationMethod = SettlementPriceMethodManual
	manual.SampleCount = 1
	manual.SourceSnapshotIds = `["external-case-1"]`
	manual.CreatedBy, manual.ConfirmedBy = 8, 9
	if err := ValidateSettlementPriceEvidence(contract, &manual, true); err != nil {
		t.Fatalf("valid manual correction rejected: %v", err)
	}
	manual.ConfirmedBy = manual.CreatedBy
	if err := ValidateSettlementPriceEvidence(contract, &manual, true); err == nil {
		t.Fatal("manual creator must not confirm the same version")
	}
}

func TestValidateSettlementPriceEvidenceRejectsMalformedEvidence(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*models.TOptionContract, *models.TOptionSettlementPrice)
	}{
		{"wrong window", func(_ *models.TOptionContract, p *models.TOptionSettlementPrice) { p.WindowStart++ }},
		{"count mismatch", func(_ *models.TOptionContract, p *models.TOptionSettlementPrice) { p.SampleCount = 2 }},
		{"duplicate ids", func(_ *models.TOptionContract, p *models.TOptionSettlementPrice) {
			p.SourceSnapshotIds = `["a","a","c"]`
		}},
		{"blank id", func(_ *models.TOptionContract, p *models.TOptionSettlementPrice) {
			p.SourceSnapshotIds = `["a"," ","c"]`
		}},
		{"insufficient", func(c *models.TOptionContract, p *models.TOptionSettlementPrice) { c.SettlementMinSamples = 4 }},
		{"admin automatic", func(_ *models.TOptionContract, p *models.TOptionSettlementPrice) { p.CreatedBy = 1 }},
		{"unknown type", func(_ *models.TOptionContract, p *models.TOptionSettlementPrice) { p.PriceSource = "guess" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contract, price := settlementEvidenceFixture()
			tt.mutate(contract, price)
			if err := ValidateSettlementPriceEvidence(contract, price, false); err == nil {
				t.Fatal("malformed settlement evidence accepted")
			}
		})
	}
}

func TestNormalizeAndCompareSettlementPriceEvidence(t *testing.T) {
	ids, canonical, err := NormalizeSettlementPriceSourceIDs(`[" a ","b"]`)
	if err != nil || len(ids) != 2 || ids[0] != "a" || canonical != `["a","b"]` {
		t.Fatalf("unexpected normalized evidence ids=%v json=%s err=%v", ids, canonical, err)
	}
	_, left := settlementEvidenceFixture()
	right := *left
	right.SourceSnapshotIds = `[ "a", "b", "c" ]`
	if !SameSettlementPriceEvidence(left, &right) {
		t.Fatal("equivalent evidence JSON should compare equal")
	}
	right.SourceSnapshotIds = `["a","b","d"]`
	if SameSettlementPriceEvidence(left, &right) {
		t.Fatal("different evidence must not compare equal")
	}
	if _, _, err := NormalizeSettlementPriceSourceIDs(`["` + strings.Repeat("x", 129) + `"]`); err == nil {
		t.Fatal("oversized evidence identity must be rejected before database truncation")
	}
}
