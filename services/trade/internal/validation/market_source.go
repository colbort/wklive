package validation

import (
	"fmt"
	"strings"

	"wklive/proto/trade"
)

// AuthoritativeQuoteSources validates one or more market sources. Each source
// must identify the category, market and symbol used by the market archive.
func AuthoritativeQuoteSources(field, value string) error {
	sources := strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == '|' || r == ';'
	})
	if len(sources) == 0 {
		return fmt.Errorf("%s is required", field)
	}
	for _, source := range sources {
		parts := strings.Split(strings.TrimSpace(source), ":")
		if len(parts) != 3 ||
			strings.TrimSpace(parts[0]) == "" ||
			strings.TrimSpace(parts[1]) == "" ||
			strings.TrimSpace(parts[2]) == "" {
			return fmt.Errorf("%s must use category:market:symbol format: %q", field, source)
		}
	}
	return nil
}

func FundingRateSource(contractType int64, value string) error {
	if trade.ContractType(contractType) != trade.ContractType_CONTRACT_TYPE_PERPETUAL {
		return nil
	}
	if strings.TrimSpace(value) != "premium-v1" {
		return fmt.Errorf("funding_rate_source must be premium-v1 for perpetual contracts")
	}
	return nil
}
