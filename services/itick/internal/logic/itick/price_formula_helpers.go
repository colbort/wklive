package iticklogic

import (
	"encoding/json"
	"errors"
	"strings"

	"wklive/proto/itick"
	"wklive/services/itick/internal/priceengine"
	"wklive/services/itick/models"

	"github.com/shopspring/decimal"
)

func normalizePriceFormulaReq(in *itick.CreatePriceFormulaReq) ([]priceengine.Component, error) {
	if in == nil || strings.TrimSpace(in.FormulaNo) == "" || strings.TrimSpace(in.FormulaVersion) == "" || strings.TrimSpace(in.Symbol) == "" {
		return nil, errors.New("formula_no, formula_version and symbol are required")
	}
	in.Authority = strings.ToLower(strings.TrimSpace(in.Authority))
	in.SnapshotKind = strings.ToUpper(strings.TrimSpace(in.SnapshotKind))
	in.CategoryCode = strings.ToLower(strings.TrimSpace(in.CategoryCode))
	in.Market = strings.ToUpper(strings.TrimSpace(in.Market))
	in.Symbol = strings.ToUpper(strings.TrimSpace(in.Symbol))
	if in.MaxLookbackMs <= 0 || in.IntervalMs <= 0 || in.MaxDeviationBps < 0 || in.MaxDeviationBps > 10000 {
		return nil, errors.New("invalid formula timing or deviation configuration")
	}
	if in.Algorithm != itick.PriceAlgorithm_PRICE_ALGORITHM_WEIGHTED_MEAN &&
		in.Algorithm != itick.PriceAlgorithm_PRICE_ALGORITHM_MEDIAN &&
		in.Algorithm != itick.PriceAlgorithm_PRICE_ALGORITHM_PREMIUM_RATE &&
		in.Algorithm != itick.PriceAlgorithm_PRICE_ALGORITHM_INDEX_BASIS {
		return nil, errors.New("unsupported price formula algorithm")
	}
	if len(in.Components) == 0 ||
		(in.Algorithm == itick.PriceAlgorithm_PRICE_ALGORITHM_PREMIUM_RATE && len(in.Components) != 2) ||
		(in.Algorithm == itick.PriceAlgorithm_PRICE_ALGORITHM_INDEX_BASIS && len(in.Components) != 2 && len(in.Components) != 3) {
		return nil, errors.New("invalid price formula components")
	}
	if in.MinInputCount == 0 {
		in.MinInputCount = int64(len(in.Components))
	}
	if in.MinInputCount < 1 || in.MinInputCount > int64(len(in.Components)) ||
		in.Algorithm == itick.PriceAlgorithm_PRICE_ALGORITHM_PREMIUM_RATE && in.MinInputCount != 2 ||
		in.Algorithm == itick.PriceAlgorithm_PRICE_ALGORITHM_INDEX_BASIS && in.MinInputCount != int64(len(in.Components)) {
		return nil, errors.New("min_input_count must be within components and formula-required inputs must all be present")
	}
	if in.SnapshotKind == "DELIVERY" && in.MinInputCount < 3 {
		return nil, errors.New("DELIVERY requires at least 3 accepted inputs")
	}
	components := make([]priceengine.Component, 0, len(in.Components))
	for _, item := range in.Components {
		if item == nil {
			return nil, errors.New("nil price formula component")
		}
		component := priceengine.Component{Authority: strings.ToLower(strings.TrimSpace(item.Authority)), Kind: strings.ToUpper(strings.TrimSpace(item.SnapshotKind)), CategoryCode: strings.ToLower(strings.TrimSpace(item.CategoryCode)), Market: strings.ToUpper(strings.TrimSpace(item.Market)), Symbol: strings.ToUpper(strings.TrimSpace(item.Symbol)), Weight: strings.TrimSpace(item.Weight)}
		if component.Authority == "" || component.Symbol == "" {
			return nil, errors.New("component authority and symbol are required")
		}
		if component.Kind == "" {
			component.Kind = "FINAL_QUOTE"
		}
		weight, err := decimal.NewFromString(component.Weight)
		if err != nil || !weight.IsPositive() {
			return nil, errors.New("component weight must be positive decimal")
		}
		components = append(components, component)
	}
	if in.SnapshotKind == "INDEX" {
		if len(components) < 3 || in.MinInputCount < 3 ||
			(in.Algorithm != itick.PriceAlgorithm_PRICE_ALGORITHM_MEDIAN &&
				in.Algorithm != itick.PriceAlgorithm_PRICE_ALGORITHM_WEIGHTED_MEAN) {
			return nil, errors.New("INDEX requires MEDIAN or WEIGHTED_MEAN with at least 3 accepted sources")
		}
		for _, component := range components {
			if component.Kind != "FINAL_QUOTE" || component.Authority == in.Authority {
				return nil, errors.New("INDEX components must be independent FINAL_QUOTE authorities")
			}
		}
		if distinctPriceSourceCount(components) != len(components) {
			return nil, errors.New("INDEX components must identify distinct sources")
		}
	}
	if in.SnapshotKind == "DELIVERY" && distinctPriceSourceCount(components) != len(components) {
		return nil, errors.New("DELIVERY components must identify distinct sources")
	}
	if in.Algorithm == itick.PriceAlgorithm_PRICE_ALGORITHM_INDEX_BASIS {
		if in.SnapshotKind != "MARK" || in.MaxDeviationBps <= 0 ||
			components[0].Kind != "INDEX" || components[0].Authority != in.Authority ||
			components[1].Kind != "FINAL_QUOTE" || components[1].Authority == in.Authority ||
			components[0].Symbol != in.Symbol || components[1].Symbol != in.Symbol ||
			len(components) == 3 && (components[2].Kind != "MARK" ||
				components[2].Authority != in.Authority || components[2].Symbol != in.Symbol) {
			return nil, errors.New("INDEX_BASIS requires MARK output, positive basis cap, INDEX, independent FINAL_QUOTE and optional previous MARK for the same symbol")
		}
	}
	return components, nil
}

func distinctPriceSourceCount(components []priceengine.Component) int {
	sources := make(map[string]struct{}, len(components))
	for _, component := range components {
		key := strings.Join([]string{
			component.Authority, component.Kind, component.CategoryCode,
			component.Market, component.Symbol,
		}, "\x00")
		sources[key] = struct{}{}
	}
	return len(sources)
}

func toPriceFormulaProto(row *models.TItickPriceFormula) *itick.PriceFormulaData {
	if row == nil {
		return nil
	}
	var components []priceengine.Component
	_ = json.Unmarshal([]byte(row.Components), &components)
	result := &itick.PriceFormulaData{Id: row.Id, FormulaNo: row.FormulaNo, Authority: row.Authority, SnapshotKind: row.SnapshotKind, CategoryCode: row.CategoryCode, Market: row.Market, Symbol: row.Symbol, Algorithm: itick.PriceAlgorithm(row.Algorithm), FormulaVersion: row.FormulaVersion, MaxLookbackMs: row.MaxLookbackMs, MaxDeviationBps: row.MaxDeviationBps, MinInputCount: row.MinInputCount, IntervalMs: row.IntervalMs, LastTargetTime: row.LastTargetTime, Status: int32(row.Status), Version: row.Version, CreateTimes: row.CreateTimes, UpdateTimes: row.UpdateTimes}
	for _, component := range components {
		result.Components = append(result.Components, &itick.PriceFormulaComponent{Authority: component.Authority, SnapshotKind: component.Kind, CategoryCode: component.CategoryCode, Market: component.Market, Symbol: component.Symbol, Weight: component.Weight})
	}
	return result
}
