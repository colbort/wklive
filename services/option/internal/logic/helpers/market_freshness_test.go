package helpers

import (
	"testing"

	"wklive/services/option/models"

	"github.com/shopspring/decimal"
)

func TestMarketFreshnessSeparatesUnderlyingAndMark(t *testing.T) {
	market := &models.TOptionMarket{
		UnderlyingPrice:        decimal.NewFromInt(100),
		MarkPrice:              decimal.NewFromInt(5),
		UnderlyingSnapshotTime: 100,
		MarkSnapshotTime:       60,
		SnapshotTime:           100,
	}
	if !IsUnderlyingFresh(market, 110, 30) {
		t.Fatal("fresh underlying rejected")
	}
	if IsMarkFresh(market, 110, 30) {
		t.Fatal("stale mark accepted because compatibility timestamp is fresh")
	}
	if IsRiskMarketFresh(market, 110, 30) {
		t.Fatal("risk market accepted with stale mark")
	}

	market.MarkSnapshotTime = 105
	if !IsRiskMarketFresh(market, 110, 30) {
		t.Fatal("fresh underlying and mark rejected")
	}
}

func TestFutureMarketTimestampIsRejected(t *testing.T) {
	if IsFreshTimestamp(101, 100, 30) {
		t.Fatal("future timestamp accepted")
	}
}

func TestGreeksFreshnessRequiresApprovedPositiveThreshold(t *testing.T) {
	market := &models.TOptionMarket{GreeksSnapshotTime: 90}
	if IsGreeksFresh(market, 100, 0) {
		t.Fatal("unconfigured Greeks threshold must fail closed")
	}
	if !IsGreeksFresh(market, 100, 10) {
		t.Fatal("Greeks timestamp at the approved boundary should be fresh")
	}
	if IsGreeksFresh(market, 101, 10) {
		t.Fatal("Greeks timestamp older than the approved threshold was accepted")
	}
	market.GreeksSnapshotTime = 102
	if IsGreeksFresh(market, 101, 10) {
		t.Fatal("future Greeks timestamp was accepted")
	}
}
