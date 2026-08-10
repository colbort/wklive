package models

import "testing"

func TestKlineSourcePriorityOrder(t *testing.T) {
	rest := KlineSourcePriority(KlineSourceRest)
	exchangeRest := KlineSourcePriority(KlineSourceExchangeRest)
	derived := KlineSourcePriority(KlineSourceDerived)
	realtime := KlineSourcePriority(KlineSourceRealtime)
	if !(rest > exchangeRest && exchangeRest > derived && derived > realtime) {
		t.Fatalf("unexpected source priorities: rest=%d exchangeRest=%d derived=%d realtime=%d",
			rest, exchangeRest, derived, realtime)
	}
}
