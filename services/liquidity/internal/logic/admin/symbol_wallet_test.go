package adminlogic

import (
	"testing"

	"wklive/proto/common"
	"wklive/proto/market"
)

func TestWalletTypeForLiquidity(t *testing.T) {
	tests := []struct {
		name         string
		productType  common.ProductType
		categoryType int64
		want         common.WalletType
	}{
		{
			name:         "crypto spot uses spot wallet",
			productType:  common.ProductType_PRODUCT_TYPE_SPOT,
			categoryType: int64(market.CategoryType_CATEGORY_TYPE_CRYPTO),
			want:         common.WalletType_WALLET_TYPE_SPOT,
		},
		{
			name:         "stock spot uses funding wallet",
			productType:  common.ProductType_PRODUCT_TYPE_SPOT,
			categoryType: int64(market.CategoryType_CATEGORY_TYPE_STOCK),
			want:         common.WalletType_WALLET_TYPE_FUNDING,
		},
		{
			name:         "derivative uses contract wallet",
			productType:  common.ProductType_PRODUCT_TYPE_DERIVATIVE,
			categoryType: int64(market.CategoryType_CATEGORY_TYPE_STOCK),
			want:         common.WalletType_WALLET_TYPE_CONTRACT,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := walletTypeForLiquidity(test.productType, test.categoryType); got != test.want {
				t.Fatalf("walletTypeForLiquidity(%s, %d)=%s want=%s", test.productType, test.categoryType, got, test.want)
			}
		})
	}
}
