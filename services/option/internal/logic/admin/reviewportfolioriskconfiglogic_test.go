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
