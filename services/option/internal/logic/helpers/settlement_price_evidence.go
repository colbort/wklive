package helpers

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"wklive/proto/option"
	"wklive/services/option/models"
)

const (
	SettlementPriceSourceAutomatic = "authoritative-market"
	SettlementPriceMethodAutomatic = "MEDIAN"
	SettlementPriceSourceManual    = "manual-correction"
	SettlementPriceMethodManual    = "MANUAL"
)

// NormalizeSettlementPriceSourceIDs validates an evidence array and returns a
// stable JSON encoding. Snapshot references are evidence identities, so blanks
// and duplicates are rejected instead of silently changing sample weight.
func NormalizeSettlementPriceSourceIDs(raw string) ([]string, string, error) {
	var sourceIDs []string
	if err := json.Unmarshal([]byte(raw), &sourceIDs); err != nil || len(sourceIDs) == 0 {
		return nil, "", errors.New("settlement price source snapshot ids must be a non-empty JSON array")
	}
	seen := make(map[string]struct{}, len(sourceIDs))
	for i, sourceID := range sourceIDs {
		sourceID = strings.TrimSpace(sourceID)
		if sourceID == "" {
			return nil, "", errors.New("settlement price source snapshot id is required")
		}
		if len(sourceID) > 128 {
			return nil, "", errors.New("settlement price source snapshot id exceeds 128 bytes")
		}
		if _, exists := seen[sourceID]; exists {
			return nil, "", fmt.Errorf("duplicate settlement price source snapshot id: %s", sourceID)
		}
		seen[sourceID] = struct{}{}
		sourceIDs[i] = sourceID
	}
	canonical, err := json.Marshal(sourceIDs)
	if err != nil {
		return nil, "", err
	}
	return sourceIDs, string(canonical), nil
}

// ValidateSettlementPriceEvidence is the use-point fail-closed gate shared by
// automatic settlement and manual review. Database constraints remain a second
// line of defense; callers must not settle merely because status is CONFIRMED.
func ValidateSettlementPriceEvidence(
	contract *models.TOptionContract,
	price *models.TOptionSettlementPrice,
	requireConfirmed bool,
) error {
	if contract == nil || price == nil || contract.Id <= 0 ||
		price.ContractId != contract.Id || price.TenantId != contract.TenantId {
		return errors.New("settlement price contract identity mismatch")
	}
	if contract.ExpireTime <= 0 || contract.SettlementWindowSeconds <= 0 ||
		contract.SettlementMinSamples <= 0 {
		return errors.New("settlement price contract rule is incomplete")
	}
	if price.WindowStart != contract.ExpireTime-contract.SettlementWindowSeconds ||
		price.WindowEnd != contract.ExpireTime {
		return errors.New("settlement price window does not match contract")
	}
	if !price.DeliveryPrice.IsPositive() {
		return errors.New("settlement price must be positive")
	}
	sourceIDs, _, err := NormalizeSettlementPriceSourceIDs(price.SourceSnapshotIds)
	if err != nil {
		return err
	}
	if price.SampleCount != int64(len(sourceIDs)) {
		return errors.New("settlement price sample count does not match evidence")
	}

	source := strings.TrimSpace(price.PriceSource)
	method := strings.TrimSpace(price.CalculationMethod)
	switch {
	case source == SettlementPriceSourceAutomatic && method == SettlementPriceMethodAutomatic:
		if price.CreatedBy != 0 {
			return errors.New("automatic settlement price must be system-created")
		}
		if source != strings.ToLower(strings.TrimSpace(contract.SettlementPriceSource)) ||
			method != strings.ToUpper(strings.TrimSpace(contract.SettlementPriceMethod)) {
			return errors.New("automatic settlement price rule does not match contract")
		}
		if price.SampleCount < contract.SettlementMinSamples {
			return errors.New("automatic settlement price has insufficient samples")
		}
	case source == SettlementPriceSourceManual && method == SettlementPriceMethodManual:
		if price.CreatedBy <= 0 {
			return errors.New("manual settlement price requires a creator")
		}
	default:
		return errors.New("unsupported settlement price evidence type")
	}

	if requireConfirmed {
		if price.Status != int64(option.SettlementPriceStatus_SETTLEMENT_PRICE_STATUS_CONFIRMED) ||
			price.ConfirmedBy <= 0 || price.ConfirmedAt <= 0 {
			return errors.New("settlement price is not independently confirmed")
		}
		if price.CreatedBy > 0 && price.CreatedBy == price.ConfirmedBy {
			return errors.New("settlement price creator cannot confirm the same version")
		}
	} else if price.Status != int64(option.SettlementPriceStatus_SETTLEMENT_PRICE_STATUS_PENDING) {
		return errors.New("settlement price evidence is not pending")
	}
	return nil
}

func SameSettlementPriceEvidence(left, right *models.TOptionSettlementPrice) bool {
	if left == nil || right == nil {
		return false
	}
	_, leftJSON, leftErr := NormalizeSettlementPriceSourceIDs(left.SourceSnapshotIds)
	_, rightJSON, rightErr := NormalizeSettlementPriceSourceIDs(right.SourceSnapshotIds)
	return leftErr == nil && rightErr == nil &&
		left.PriceSource == right.PriceSource &&
		left.WindowStart == right.WindowStart && left.WindowEnd == right.WindowEnd &&
		left.SampleCount == right.SampleCount &&
		left.CalculationMethod == right.CalculationMethod &&
		left.DeliveryPrice.Equal(right.DeliveryPrice) && leftJSON == rightJSON
}
