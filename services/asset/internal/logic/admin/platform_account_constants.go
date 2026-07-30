package adminlogic

const (
	insuranceFundAccountType     = "INSURANCE_FUND"
	fundingDifferenceAccountType = "FUNDING_DIFFERENCE"
	feeRevenueAccountType        = "FEE_REVENUE"
	optionBackstopAccountType    = "OPTION_BACKSTOP"
)

func isConfigurablePlatformAccountType(accountType string) bool {
	switch accountType {
	case insuranceFundAccountType, fundingDifferenceAccountType, feeRevenueAccountType,
		optionBackstopAccountType:
		return true
	default:
		return false
	}
}
