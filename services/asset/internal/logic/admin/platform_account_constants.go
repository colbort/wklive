package adminlogic

const (
	insuranceFundAccountType     = "INSURANCE_FUND"
	fundingDifferenceAccountType = "FUNDING_DIFFERENCE"
	feeRevenueAccountType        = "FEE_REVENUE"
	optionBackstopAccountType    = "OPTION_BACKSTOP"
	stakingRewardAccountType     = "STAKING_REWARD"
)

func isConfigurablePlatformAccountType(accountType string) bool {
	switch accountType {
	case insuranceFundAccountType, fundingDifferenceAccountType, feeRevenueAccountType,
		optionBackstopAccountType, stakingRewardAccountType:
		return true
	default:
		return false
	}
}
