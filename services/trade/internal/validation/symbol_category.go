package validation

import (
	"fmt"

	"wklive/proto/common"
	"wklive/proto/trade"
)

// SymbolCategoryConfig keeps the Market category and trade product shape
// consistent. Crypto supports all existing trade product types; futures are
// derivative symbols; the remaining Market categories are spot-only for now.
func SymbolCategoryConfig(categoryType int64, productType common.ProductType, contractType common.ContractType, contractValueType trade.ContractValueType) error {
	if categoryType < 1 || categoryType > 6 {
		return fmt.Errorf("invalid category type: %d", categoryType)
	}

	if categoryType == 2 {
		if productType < common.ProductType_PRODUCT_TYPE_SPOT || productType > common.ProductType_PRODUCT_TYPE_SECONDS {
			return fmt.Errorf("invalid crypto product type: %d", productType)
		}
		return validateContractShape(productType, contractType, contractValueType)
	}

	if categoryType == 4 {
		if productType != common.ProductType_PRODUCT_TYPE_DERIVATIVE {
			return fmt.Errorf("future category must use derivative product type")
		}
		return validateContractShape(productType, contractType, contractValueType)
	}

	if productType != common.ProductType_PRODUCT_TYPE_SPOT ||
		contractType != common.ContractType_CONTRACT_TYPE_NOT_APPLICABLE ||
		contractValueType != trade.ContractValueType_CONTRACT_VALUE_TYPE_NOT_APPLICABLE {
		return fmt.Errorf("category %d only supports spot symbols", categoryType)
	}
	return nil
}

func validateContractShape(productType common.ProductType, contractType common.ContractType, contractValueType trade.ContractValueType) error {
	if productType != common.ProductType_PRODUCT_TYPE_DERIVATIVE {
		if contractType != common.ContractType_CONTRACT_TYPE_NOT_APPLICABLE ||
			contractValueType != trade.ContractValueType_CONTRACT_VALUE_TYPE_NOT_APPLICABLE {
			return fmt.Errorf("non-derivative symbols cannot use contract settings")
		}
		return nil
	}
	if (contractType != common.ContractType_CONTRACT_TYPE_PERPETUAL && contractType != common.ContractType_CONTRACT_TYPE_DELIVERY) ||
		(contractValueType != trade.ContractValueType_CONTRACT_VALUE_TYPE_LINEAR && contractValueType != trade.ContractValueType_CONTRACT_VALUE_TYPE_INVERSE) {
		return fmt.Errorf("derivative symbols require a valid contract type and value type")
	}
	return nil
}
