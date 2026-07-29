package priceengine

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	market "wklive/common/market"
	"wklive/proto/itick"
	"wklive/services/itick/models"

	"github.com/shopspring/decimal"
)

type Component struct {
	Authority    string `json:"authority"`
	Kind         string `json:"kind"`
	CategoryCode string `json:"category_code"`
	Market       string `json:"market"`
	Symbol       string `json:"symbol"`
	Weight       string `json:"weight"`
}

type EvaluationAudit struct {
	FormulaNo        string  `json:"formula_no"`
	FormulaVersion   string  `json:"formula_version"`
	Algorithm        int64   `json:"algorithm"`
	TargetTime       int64   `json:"target_time"`
	AllInputs        []Input `json:"all_inputs"`
	AcceptedInputs   []Input `json:"accepted_inputs"`
	RejectedInputs   []Input `json:"rejected_inputs"`
	MaxDeviationBps  int64   `json:"max_deviation_bps"`
	MinInputCount    int64   `json:"min_input_count"`
	RawBasisRate     string  `json:"raw_basis_rate,omitempty"`
	AppliedBasisRate string  `json:"applied_basis_rate,omitempty"`
	UnsmoothedMark   string  `json:"unsmoothed_mark,omitempty"`
	OutputPrice      string  `json:"output_price"`
}

var ErrInputUnavailable = errors.New("price engine input unavailable")

type Archive interface {
	FindAtOrBefore(context.Context, string, string, string, string, string, int64, int64) (*models.TItickAuthoritativeSnapshot, error)
	InsertImmutableAndEnqueue(context.Context, *models.TItickAuthoritativeSnapshot, string) error
}
type Engine struct {
	formulas models.TItickPriceFormulaModel
	archive  Archive
}

func New(formulas models.TItickPriceFormulaModel, archive Archive) *Engine {
	return &Engine{formulas: formulas, archive: archive}
}

