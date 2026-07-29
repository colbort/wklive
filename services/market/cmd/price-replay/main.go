package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"wklive/services/market/internal/priceengine"
)

func main() {
	intervalMs := flag.Int64(
		"interval-ms", 0,
		"require each formula/version target sequence to be contiguous at this interval",
	)
	jsonOutput := flag.Bool("json", false, "print a machine-readable replay report")
	flag.Usage = func() {
		fmt.Fprintf(
			flag.CommandLine.Output(),
			"usage: price-replay [--interval-ms N] [--json] <audit.json|audit-array.json|audit.jsonl> [...]\n",
		)
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(2)
	}

	var records [][]byte
	for _, path := range flag.Args() {
		raw, err := os.ReadFile(path)
		if err != nil {
			log.Fatalf("read %s: %v", path, err)
		}
		fileRecords, err := decodeAuditRecords(raw)
		if err != nil {
			log.Fatalf("decode %s: %v", path, err)
		}
		records = append(records, fileRecords...)
	}
	report, err := priceengine.ReplayEvaluationAuditWindow(records, *intervalMs)
	if err != nil {
		log.Fatalf("replay failed: %v", err)
	}
	if *jsonOutput {
		raw, marshalErr := json.MarshalIndent(report, "", "  ")
		if marshalErr != nil {
			log.Fatalf("encode replay report: %v", marshalErr)
		}
		fmt.Println(string(raw))
		return
	}
	fmt.Printf(
		"replay verified: records=%d formulas=%d target_time=%d..%d interval_ms=%d\n",
		report.RecordCount, report.FormulaCount,
		report.FirstTargetTime, report.LastTargetTime, report.ExpectedIntervalMs,
	)
	for _, formula := range report.Formulas {
		fmt.Printf(
			"formula=%s version=%s records=%d target_time=%d..%d price=%s..%s min_inputs=%d rejected=%d\n",
			formula.FormulaNo, formula.FormulaVersion, formula.RecordCount,
			formula.FirstTargetTime, formula.LastTargetTime,
			formula.MinimumOutputPrice, formula.MaximumOutputPrice,
			formula.MinimumInputCount, formula.RejectedInputCount,
		)
	}
}

func decodeAuditRecords(raw []byte) ([][]byte, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return nil, fmt.Errorf("audit file is empty")
	}
	if raw[0] == '[' {
		var records []json.RawMessage
		if err := json.Unmarshal(raw, &records); err != nil {
			return nil, err
		}
		if len(records) == 0 {
			return nil, fmt.Errorf("audit array is empty")
		}
		result := make([][]byte, len(records))
		for index := range records {
			result[index] = append([]byte(nil), records[index]...)
		}
		return result, nil
	}

	decoder := json.NewDecoder(bytes.NewReader(raw))
	var records [][]byte
	for {
		var record json.RawMessage
		if err := decoder.Decode(&record); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		records = append(records, append([]byte(nil), record...))
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("audit file contains no records")
	}
	return records, nil
}
