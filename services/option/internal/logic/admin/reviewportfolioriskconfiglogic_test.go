package adminlogic

import (
	"testing"

	"wklive/proto/option"
	"wklive/services/option/models"
)

func TestValidatePortfolioRiskConfigReviewEnforcesFourEyesAndLatestVersion(t *testing.T) {
	item := &models.TOptionPortfolioRiskConfig{
		Id: 10, CreatedBy: 100,
		Status: int64(option.PortfolioRiskConfigStatus_PORTFOLIO_RISK_CONFIG_STATUS_PENDING),
	}
	if err := validatePortfolioRiskConfigReview(item, item, 101); err != nil {
		t.Fatal(err)
	}
	if err := validatePortfolioRiskConfigReview(item, item, 100); err == nil {
		t.Fatal("creator must not approve their own portfolio risk config")
	}
	older := *item
	older.Id = 9
	if err := validatePortfolioRiskConfigReview(&older, item, 101); err == nil {
		t.Fatal("non-latest portfolio risk config must not be reviewed")
	}
	item.Status = int64(option.PortfolioRiskConfigStatus_PORTFOLIO_RISK_CONFIG_STATUS_APPROVED)
	if err := validatePortfolioRiskConfigReview(item, item, 101); err == nil {
		t.Fatal("approved portfolio risk config must not be reviewed again")
	}
}

func TestPortfolioRiskConfigApprovalRequiresFutureEffectiveTime(t *testing.T) {
	for _, tt := range []struct {
		name          string
		effectiveFrom int64
		now           int64
		wantValid     bool
	}{
		{name: "future", effectiveFrom: 101, now: 100, wantValid: true},
		{name: "exact boundary", effectiveFrom: 100, now: 100, wantValid: false},
		{name: "retroactive", effectiveFrom: 99, now: 100, wantValid: false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			valid := portfolioRiskConfigEffectiveTimeValid(tt.effectiveFrom, tt.now)
			if valid != tt.wantValid {
				t.Fatalf("effectiveFrom/now=%d/%d valid=%t want=%t",
					tt.effectiveFrom, tt.now, valid, tt.wantValid)
			}
		})
	}
}
