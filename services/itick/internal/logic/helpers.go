package logic

import (
	"strings"

	"wklive/proto/itick"
	"wklive/proto/system"
	"wklive/services/itick/internal/socket/client"
	"wklive/services/itick/models"
)

func toCategoryProto(item *models.TItickCategory) *itick.ItickCategory {
	if item == nil {
		return nil
	}
	return &itick.ItickCategory{
		Id:           item.Id,
		CategoryType: itick.CategoryType(item.CategoryType),
		CategoryCode: item.CategoryCode,
		CategoryName: item.CategoryName,
		Enabled:      item.Enabled,
		AppVisible:   item.AppVisible,
		Sort:         item.Sort,
		Icon:         item.Icon,
		Remark:       item.Remark,
		CreateTimes:  item.CreateTimes,
		UpdateTimes:  item.UpdateTimes,
	}
}

func toProductProto(item *models.TItickProduct) *itick.ItickProduct {
	if item == nil {
		return nil
	}
	return &itick.ItickProduct{
		Id:           item.Id,
		CategoryType: itick.CategoryType(item.CategoryType),
		CategoryName: item.CategoryName,
		CategoryCode: item.CategoryCode,
		Market:       item.Market,
		Symbol:       item.Symbol,
		Code:         item.Code,
		Name:         item.Name,
		DisplayName:  item.DisplayName,
		BaseCoin:     item.BaseCoin,
		QuoteCoin:    item.QuoteCoin,
		Enabled:      item.Enabled,
		AppVisible:   item.AppVisible,
		Sort:         item.Sort,
		Icon:         item.Icon,
		Remark:       item.Remark,
		CreateTimes:  item.CreateTimes,
		UpdateTimes:  item.UpdateTimes,
	}
}

func toQuoteProto(item *models.TItickQuote) *itick.Quote {
	if item == nil {
		return nil
	}
	return &itick.Quote{
		CategoryCode:   item.CategoryCode,
		Market:         item.Market,
		Symbol:         item.Symbol,
		LastPrice:      item.LastPrice,
		OpenPrice:      item.OpenPrice,
		HighPrice:      item.HighPrice,
		LowPrice:       item.LowPrice,
		PrevClosePrice: item.PrevClosePrice,
		ChangeValue:    item.ChangeValue,
		ChangeRate:     item.ChangeRate,
		Volume:         item.Volume,
		Turnover:       item.Turnover,
		QuoteTs:        item.QuoteTs,
		TradeStatus:    item.TradeStatus,
	}
}

func toQuotePayloadProto(categoryCode, market, symbol string, item *client.QuotePayload) *itick.Quote {
	if item == nil {
		return nil
	}
	return &itick.Quote{
		CategoryCode: categoryCode,
		Market:       market,
		Symbol:       symbol,
		LastPrice:    item.LastPrice,
		OpenPrice:    item.Open,
		HighPrice:    item.High,
		LowPrice:     item.LastPrice,
		Volume:       item.Volume,
		Turnover:     item.Turnover,
		QuoteTs:      item.Ts,
	}
}

func toKlineProto(kType itick.KlineType, item *models.CoinKline) *itick.Kline {
	if item == nil {
		return nil
	}
	return &itick.Kline{
		CategoryCode: item.CategoryCode,
		Market:       item.Market,
		Symbol:       item.Symbol,
		KType:        kType,
		Ts:           item.Ts,
		Open:         item.Open,
		High:         item.High,
		Low:          item.Low,
		Close:        item.Close,
		Volume:       item.Volume,
		Turnover:     item.Turnover,
	}
}

func toTenantCategoryProto(item *models.TItickTenantCategory, category *models.TItickCategory, tenant *system.SysTenantItem) *itick.ItickTenantCategory {
	if item == nil {
		return nil
	}

	data := &itick.ItickTenantCategory{
		Id:          item.Id,
		TenantId:    item.TenantId,
		CategoryId:  item.CategoryId,
		Enabled:     item.Enabled,
		AppVisible:  item.AppVisible,
		Sort:        item.Sort,
		Remark:      item.Remark,
		CreateTimes: item.CreateTimes,
		UpdateTimes: item.UpdateTimes,
	}
	if category != nil {
		data.CategoryType = itick.CategoryType(category.CategoryType)
		data.CategoryCode = category.CategoryCode
		data.CategoryName = category.CategoryName
		data.Icon = category.Icon
	}
	if tenant != nil {
		data.TenantName = tenant.TenantName
	}
	return data
}

func toTenantProductProto(item *models.TItickTenantProduct, product *models.TItickProduct, tenant *system.SysTenantItem) *itick.ItickTenantProduct {
	if item == nil {
		return nil
	}

	data := &itick.ItickTenantProduct{
		Id:          item.Id,
		TenantId:    item.TenantId,
		ProductId:   item.ProductId,
		Enabled:     item.Enabled,
		AppVisible:  item.AppVisible,
		Sort:        item.Sort,
		Remark:      item.Remark,
		CreateTimes: item.CreateTimes,
		UpdateTimes: item.UpdateTimes,
	}
	if product != nil {
		data.CategoryType = itick.CategoryType(product.CategoryType)
		data.CategoryCode = product.CategoryCode
		data.CategoryName = product.CategoryName
		data.Market = product.Market
		data.Symbol = product.Symbol
		data.Code = product.Code
		data.Name = product.Name
		data.DisplayName = product.DisplayName
		data.BaseCoin = product.BaseCoin
		data.QuoteCoin = product.QuoteCoin
		data.Icon = product.Icon
	}
	if tenant != nil {
		data.TenantName = tenant.TenantName
	}
	return data
}

func categoryTypeCode(categoryType itick.CategoryType) string {
	switch categoryType {
	case itick.CategoryType_CATEGORY_TYPE_FOREX:
		return "forex"
	case itick.CategoryType_CATEGORY_TYPE_CRYPTO:
		return "crypto"
	case itick.CategoryType_CATEGORY_TYPE_STOCK:
		return "stock"
	case itick.CategoryType_CATEGORY_TYPE_FUTURE:
		return "future"
	case itick.CategoryType_CATEGORY_TYPE_INDICES:
		return "indices"
	case itick.CategoryType_CATEGORY_TYPE_FUND:
		return "fund"
	default:
		return ""
	}
}

func statusMatches(filter int32, actual int64) bool {
	switch filter {
	case 0:
		return true
	case 1:
		return actual == 1
	case 2:
		return actual == 0 || actual == 2
	default:
		return true
	}
}

func keywordMatches(keyword string, parts ...string) bool {
	keyword = strings.ToLower(strings.TrimSpace(keyword))
	if keyword == "" {
		return true
	}
	for _, part := range parts {
		if strings.Contains(strings.ToLower(part), keyword) {
			return true
		}
	}
	return false
}
