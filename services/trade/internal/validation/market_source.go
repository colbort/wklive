package validation

import (
	"fmt"
	"strings"

	"wklive/proto/common"
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
	if common.ContractType(contractType) != common.ContractType_CONTRACT_TYPE_PERPETUAL {
		return nil
	}
	if strings.TrimSpace(value) != "premium-v1" {
		return fmt.Errorf("funding_rate_source must be premium-v1 for perpetual contracts")
	}
	return nil
}

// ContractPriceSources validates the authoritative market dimensions needed by
// each contract lifecycle. Delivery must never accept a bare symbol because
// price locking has no safe way to infer its category or market.
func ContractPriceSources(
	contractType int64,
	markPriceSource, settlementPriceSource string,
	settlementWindowSeconds int64,
	settlementPriceAlgorithm string,
) error {
	switch common.ContractType(contractType) {
	case common.ContractType_CONTRACT_TYPE_PERPETUAL:
		return AuthoritativeQuoteSources("mark_price_source", markPriceSource)
	case common.ContractType_CONTRACT_TYPE_DELIVERY:
		if err := AuthoritativeQuoteSources("settlement_price_source", settlementPriceSource); err != nil {
			return err
		}
		if settlementWindowSeconds <= 0 {
			return fmt.Errorf("settlement_window_seconds must be positive for delivery contracts")
		}
		if strings.TrimSpace(settlementPriceAlgorithm) == "" {
			return fmt.Errorf("settlement_price_algorithm is required for delivery contracts")
		}
		return nil
	default:
		return fmt.Errorf("unsupported contract type: %d", contractType)
	}
}