func (e *Engine) RunOnce(ctx context.Context, now int64) error {
	rows, err := e.formulas.FindDue(ctx, now, 100)
	if err != nil {
		return err
	}
	var firstErr error
	for _, f := range rows {
		if f.IntervalMs <= 0 || f.MaxLookbackMs < 0 {
			if firstErr == nil {
				firstErr = errors.New("invalid price formula timing configuration")
			}
			continue
		}
		target := now / f.IntervalMs * f.IntervalMs
		claimed, claimErr := e.formulas.ClaimTarget(ctx, f.Id, f.RunVersion, target, now)
		if claimErr != nil {
			return claimErr
		}
		if !claimed {
			continue
		}
		if err = e.evaluate(ctx, f, target); err != nil {
			_ = e.formulas.ReleaseTarget(ctx, f.Id, f.RunVersion+1, target, f.LastTargetTime)
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

func (e *Engine) evaluate(ctx context.Context, f *models.TItickPriceFormula, target int64) error {
	var components []Component
	if err := json.Unmarshal([]byte(f.Components), &components); err != nil {
		return err
	}
	if len(components) == 0 {
		return errors.New("price formula components are empty")
	}
	if itick.PriceAlgorithm(f.Algorithm) == itick.PriceAlgorithm_PRICE_ALGORITHM_PREMIUM_RATE && len(components) != 2 {
		return errors.New("PREMIUM_RATE requires exactly mark and index components")
	}
	if itick.PriceAlgorithm(f.Algorithm) == itick.PriceAlgorithm_PRICE_ALGORITHM_INDEX_BASIS {
		if !strings.EqualFold(f.SnapshotKind, "MARK") || (len(components) != 2 && len(components) != 3) ||
			!strings.EqualFold(components[0].Kind, "INDEX") ||
			!strings.EqualFold(components[0].Authority, f.Authority) ||
			!strings.EqualFold(components[0].Symbol, f.Symbol) ||
			!strings.EqualFold(components[1].Kind, "FINAL_QUOTE") ||
			strings.EqualFold(components[1].Authority, f.Authority) ||
			!strings.EqualFold(components[1].Symbol, f.Symbol) ||
			len(components) == 3 && (!strings.EqualFold(components[2].Kind, "MARK") ||
				!strings.EqualFold(components[2].Authority, f.Authority) ||
				!strings.EqualFold(components[2].Symbol, f.Symbol)) {
			return errors.New("INDEX_BASIS requires MARK output and ordered INDEX, FINAL_QUOTE, optional previous MARK components")
		}
	}
	inputs := make([]Input, 0, len(components))
	sourceTime := target
	for componentIndex, c := range components {
		kind := strings.ToUpper(strings.TrimSpace(c.Kind))
		if kind == "" {
			kind = "FINAL_QUOTE"
		}
		lookupTarget := target
		if itick.PriceAlgorithm(f.Algorithm) == itick.PriceAlgorithm_PRICE_ALGORITHM_INDEX_BASIS && componentIndex == 2 {
			// The smoothing input is the previous MARK, never the output being
			// evaluated at this same target.
			lookupTarget = target - 1
		}
		row, err := e.archive.FindAtOrBefore(ctx, c.Authority, kind, c.CategoryCode, c.Market, c.Symbol, lookupTarget, target-f.MaxLookbackMs)
		if err != nil {
			return fmt.Errorf("%w: formula=%s authority=%s kind=%s category=%s market=%s symbol=%s target=%d: %v",
				ErrInputUnavailable, f.FormulaNo, c.Authority, kind, c.CategoryCode, c.Market, c.Symbol, target, err)
		}
		if strings.TrimSpace(row.SnapshotId) == "" {
			return fmt.Errorf("%w: formula=%s authority=%s kind=%s returned empty snapshot id",
				ErrInputUnavailable, f.FormulaNo, c.Authority, kind)
		}
		weight, err := decimal.NewFromString(c.Weight)
		if err != nil {
			return err
		}
		inputs = append(inputs, Input{Price: row.Price, Weight: weight, SnapshotID: row.SnapshotId})
		if row.SourceTimestamp < sourceTime {
			sourceTime = row.SourceTimestamp
		}
	}
	allInputs := append([]Input(nil), inputs...)
	inputs, rejectedInputs := deduplicateInputsWithAudit(inputs)
	if itick.PriceAlgorithm(f.Algorithm) != itick.PriceAlgorithm_PRICE_ALGORITHM_INDEX_BASIS {
		var deviationRejected []Input
		inputs, deviationRejected = filterDeviationWithAudit(inputs, f.MaxDeviationBps)
		rejectedInputs = append(rejectedInputs, deviationRejected...)
	}
	minInputCount := effectiveMinInputCount(f)
	if int64(len(inputs)) < minInputCount {
		return fmt.Errorf("%w: formula=%s accepted=%d required=%d rejected=%d target=%d",
			ErrInputUnavailable, f.FormulaNo, len(inputs), minInputCount, len(rejectedInputs), target)
	}
	var price decimal.Decimal
	var rawBasis, appliedBasis decimal.Decimal
	var err error
	if itick.PriceAlgorithm(f.Algorithm) == itick.PriceAlgorithm_PRICE_ALGORITHM_INDEX_BASIS {
		price, rawBasis, appliedBasis, err = CalculateIndexBasis(inputs, f.MaxDeviationBps)
	} else {
		price, err = Calculate(itick.PriceAlgorithm(f.Algorithm), inputs)
	}
	if err != nil {
		return err
	}
	audit := EvaluationAudit{
		FormulaNo:       f.FormulaNo,
		FormulaVersion:  f.FormulaVersion,
		Algorithm:       f.Algorithm,
		TargetTime:      target,
		AllInputs:       allInputs,
		AcceptedInputs:  inputs,
		RejectedInputs:  rejectedInputs,
		MaxDeviationBps: f.MaxDeviationBps,
		MinInputCount:   minInputCount,
	}
	if itick.PriceAlgorithm(f.Algorithm) == itick.PriceAlgorithm_PRICE_ALGORITHM_INDEX_BASIS {
		audit.RawBasisRate = rawBasis.String()
		audit.AppliedBasisRate = appliedBasis.String()
		audit.UnsmoothedMark = inputs[0].Price.Mul(decimal.NewFromInt(1).Add(appliedBasis)).String()
	}
	audit.OutputPrice = price.String()
	raw, err := json.Marshal(audit)
	if err != nil {
		return err
	}
	id := deterministicSnapshotID(f, price, raw)
	s := &market.SettlementSnapshot{SnapshotID: id, Kind: f.SnapshotKind, CategoryCode: f.CategoryCode, Market: f.Market, Symbol: f.Symbol, Price: price.String(), Source: "price-engine", SourceTimestamp: sourceTime, SnapshotTimestamp: time.Now().UnixMilli(), Revision: target, FormulaVersion: f.FormulaVersion, Authority: f.Authority, Confirmed: true}
	payload, _ := json.Marshal(map[string]any{"snapshot": s})
	return e.archive.InsertImmutableAndEnqueue(ctx, &models.TItickAuthoritativeSnapshot{SnapshotId: id, Authority: f.Authority, SnapshotKind: f.SnapshotKind, CategoryCode: f.CategoryCode, Market: f.Market, Symbol: f.Symbol, Price: price, SourceTimestamp: sourceTime, SnapshotTimestamp: s.SnapshotTimestamp, Revision: target, FormulaVersion: f.FormulaVersion, RawPayload: string(raw), CreateTimes: s.SnapshotTimestamp}, string(payload))
}

// ReplayEvaluationAudit recalculates a published formula solely from its
// immutable audit payload and rejects any output mismatch. It is intentionally
// independent of current formula rows and live market state.
func ReplayEvaluationAudit(raw []byte) (decimal.Decimal, error) {
	var audit EvaluationAudit
	if err := json.Unmarshal(raw, &audit); err != nil {
		return decimal.Zero, fmt.Errorf("decode evaluation audit: %w", err)
	}
	if strings.TrimSpace(audit.FormulaNo) == "" || strings.TrimSpace(audit.FormulaVersion) == "" ||
		audit.TargetTime <= 0 || len(audit.AllInputs) == 0 || len(audit.AcceptedInputs) == 0 ||
		audit.MinInputCount <= 0 || strings.TrimSpace(audit.OutputPrice) == "" {
		return decimal.Zero, errors.New("evaluation audit is incomplete")
	}
	if int64(len(audit.AcceptedInputs)) < audit.MinInputCount {
		return decimal.Zero, fmt.Errorf(
			"evaluation audit accepted inputs below minimum: accepted=%d required=%d",
			len(audit.AcceptedInputs), audit.MinInputCount,
		)
	}
	seen := make(map[string]struct{}, len(audit.AcceptedInputs))
	for _, input := range audit.AcceptedInputs {
		if strings.TrimSpace(input.SnapshotID) == "" {
			return decimal.Zero, errors.New("evaluation audit contains an input without snapshot id")
		}
		if _, exists := seen[input.SnapshotID]; exists {
			return decimal.Zero, errors.New("evaluation audit contains duplicate accepted snapshot ids")
		}
		seen[input.SnapshotID] = struct{}{}
	}
	algorithm := itick.PriceAlgorithm(audit.Algorithm)
	expectedAccepted, expectedRejected := deduplicateInputsWithAudit(
		append([]Input(nil), audit.AllInputs...),
	)
	if algorithm != itick.PriceAlgorithm_PRICE_ALGORITHM_INDEX_BASIS {
		var deviationRejected []Input
		expectedAccepted, deviationRejected = filterDeviationWithAudit(
			expectedAccepted, audit.MaxDeviationBps,
		)
		expectedRejected = append(expectedRejected, deviationRejected...)
	}
	if !equalAuditInputs(expectedAccepted, audit.AcceptedInputs) ||
		!equalAuditInputs(expectedRejected, audit.RejectedInputs) {
		return decimal.Zero, errors.New("evaluation audit accepted/rejected input partition mismatch")
	}
	var replayed decimal.Decimal
	var err error
	if algorithm == itick.PriceAlgorithm_PRICE_ALGORITHM_INDEX_BASIS {
		replayed, _, _, err = CalculateIndexBasis(audit.AcceptedInputs, audit.MaxDeviationBps)
	} else {
		replayed, err = Calculate(algorithm, audit.AcceptedInputs)
	}
	if err != nil {
		return decimal.Zero, fmt.Errorf("recalculate evaluation audit: %w", err)
	}
	expected, err := decimal.NewFromString(audit.OutputPrice)
	if err != nil {
		return decimal.Zero, errors.New("evaluation audit output price is invalid")
	}
	if !replayed.Equal(expected) {
		return decimal.Zero, fmt.Errorf("evaluation audit output mismatch: replayed=%s recorded=%s", replayed, expected)
	}
	return replayed, nil
}

func equalAuditInputs(left, right []Input) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index].SnapshotID != right[index].SnapshotID ||
			!left[index].Price.Equal(right[index].Price) ||
			!left[index].Weight.Equal(right[index].Weight) {
			return false
		}
	}
	return true
}

