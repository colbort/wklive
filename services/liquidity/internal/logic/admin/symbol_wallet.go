package adminlogic

import (
	"wklive/proto/common"
	"wklive/proto/market"
)

func walletTypeForLiquidity(productType common.ProductType, categoryType int64) common.WalletType {
	if productType == common.ProductType_PRODUCT_TYPE_SPOT {
		if categoryType == int64(market.CategoryType_CATEGORY_TYPE_STOCK) {
			return common.WalletType_WALLET_TYPE_FUNDING
		}
		return common.WalletType_WALLET_TYPE_SPOT
	}
	return common.WalletType_WALLET_TYPE_CONTRACT
}
