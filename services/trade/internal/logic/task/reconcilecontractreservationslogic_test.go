package tasklogic

import (
	"testing"

	"wklive/proto/asset"
	"wklive/proto/trade"
)

func TestAssetFreezeMatchesReservation(t *testing.T) {
	row := &contractReservationAudit{
		TenantId:       1,
		UserId:         2,
		ReservationNo:  "ORDER-1",
		Asset:          "USDT",
		ReservedAmount: auditDecimal("100"),
		ConsumedAmount: auditDecimal("30"),
		ReleasedAmount: auditDecimal("20"),
		Status:         int64(trade.AssetReservationStatus_ASSET_RESERVATION_STATUS_PART_CONSUMED),
	}
	freeze := &asset.AssetFreeze{
		TenantId:       1,
		UserId:         2,
		Coin:           "USDT",
		BizType:        asset.BizType_BIZ_TYPE_TRADE,
		BizNo:          "ORDER-1",
		Amount:         "100",
		UsedAmount:     "30",
		UnfreezeAmount: "20",
		RemainAmount:   "50",
		Status:         asset.FreezeStatus_FREEZE_STATUS_PARTIAL_RELEASED,
	}
	if matched, detail := assetFreezeMatchesReservation(row, []*asset.AssetFreeze{freeze}); !matched {
		t.Fatalf("matching reservation rejected: %s", detail)
	}
	freeze.RemainAmount = "49"
	if matched, _ := assetFreezeMatchesReservation(row, []*asset.AssetFreeze{freeze}); matched {
		t.Fatal("amount conservation mismatch was accepted")
	}
}

func TestFailedReservationMayHaveNoAssetFreeze(t *testing.T) {
	row := &contractReservationAudit{
		Status: int64(trade.AssetReservationStatus_ASSET_RESERVATION_STATUS_FAILED),
	}
	if matched, detail := assetFreezeMatchesReservation(row, nil); !matched {
		t.Fatalf("definitively failed reservation without Asset freeze should match: %s", detail)
	}
}
