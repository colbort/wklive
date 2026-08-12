package adminlogic

import (
	"context"
	"fmt"
	"strings"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/common"
	"wklive/proto/market"
	"wklive/proto/trade"
)

const stockCategoryType int64 = 3

func resolveStockTenantProduct(
	ctx context.Context,
	client market.MarketClient,
	tenantID int64,
	tenantProductID int64,
) (*market.MarketTenantProduct, *common.RespBase, error) {
	if tenantProductID <= 0 {
		return nil, helper.ErrResp(i18n.ParamError, "tenantProductId is required for stock symbol"), nil
	}
	resp, err := client.ResolveTenantProduct(ctx, &market.ResolveTenantProductReq{
		TenantId:        tenantID,
		TenantProductId: tenantProductID,
	})
	if err != nil {
		return nil, nil, err
	}
	if resp == nil || resp.GetBase() == nil || resp.GetBase().GetCode() != 200 || resp.GetData() == nil {
		if resp != nil && resp.GetBase() != nil {
			return nil, resp.GetBase(), nil
		}
		return nil, helper.ErrResp(i18n.BusinessDataNotFound, "enabled tenant product not found"), nil
	}
	if resp.GetData().GetCategoryType() != market.CategoryType_CATEGORY_TYPE_STOCK {
		return nil, helper.ErrResp(i18n.ParamError, "tenant product is not a stock"), nil
	}
	if strings.TrimSpace(resp.GetData().GetBaseCoin()) == "" || strings.TrimSpace(resp.GetData().GetQuoteCoin()) == "" {
		return nil, helper.ErrResp(i18n.ParamError, "stock product assets are incomplete"), nil
	}
	return resp.GetData(), nil, nil
}

func applyStockTenantProduct(in *trade.CreateSymbolReq, product *market.MarketTenantProduct) {
	in.CategoryType = stockCategoryType
	in.Market = strings.ToUpper(strings.TrimSpace(product.GetMarket()))
	in.Symbol = strings.ToUpper(strings.TrimSpace(product.GetSymbol()))
	if strings.TrimSpace(in.DisplaySymbol) == "" {
		in.DisplaySymbol = in.Symbol
	}
	in.ProductType = common.ProductType_PRODUCT_TYPE_SPOT
	in.BaseAsset = strings.ToUpper(strings.TrimSpace(product.GetBaseCoin()))
	in.QuoteAsset = strings.ToUpper(strings.TrimSpace(product.GetQuoteCoin()))
	in.SettleAsset = in.QuoteAsset
	in.MarginAsset = ""
	in.ContractType = common.ContractType_CONTRACT_TYPE_NOT_APPLICABLE
	in.ContractValueType = trade.ContractValueType_CONTRACT_VALUE_TYPE_NOT_APPLICABLE
}

func stockProductMismatchMessage(product *market.MarketTenantProduct) string {
	return fmt.Sprintf("stock tenant product mismatch: tenant_product_id=%d market=%s symbol=%s",
		product.GetId(), product.GetMarket(), product.GetSymbol())
}
