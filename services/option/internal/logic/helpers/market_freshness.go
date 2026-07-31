package helpers

import "wklive/services/option/models"

func IsFreshTimestamp(snapshotTime, now, maxAge int64) bool {
	return snapshotTime > 0 &&
		snapshotTime <= now &&
		maxAge >= 0 &&
		now-snapshotTime <= maxAge
}

func IsUnderlyingFresh(market *models.TOptionMarket, now, maxAge int64) bool {
	return market != nil &&
		market.UnderlyingPrice.IsPositive() &&
		IsFreshTimestamp(market.UnderlyingSnapshotTime, now, maxAge)
}

func IsMarkFresh(market *models.TOptionMarket, now, maxAge int64) bool {
	return market != nil &&
		market.MarkPrice.IsPositive() &&
		IsFreshTimestamp(market.MarkSnapshotTime, now, maxAge)
}

func IsRiskMarketFresh(market *models.TOptionMarket, now, maxAge int64) bool {
	return IsUnderlyingFresh(market, now, maxAge) &&
		IsMarkFresh(market, now, maxAge)
}
