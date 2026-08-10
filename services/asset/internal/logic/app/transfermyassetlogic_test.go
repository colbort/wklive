package applogic

import (
	"context"
	"errors"
	"testing"

	"wklive/services/asset/models"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

func TestExchangeTransferAmountRoutesByCoinType(t *testing.T) {
	quotes := map[string]decimal.Decimal{
		"forex:GB:USDCNY":   decimal.RequireFromString("6"),
		"crypto:BA:BTCUSDT": decimal.RequireFromString("60000"),
		"crypto:BA:ETHUSDT": decimal.RequireFromString("3000"),
	}
	reader := func(_ context.Context, categoryCode, market, symbol string) (decimal.Decimal, error) {
		price, ok := quotes[categoryCode+":"+market+":"+symbol]
		if !ok {
			return decimal.Zero, redis.Nil
		}
		return price, nil
	}

	tests := []struct {
		name       string
		fromCoin   string
		fromType   int64
		toCoin     string
		toType     int64
		amount     string
		toDecimals int64
		want       string
	}{
		{
			name:       "same coin account transfer stays one to one",
			fromCoin:   "USDT",
			fromType:   assetCoinTypeCrypto,
			toCoin:     "USDT",
			toType:     assetCoinTypeCrypto,
			amount:     "12.345678",
			toDecimals: 2,
			want:       "12.345678",
		},
		{
			name:       "fiat to fiat uses direct forex quote",
			fromCoin:   "USD",
			fromType:   assetCoinTypeFiat,
			toCoin:     "CNY",
			toType:     assetCoinTypeFiat,
			amount:     "2.25",
			toDecimals: 2,
			want:       "13.5",
		},
		{
			name:       "fiat to fiat uses inverse forex quote",
			fromCoin:   "CNY",
			fromType:   assetCoinTypeFiat,
			toCoin:     "USD",
			toType:     assetCoinTypeFiat,
			amount:     "6",
			toDecimals: 2,
			want:       "1",
		},
		{
			name:       "crypto to crypto crosses through USDT",
			fromCoin:   "BTC",
			fromType:   assetCoinTypeCrypto,
			toCoin:     "ETH",
			toType:     assetCoinTypeCrypto,
			amount:     "1",
			toDecimals: 8,
			want:       "20",
		},
		{
			name:       "fiat to crypto crosses USD and USDT",
			fromCoin:   "USD",
			fromType:   assetCoinTypeFiat,
			toCoin:     "BTC",
			toType:     assetCoinTypeCrypto,
			amount:     "60000",
			toDecimals: 8,
			want:       "1",
		},
		{
			name:       "crypto to fiat crosses USDT and USD",
			fromCoin:   "BTC",
			fromType:   assetCoinTypeCrypto,
			toCoin:     "CNY",
			toType:     assetCoinTypeFiat,
			amount:     "1",
			toDecimals: 2,
			want:       "360000",
		},
		{
			name:       "USDT to USD uses accounting peg",
			fromCoin:   "USDT",
			fromType:   assetCoinTypeCrypto,
			toCoin:     "USD",
			toType:     assetCoinTypeFiat,
			amount:     "100",
			toDecimals: 2,
			want:       "100",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := exchangeTransferAmount(
				context.Background(),
				&models.TAssetCoinConfig{Coin: test.fromCoin, CoinType: test.fromType, DecimalPlaces: 18},
				&models.TAssetCoinConfig{Coin: test.toCoin, CoinType: test.toType, DecimalPlaces: test.toDecimals},
				decimal.RequireFromString(test.amount),
				reader,
			)
			if err != nil {
				t.Fatalf("exchangeTransferAmount() error = %v", err)
			}
			if !got.Equal(decimal.RequireFromString(test.want)) {
				t.Fatalf("exchangeTransferAmount() = %s, want %s", got, test.want)
			}
		})
	}
}

func TestExchangeTransferAmountRejectsMissingRate(t *testing.T) {
	reader := func(context.Context, string, string, string) (decimal.Decimal, error) {
		return decimal.Zero, redis.Nil
	}

	_, err := exchangeTransferAmount(
		context.Background(),
		&models.TAssetCoinConfig{Coin: "USDC", CoinType: assetCoinTypeCrypto, DecimalPlaces: 8},
		&models.TAssetCoinConfig{Coin: "CNY", CoinType: assetCoinTypeFiat, DecimalPlaces: 2},
		decimal.NewFromInt(1),
		reader,
	)
	if !errors.Is(err, errExchangeRateUnavailable) {
		t.Fatalf("exchangeTransferAmount() error = %v, want %v", err, errExchangeRateUnavailable)
	}
}

func TestResolveExchangeRatePropagatesQuoteInfrastructureError(t *testing.T) {
	wantErr := errors.New("market cache unavailable")
	reader := func(context.Context, string, string, string) (decimal.Decimal, error) {
		return decimal.Zero, wantErr
	}

	_, err := resolveExchangeRate(
		context.Background(),
		"BTC",
		assetCoinTypeCrypto,
		"ETH",
		assetCoinTypeCrypto,
		reader,
	)
	if !errors.Is(err, wantErr) {
		t.Fatalf("resolveExchangeRate() error = %v, want %v", err, wantErr)
	}
}
