package applogic

import (
	"testing"

	"wklive/proto/trade"
	"wklive/services/trade/models"

	"github.com/shopspring/decimal"
)

func TestInverseBuyReservationRiskPrice(t *testing.T) {
	limit := decimal.RequireFromString("51000")
	tests := []struct {
		name  string
		sells []*models.TTradeOrder
		want  string
	}{
		{name: "resting order uses own limit", want: "51000"},
		{
			name: "crossing inverse buy uses lowest maker price",
			sells: []*models.TTradeOrder{
				{OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT), Price: decimal.RequireFromString("50500")},
				{OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT), Price: decimal.RequireFromString("50000")},
			},
			want: "50000",
		},
		{
			name: "non crossing and market asks do not change risk price",
			sells: []*models.TTradeOrder{
				{OrderType: int64(trade.OrderType_ORDER_TYPE_MARKET), Price: decimal.Zero},
				{OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT), Price: decimal.RequireFromString("52000")},
			},
			want: "51000",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := inverseBuyReservationRiskPrice(limit, tt.sells)
			if !got.Equal(decimal.RequireFromString(tt.want)) {
				t.Fatalf("risk price=%s, want %s", got, tt.want)
			}
		})
	}
}