func deduplicateInputsWithAudit(inputs []Input) ([]Input, []Input) {
	accepted := make([]Input, 0, len(inputs))
	rejected := make([]Input, 0)
	seen := make(map[string]struct{}, len(inputs))
	for _, input := range inputs {
		if _, exists := seen[input.SnapshotID]; exists {
			rejected = append(rejected, input)
			continue
		}
		seen[input.SnapshotID] = struct{}{}
		accepted = append(accepted, input)
	}
	return accepted, rejected
}

func effectiveMinInputCount(f *models.TItickPriceFormula) int64 {
	if f == nil {
		return 1
	}
	minimum := f.MinInputCount
	if minimum <= 0 {
		minimum = 1
	}
	// Delivery is an irreversible financial terminal price. Existing formulas
	// created before min_input_count was introduced are also held to three
	// accepted sources instead of silently retaining a one-source fallback.
	if strings.EqualFold(f.SnapshotKind, "DELIVERY") && minimum < 3 {
		minimum = 3
	}
	return minimum
}

func filterDeviation(in []Input, bps int64) []Input {
	accepted, _ := filterDeviationWithAudit(in, bps)
	return accepted
}

func filterDeviationWithAudit(in []Input, bps int64) ([]Input, []Input) {
	if bps <= 0 || len(in) < 3 {
		return in, nil
	}
	prices := make([]decimal.Decimal, len(in))
	for i := range in {
		prices[i] = in[i].Price
	}
	sort.Slice(prices, func(i, j int) bool { return prices[i].LessThan(prices[j]) })
	median := prices[len(prices)/2]
	limit := decimal.NewFromInt(bps).Div(decimal.NewFromInt(10000))
	out := in[:0]
	rejected := make([]Input, 0)
	for _, v := range in {
		if v.Price.Sub(median).Abs().Div(median).LessThanOrEqual(limit) {
			out = append(out, v)
		} else {
			rejected = append(rejected, v)
		}
	}
	return out, rejected
}

func deterministicSnapshotID(f *models.TItickPriceFormula, price decimal.Decimal, raw []byte) string {
	sum := sha256.Sum256(append([]byte(f.Authority+"|"+f.SnapshotKind+"|"+f.CategoryCode+"|"+f.Market+"|"+f.Symbol+"|"+price.String()+"|"), raw...))
	return hex.EncodeToString(sum[:])
}
