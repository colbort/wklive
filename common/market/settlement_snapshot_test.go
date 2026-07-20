package cache

import "testing"

func TestSettlementSnapshotDigestBindsRevisionAndFormula(t *testing.T) {
	base := &SettlementSnapshot{Kind: "FUNDING", MarkPrice: "100", IndexPrice: "99", FundingRate: "0.01", SourceTimestamp: 1000, SnapshotTimestamp: 1001, Revision: 7, FormulaVersion: "premium-v1", Confirmed: true}
	a := snapshotDigest(base)
	copy := *base
	copy.Revision++
	if a == snapshotDigest(&copy) {
		t.Fatal("revision must change snapshot id")
	}
	copy = *base
	copy.FormulaVersion = "premium-v2"
	if a == snapshotDigest(&copy) {
		t.Fatal("formula version must change snapshot id")
	}
	copy = *base
	copy.SnapshotTimestamp++
	if a != snapshotDigest(&copy) {
		t.Fatal("read time must not change source snapshot id")
	}
}
