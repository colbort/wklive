package assetlogic

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestInsuranceCoverageSupportsPartialAndZero(t *testing.T) {
	cases := []struct{ r, a, w string }{{"10", "20", "10"}, {"10", "3.25", "3.25"}, {"10", "0", "0"}}
	for _, c := range cases {
		got := insuranceCoverage(decimal.RequireFromString(c.r), decimal.RequireFromString(c.a))
		if got.String() != c.w {
			t.Fatalf("coverage(%s,%s)=%s want %s", c.r, c.a, got, c.w)
		}
	}
}
