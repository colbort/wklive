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
	FormulaNo       string  `json:"formula_no"`
	FormulaVersion  string  `json:"formula_version"`
	Algorithm       int64   `json:"algorithm"`
	TargetTime      int64   `json:"target_time"`
	AllInputs       []Input `json:"all_inputs"`
	AcceptedInputs  []Input `json:"accepted_inputs"`
	RejectedInputs  []Input `json:"rejected_inputs"`
	MaxDeviationBps int64   `json:"max_deviation_bps"`
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
	inputs := make([]Input, 0, len(components))
	sourceTime := target
	for _, c := range components {
		kind := strings.ToUpper(strings.TrimSpace(c.Kind))
		if kind == "" {
			kind = "FINAL_QUOTE"
		}
		row, err := e.archive.FindAtOrBefore(ctx, c.Authority, kind, c.CategoryCode, c.Market, c.Symbol, target, target-f.MaxLookbackMs)
		if err != nil {
			return fmt.Errorf("%w: formula=%s authority=%s kind=%s category=%s market=%s symbol=%s target=%d: %v",
				ErrInputUnavailable, f.FormulaNo, c.Authority, kind, c.CategoryCode, c.Market, c.Symbol, target, err)
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
	inputs, rejectedInputs := filterDeviationWithAudit(inputs, f.MaxDeviationBps)
	price, err := Calculate(itick.PriceAlgorithm(f.Algorithm), inputs)
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
	}
	raw, err := json.Marshal(audit)
	if err != nil {
		return err
	}
	id := deterministicSnapshotID(f, price, raw)
	s := &market.SettlementSnapshot{SnapshotID: id, Kind: f.SnapshotKind, CategoryCode: f.CategoryCode, Market: f.Market, Symbol: f.Symbol, Price: price.String(), Source: "price-engine", SourceTimestamp: sourceTime, SnapshotTimestamp: time.Now().UnixMilli(), Revision: target, FormulaVersion: f.FormulaVersion, Authority: f.Authority, Confirmed: true}
	payload, _ := json.Marshal(map[string]any{"snapshot": s})
	return e.archive.InsertImmutableAndEnqueue(ctx, &models.TItickAuthoritativeSnapshot{SnapshotId: id, Authority: f.Authority, SnapshotKind: f.SnapshotKind, CategoryCode: f.CategoryCode, Market: f.Market, Symbol: f.Symbol, Price: price, SourceTimestamp: sourceTime, SnapshotTimestamp: s.SnapshotTimestamp, Revision: target, FormulaVersion: f.FormulaVersion, RawPayload: string(raw), CreateTimes: s.SnapshotTimestamp}, string(payload))
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
