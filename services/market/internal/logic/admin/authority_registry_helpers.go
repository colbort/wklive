package adminlogic

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"wklive/proto/common"
	"wklive/proto/market"
	"wklive/services/market/models"
)

var authorityKindOrder = []string{
	"FINAL_QUOTE",
	"INDEX",
	"MARK",
	"FUNDING",
	"DELIVERY",
}

func normalizeAuthorityRegistry(
	authority, providerCode, producerType string,
	allowedKinds []string,
) (string, string, string, []string, string, error) {
	authority = strings.ToLower(strings.TrimSpace(authority))
	providerCode = strings.ToUpper(strings.TrimSpace(providerCode))
	producerType = strings.ToUpper(strings.TrimSpace(producerType))
	if !validAuthorityToken(authority, false) {
		return "", "", "", nil, "", errors.New("authority must use 1-32 lowercase letters, digits, '-' or '_'")
	}
	if !validAuthorityToken(providerCode, true) {
		return "", "", "", nil, "", errors.New("provider_code must use 1-32 uppercase letters, digits, '-' or '_'")
	}
	if !validAuthorityToken(producerType, true) {
		return "", "", "", nil, "", errors.New("producer_type must use 1-32 uppercase letters, digits, '-' or '_'")
	}
	requested := make(map[string]struct{}, len(allowedKinds))
	for _, kind := range allowedKinds {
		kind = strings.ToUpper(strings.TrimSpace(kind))
		if !isAuthorityKind(kind) {
			return "", "", "", nil, "", fmt.Errorf("unsupported snapshot kind: %s", kind)
		}
		requested[kind] = struct{}{}
	}
	if len(requested) == 0 {
		return "", "", "", nil, "", errors.New("allowed_kinds is required")
	}
	normalized := make([]string, 0, len(requested))
	for _, kind := range authorityKindOrder {
		if _, ok := requested[kind]; ok {
			normalized = append(normalized, kind)
		}
	}
	raw, err := json.Marshal(normalized)
	if err != nil {
		return "", "", "", nil, "", err
	}
	return authority, providerCode, producerType, normalized, string(raw), nil
}

func validAuthorityToken(value string, upper bool) bool {
	if len(value) == 0 || len(value) > 32 {
		return false
	}
	for index := 0; index < len(value); index++ {
		character := value[index]
		isLetter := character >= 'a' && character <= 'z'
		if upper {
			isLetter = character >= 'A' && character <= 'Z'
		}
		if !isLetter &&
			(character < '0' || character > '9') &&
			character != '-' &&
			character != '_' {
			return false
		}
	}
	first := value[0]
	last := value[len(value)-1]
	return first != '-' && first != '_' && last != '-' && last != '_'
}

func isAuthorityKind(kind string) bool {
	for _, allowed := range authorityKindOrder {
		if kind == allowed {
			return true
		}
	}
	return false
}

func authorityKinds(raw string) ([]string, error) {
	var kinds []string
	if err := json.Unmarshal([]byte(raw), &kinds); err != nil {
		return nil, fmt.Errorf("invalid authority allowed_kinds: %w", err)
	}
	return kinds, nil
}

func removesAuthorityKind(oldRaw string, next []string) (bool, error) {
	oldKinds, err := authorityKinds(oldRaw)
	if err != nil {
		return false, err
	}
	nextSet := make(map[string]struct{}, len(next))
	for _, kind := range next {
		nextSet[kind] = struct{}{}
	}
	for _, kind := range oldKinds {
		if _, ok := nextSet[strings.ToUpper(strings.TrimSpace(kind))]; !ok {
			return true, nil
		}
	}
	return false, nil
}

func authorityRegistryProto(row *models.TItickAuthorityRegistry) (*market.AuthorityRegistryData, error) {
	if row == nil {
		return nil, errors.New("authority registry row is nil")
	}
	kinds, err := authorityKinds(row.AllowedKinds)
	if err != nil {
		return nil, err
	}
	return &market.AuthorityRegistryData{
		Id:           row.Id,
		Authority:    row.Authority,
		ProviderCode: row.ProviderCode,
		ProducerType: row.ProducerType,
		AllowedKinds: kinds,
		Status:       common.Enable(row.Status),
		Version:      row.Version,
		CreateTimes:  row.CreateTimes,
		UpdateTimes:  row.UpdateTimes,
	}, nil
}
