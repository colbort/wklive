package iticktasklogic

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
	if in.Algorithm != itick.PriceAlgorithm_PRICE_ALGORITHM_WEIGHTED_MEAN && in.Algorithm != itick.PriceAlgorithm_PRICE_ALGORITHM_MEDIAN && in.Algorithm != itick.PriceAlgorithm_PRICE_ALGORITHM_PREMIUM_RATE {
		return nil, errors.New("unsupported price formula algorithm")
	}
	if len(in.Components) == 0 || (in.Algorithm == itick.PriceAlgorithm_PRICE_ALGORITHM_PREMIUM_RATE && len(in.Components) != 2) {
		return nil, errors.New("invalid price formula components")
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
	return components, nil
}

func toPriceFormulaProto(row *models.TItickPriceFormula) *itick.PriceFormulaData {
	if row == nil {
		return nil
	}
	var components []priceengine.Component
	_ = json.Unmarshal([]byte(row.Components), &components)
	result := &itick.PriceFormulaData{Id: row.Id, FormulaNo: row.FormulaNo, Authority: row.Authority, SnapshotKind: row.SnapshotKind, CategoryCode: row.CategoryCode, Market: row.Market, Symbol: row.Symbol, Algorithm: itick.PriceAlgorithm(row.Algorithm), FormulaVersion: row.FormulaVersion, MaxLookbackMs: row.MaxLookbackMs, MaxDeviationBps: row.MaxDeviationBps, IntervalMs: row.IntervalMs, LastTargetTime: row.LastTargetTime, Status: int32(row.Status), Version: row.Version, CreateTimes: row.CreateTimes, UpdateTimes: row.UpdateTimes}
	for _, component := range components {
		result.Components = append(result.Components, &itick.PriceFormulaComponent{Authority: component.Authority, SnapshotKind: component.Kind, CategoryCode: component.CategoryCode, Market: component.Market, Symbol: component.Symbol, Weight: component.Weight})
	}
	return result
}
