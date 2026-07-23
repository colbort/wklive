package tradetasklogic

import (
	"errors"
	"strings"
	"testing"

	"wklive/proto/asset"
	"wklive/proto/common"
	"wklive/proto/trade"
	"wklive/services/trade/models"

	"github.com/shopspring/decimal"
)

func TestValidateSecondsAssetResponse(t *testing.T) {
	var nilResponse *asset.ChangeAssetResp
	if err := validateSecondsAssetResponse("payout", nilResponse); err == nil {
		t.Fatal("typed nil Asset response must be rejected")
	}
	if err := validateSecondsAssetResponse("payout", &asset.ChangeAssetResp{}); err == nil {
		t.Fatal("Asset response without base must be rejected")
	}
	if err := validateSecondsAssetResponse("payout", &asset.ChangeAssetResp{Base: &common.RespBase{Code: 500}}); err == nil {
		t.Fatal("rejected Asset response must return an error")
	}
	if err := validateSecondsAssetResponse("payout", &asset.ChangeAssetResp{Base: &common.RespBase{Code: 200}}); err != nil {
		t.Fatalf("successful Asset response rejected: %v", err)
	}
}

func TestScanSecondsWorkContinuesAfterItemFailure(t *testing.T) {
	fetched := false
	processed := make([]int64, 0, 3)
	err := scanSecondsWork(func(cursor int64) ([]*models.SecondsOrderWorkItem, error) {
		if fetched {
			return nil, nil
		}
		fetched = true
		return []*models.SecondsOrderWorkItem{{TTradeOrderSeconds: models.TTradeOrderSeconds{Id: 1}}, {TTradeOrderSeconds: models.TTradeOrderSeconds{Id: 2}}, {TTradeOrderSeconds: models.TTradeOrderSeconds{Id: 3}}}, nil
	}, func(item *models.SecondsOrderWorkItem) error {
		processed = append(processed, item.Id)
		if item.Id == 1 {
			return errors.New("injected")
		}
		return nil
	})
	if err == nil {
		t.Fatal("expected aggregate error")
	}
	if len(processed) != 3 {
		t.Fatalf("processed=%v, want all work items", processed)
	}
}

func TestRunSecondsPhasesDoesNotShortCircuit(t *testing.T) {
	called := 0
	err := runSecondsPhases(
		func() error { called++; return errors.New("activation") },
		func() error { called++; return errors.New("settlement") },
		func() error { called++; return nil },
	)
	if called != 3 {
		t.Fatalf("called=%d, want all phases", called)
	}
	if err == nil || !strings.Contains(err.Error(), "activation") || !strings.Contains(err.Error(), "settlement") {
		t.Fatalf("aggregate error=%v", err)
	}
}

func TestSecondsWorkLeaseFencing(t *testing.T) {
	current := &models.TTradeOrderSeconds{SettlementStatus: int64(trade.SecondsSettlementStatus_SECONDS_SETTLEMENT_STATUS_SETTLING), UpdateTimes: 100}
	if !secondsWorkLeaseOwned(current, trade.SecondsSettlementStatus_SECONDS_SETTLEMENT_STATUS_SETTLING, 100) {
		t.Fatal("current seconds settlement lease should be owned")
	}
	if secondsWorkLeaseOwned(current, trade.SecondsSettlementStatus_SECONDS_SETTLEMENT_STATUS_SETTLING, 99) {
		t.Fatal("stale seconds settlement lease must be fenced")
	}
	current.SettlementStatus = int64(trade.SecondsSettlementStatus_SECONDS_SETTLEMENT_STATUS_SETTLED)
	if secondsWorkLeaseOwned(current, trade.SecondsSettlementStatus_SECONDS_SETTLEMENT_STATUS_SETTLING, 100) {
		t.Fatal("completed seconds order must not retain settlement lease")
	}
}

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
