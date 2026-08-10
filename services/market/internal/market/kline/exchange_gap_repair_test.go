package kline

import (
	"net/url"
	"testing"

	"wklive/services/market/models"
)

func TestSupportsExchangeKlineFallback(t *testing.T) {
	tests := []struct {
		name string
		job  *GapRepairJob
		want bool
	}{
		{name: "binance crypto", job: &GapRepairJob{Category: "crypto", Market: "BA", Exchange: "Binance"}, want: true},
		{name: "case insensitive", job: &GapRepairJob{Category: "CRYPTO", Market: "ba", Exchange: "binance"}, want: true},
		{name: "other exchange", job: &GapRepairJob{Category: "crypto", Market: "OK", Exchange: "OKX"}},
		{name: "other category", job: &GapRepairJob{Category: "forex", Market: "BA", Exchange: "Binance"}},
		{name: "nil", job: nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := supportsExchangeKlineFallback(test.job); got != test.want {
				t.Fatalf("supportsExchangeKlineFallback()=%v want=%v", got, test.want)
			}
		})
	}
}

func TestBuildBinanceSpotKlineURL(t *testing.T) {
	rawURL, err := buildBinanceSpotKlineURL(" btcusdt ", 1786004460000, 1786004700000)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Scheme != "https" || parsed.Host != "api.binance.com" || parsed.Path != "/api/v3/klines" {
		t.Fatalf("unexpected endpoint: %s", rawURL)
	}
	want := map[string]string{
		"symbol": "BTCUSDT", "interval": "1m", "startTime": "1786004460000",
		"endTime": "1786004700000", "limit": "1000",
	}
	for key, value := range want {
		if got := parsed.Query().Get(key); got != value {
			t.Fatalf("query %s=%q want=%q", key, got, value)
		}
	}
}

func TestParseBinanceSpotKlines(t *testing.T) {
	job := &GapRepairJob{Category: "crypto", Market: "BA", Exchange: "Binance", Symbol: "BTCUSDT"}
	raw := []byte(`[
		[1786004460000,"64804.71000000","64832.00000000","64804.71000000","64824.90000000","5.07542000",1786004519999,"328962.87829800",1623,"4.36858000","283140.55884240","0"],
		[1786004520000,"64824.89000000","64824.89000000","64818.76000000","64818.76000000","3.88064000",1786004579999,"251547.51228350",465,"1.57331000","101984.76334990","0"]
	]`)
	list, err := parseBinanceSpotKlines(raw, job, 1786004460000, 1786004520000)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Fatalf("len=%d want=2", len(list))
	}
	first := list[0]
	if first.Ts != 1786004460000 || first.Open != 64804.71 || first.High != 64832 ||
		first.Low != 64804.71 || first.Close != 64824.9 || first.Volume != 5.07542 ||
		first.Turnover != 328962.878298 || first.Source != models.KlineSourceExchangeRest ||
		!first.IsClosed || !first.Confirmed {
		t.Fatalf("unexpected first K line: %+v", first)
	}
}

func TestParseBinanceSpotKlinesRejectsInvalidOHLC(t *testing.T) {
	job := &GapRepairJob{Category: "crypto", Market: "BA", Exchange: "Binance", Symbol: "BTCUSDT"}
	raw := []byte(`[[1786004460000,"10","9","8","10","1",1786004519999,"10"]]`)
	if _, err := parseBinanceSpotKlines(raw, job, 1786004460000, 1786004460000); err == nil {
		t.Fatal("expected invalid OHLC error")
	}
}
