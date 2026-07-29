package priceengine

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/shopspring/decimal"
)

type ReplayWindowReport struct {
	RecordCount        int                   `json:"record_count"`
	FormulaCount       int                   `json:"formula_count"`
	FirstTargetTime    int64                 `json:"first_target_time"`
	LastTargetTime     int64                 `json:"last_target_time"`
	ExpectedIntervalMs int64                 `json:"expected_interval_ms,omitempty"`
	Formulas           []ReplayFormulaReport `json:"formulas"`
}

type ReplayFormulaReport struct {
	FormulaNo          string `json:"formula_no"`
	FormulaVersion     string `json:"formula_version"`
	RecordCount        int    `json:"record_count"`
	FirstTargetTime    int64  `json:"first_target_time"`
	LastTargetTime     int64  `json:"last_target_time"`
	MinimumOutputPrice string `json:"minimum_output_price"`
	MaximumOutputPrice string `json:"maximum_output_price"`
	MinimumInputCount  int    `json:"minimum_accepted_input_count"`
	RejectedInputCount int    `json:"rejected_input_count"`
}

type replayWindowRecord struct {
	audit EvaluationAudit
	price decimal.Decimal
}

// ReplayEvaluationAuditWindow verifies an exported historical window solely
// from immutable evaluation audits. A positive interval additionally proves
// that every formula/version has a contiguous, aligned target-time sequence.
func ReplayEvaluationAuditWindow(
	records [][]byte, expectedIntervalMs int64,
) (*ReplayWindowReport, error) {
	if len(records) == 0 {
		return nil, errors.New("evaluation audit window is empty")
	}
	if expectedIntervalMs < 0 {
		return nil, errors.New("expected interval must not be negative")
	}

	groups := make(map[string][]replayWindowRecord)
	report := &ReplayWindowReport{
		RecordCount:        len(records),
		ExpectedIntervalMs: expectedIntervalMs,
	}
	for index, raw := range records {
		var audit EvaluationAudit
		if err := json.Unmarshal(raw, &audit); err != nil {
			return nil, fmt.Errorf("record %d decode evaluation audit: %w", index+1, err)
		}
		price, err := ReplayEvaluationAudit(raw)
		if err != nil {
			return nil, fmt.Errorf(
				"record %d formula=%s target=%d: %w",
				index+1, audit.FormulaNo, audit.TargetTime, err,
			)
		}
		key := strings.TrimSpace(audit.FormulaNo) + "\x00" +
			strings.TrimSpace(audit.FormulaVersion)
		groups[key] = append(groups[key], replayWindowRecord{audit: audit, price: price})
		if report.FirstTargetTime == 0 || audit.TargetTime < report.FirstTargetTime {
			report.FirstTargetTime = audit.TargetTime
		}
		if audit.TargetTime > report.LastTargetTime {
			report.LastTargetTime = audit.TargetTime
		}
	}

	report.FormulaCount = len(groups)
	keys := make([]string, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		group := groups[key]
		sort.Slice(group, func(i, j int) bool {
			return group[i].audit.TargetTime < group[j].audit.TargetTime
		})
		if err := validateReplayTargetSequence(group, expectedIntervalMs); err != nil {
			return nil, err
		}
		first := group[0]
		formula := ReplayFormulaReport{
			FormulaNo:          first.audit.FormulaNo,
			FormulaVersion:     first.audit.FormulaVersion,
			RecordCount:        len(group),
			FirstTargetTime:    first.audit.TargetTime,
			LastTargetTime:     group[len(group)-1].audit.TargetTime,
			MinimumOutputPrice: first.price.String(),
			MaximumOutputPrice: first.price.String(),
			MinimumInputCount:  len(first.audit.AcceptedInputs),
		}
		minimumPrice, maximumPrice := first.price, first.price
		for _, record := range group {
			if record.price.LessThan(minimumPrice) {
				minimumPrice = record.price
			}
			if record.price.GreaterThan(maximumPrice) {
				maximumPrice = record.price
			}
			if len(record.audit.AcceptedInputs) < formula.MinimumInputCount {
				formula.MinimumInputCount = len(record.audit.AcceptedInputs)
			}
			formula.RejectedInputCount += len(record.audit.RejectedInputs)
		}
		formula.MinimumOutputPrice = minimumPrice.String()
		formula.MaximumOutputPrice = maximumPrice.String()
		report.Formulas = append(report.Formulas, formula)
	}
	return report, nil
}

func validateReplayTargetSequence(
	records []replayWindowRecord, expectedIntervalMs int64,
) error {
	for index, record := range records {
		if index > 0 && record.audit.TargetTime == records[index-1].audit.TargetTime {
			return fmt.Errorf(
				"formula=%s version=%s has duplicate target_time=%d",
				record.audit.FormulaNo, record.audit.FormulaVersion, record.audit.TargetTime,
			)
		}
		if expectedIntervalMs == 0 {
			continue
		}
		if record.audit.TargetTime%expectedIntervalMs != 0 {
			return fmt.Errorf(
				"formula=%s version=%s target_time=%d is not aligned to interval=%d",
				record.audit.FormulaNo, record.audit.FormulaVersion,
				record.audit.TargetTime, expectedIntervalMs,
			)
		}
		if index > 0 &&
			record.audit.TargetTime-records[index-1].audit.TargetTime != expectedIntervalMs {
			return fmt.Errorf(
				"formula=%s version=%s target sequence gap: previous=%d current=%d expected_interval=%d",
				record.audit.FormulaNo, record.audit.FormulaVersion,
				records[index-1].audit.TargetTime, record.audit.TargetTime,
				expectedIntervalMs,
			)
		}
	}
	return nil
}
