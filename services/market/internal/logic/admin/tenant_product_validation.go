package adminlogic

import (
	"context"
	"fmt"
	"strings"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/common"
	"wklive/services/market/internal/pkg/utils"
	"wklive/services/market/models"
)

func validateSelectableProduct(ctx context.Context, product *models.TItickProduct) *common.RespBase {
	if product == nil {
		return helper.ErrResp(i18n.ProductNotFound, i18n.Translate(i18n.ProductNotFound, ctx))
	}
	if product.Enabled != 1 {
		return helper.ErrResp(i18n.ParamError, fmt.Sprintf("product %d is disabled", product.Id))
	}
	if utils.NormalizeCategory(product.CategoryCode) != "stock" {
		return nil
	}
	exchange := strings.TrimSpace(product.Exchange)
	if exchange == "" {
		return nil
	}
	if utils.StockExchangeMatchesMarket(product.Market, exchange) {
		return nil
	}
	return helper.ErrResp(i18n.ParamError, fmt.Sprintf(
		"invalid stock product market/exchange: product_id=%d market=%s exchange=%s symbol=%s",
		product.Id, product.Market, product.Exchange, product.Symbol,
	))
}
