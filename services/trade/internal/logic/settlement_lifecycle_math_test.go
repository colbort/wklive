package logic

import (
	"testing"

	"wklive/proto/trade"

	"github.com/shopspring/decimal"
)

func TestSecondsResultAndPayout(t *testing.T) {
	start := decimal.NewFromInt(100)
	if got := secondsResult(int64(trade.SecondsDirection_SECONDS_DIRECTION_UP), start, decimal.RequireFromString("100.01"), decimal.RequireFromString("0.001")); got != trade.SecondsResult_SECONDS_RESULT_WIN {
		t.Fatalf("up result=%v", got)
	}
	if got := secondsResult(int64(trade.SecondsDirection_SECONDS_DIRECTION_DOWN), start, decimal.RequireFromString("99.99"), decimal.RequireFromString("0.001")); got != trade.SecondsResult_SECONDS_RESULT_WIN {
		t.Fatalf("down result=%v", got)
	}
	if got := secondsResult(int64(trade.SecondsDirection_SECONDS_DIRECTION_UP), start, decimal.RequireFromString("100.0005"), decimal.RequireFromString("0.001")); got != trade.SecondsResult_SECONDS_RESULT_DRAW {
		t.Fatalf("draw result=%v", got)
	}
	profit, fee, returned := secondsPayout(decimal.NewFromInt(10), decimal.RequireFromString("0.8"), decimal.RequireFromString("0.1"), trade.SecondsResult_SECONDS_RESULT_WIN)
	if !profit.Equal(decimal.NewFromInt(8)) || !fee.Equal(decimal.RequireFromString("0.8")) || !returned.Equal(decimal.RequireFromString("17.2")) {
		t.Fatalf("unexpected payout profit=%s fee=%s return=%s", profit, fee, returned)
	}
}

func TestQuoteSourceParsing(t *testing.T) {
	c, m, s := parseQuoteSource("crypto:BA:BTCUSDT")
	if c != "crypto" || m != "BA" || s != "BTCUSDT" {
		t.Fatalf("unexpected source %q %q %q", c, m, s)
	}
}
