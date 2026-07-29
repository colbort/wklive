package helpers

import (
	"context"
	"strconv"
	"strings"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/common"
	pb "wklive/proto/market"
	"wklive/services/market/internal/market/types"
	"wklive/services/market/models"
)

func AdminTenantWriteScopeResp(ctx context.Context, currentTenantId int64, notAllowedCode int32) (*common.RespBase, error) {
	_, allowed, forbidden, err := utils.ResolveAdminTenantWriteScopeFromMd(ctx, currentTenantId)
	if err != nil {
		return nil, i18n.StatusError(ctx, i18n.UserNotFound)
	}
	if forbidden {
		return helper.ErrResp(i18n.PermissionDenied, i18n.Translate(i18n.PermissionDenied, ctx)), nil
	}
	if !allowed {
		return helper.ErrResp(notAllowedCode, i18n.Translate(notAllowedCode, ctx)), nil
	}
	return nil, nil
}

func ToCategoryProto(item *models.TMarketCategory) *pb.MarketCategory {
	if item == nil {
		return nil
	}
	return &pb.MarketCategory{
		Id:           item.Id,
		CategoryType: pb.CategoryType(item.CategoryType),
		CategoryCode: item.CategoryCode,
		CategoryName: item.CategoryName,
		Enabled:      common.Enable(item.Enabled),
		AppVisible:   common.Switch(item.AppVisible),
		SyncPriority: pb.SyncKlinePriority(item.SyncPriority),
		Sort:         item.Sort,
		Icon:         item.Icon,
		Remark:       item.Remark,
		CreateTimes:  item.CreateTimes,
		UpdateTimes:  item.UpdateTimes,
	}
}

func ToProductProto(item *models.TMarketProduct) *pb.MarketProduct {
	if item == nil {
		return nil
	}
	return &pb.MarketProduct{
		Id:           item.Id,
		CategoryType: pb.CategoryType(item.CategoryType),
		CategoryName: item.CategoryName,
		CategoryCode: item.CategoryCode,
		Market:       item.Market,
		Symbol:       item.Symbol,
		Code:         item.Code,
		Name:         item.Name,
		DisplayName:  item.DisplayName,
		BaseCoin:     item.BaseCoin,
		QuoteCoin:    item.QuoteCoin,
		Enabled:      common.Enable(item.Enabled),
		AppVisible:   common.Switch(item.AppVisible),
		SyncPriority: pb.SyncKlinePriority(item.SyncPriority),
		Sort:         item.Sort,
		Icon:         item.Icon,
		Remark:       item.Remark,
		CreateTimes:  item.CreateTimes,
		UpdateTimes:  item.UpdateTimes,
	}
}

func ToQuoteProto(item *models.TMarketQuote) *pb.Quote {
	if item == nil {
		return nil
	}
	return &pb.Quote{
		CategoryCode:   item.CategoryCode,
		Market:         item.Market,
		Symbol:         item.Symbol,
		LastPrice:      item.LastPrice.String(),
		OpenPrice:      item.OpenPrice.String(),
		HighPrice:      item.HighPrice.String(),
		LowPrice:       item.LowPrice.String(),
		PrevClosePrice: item.PrevClosePrice.String(),
		ChangeValue:    item.ChangeValue.String(),
		ChangeRate:     item.ChangeRate.String(),
		Volume:         item.Volume.String(),
		Turnover:       item.Turnover.String(),
		QuoteTs:        item.QuoteTs,
		TradeStatus:    item.TradeStatus,
	}
}

func ToQuotePayloadProto(categoryCode, market, symbol string, item *types.QuotePayload) *pb.Quote {
	if item == nil {
		return nil
	}
	return &pb.Quote{
		CategoryCode: categoryCode,
		Market:       market,
		Symbol:       symbol,
		LastPrice:    FormatMarketDecimal(item.LastPrice),
		OpenPrice:    FormatMarketDecimal(item.Open),
		HighPrice:    FormatMarketDecimal(item.High),
		LowPrice:     FormatMarketDecimal(item.LastPrice),
		Volume:       FormatMarketDecimal(item.Volume),
		Turnover:     FormatMarketDecimal(item.Turnover),
		QuoteTs:      item.Ts,
	}
}

func ToKlineProto(kType pb.KlineType, item *models.CoinKline) *pb.Kline {
	if item == nil {
		return nil
	}
	return &pb.Kline{
		CategoryCode:  item.CategoryCode,
		Market:        item.Market,
		Symbol:        item.Symbol,
		KType:         kType,
		Ts:            item.Ts,
		Open:          FormatMarketDecimal(item.Open),
		High:          FormatMarketDecimal(item.High),
		Low:           FormatMarketDecimal(item.Low),
		Close:         FormatMarketDecimal(item.Close),
		Volume:        FormatMarketDecimal(item.Volume),
		Turnover:      FormatMarketDecimal(item.Turnover),
		Source:        item.Source,
		Revision:      item.Revision,
		IsClosed:      item.IsClosed,
		Confirmed:     item.Confirmed,
		ActualCount:   item.ActualCount,
		ExpectedCount: item.ExpectedCount,
	}
}

func FormatMarketDecimal(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func ToTenantCategoryProto(item *models.TMarketTenantCategory, category *models.TMarketCategory) *pb.MarketTenantCategory {
	if item == nil {
		return nil
	}

	data := &pb.MarketTenantCategory{
		Id:          item.Id,
		TenantId:    item.TenantId,
		CategoryId:  item.CategoryId,
		Enabled:     common.Enable(item.Enabled),
		AppVisible:  common.Switch(item.AppVisible),
		Sort:        item.Sort,
		Remark:      item.Remark,
		CreateTimes: item.CreateTimes,
		UpdateTimes: item.UpdateTimes,
	}
	if category != nil {
		data.CategoryType = pb.CategoryType(category.CategoryType)
		data.CategoryCode = category.CategoryCode
		data.CategoryName = category.CategoryName
		data.Icon = category.Icon
	}
	return data
}

func ToTenantProductProto(item *models.TMarketTenantProduct, product *models.TMarketProduct) *pb.MarketTenantProduct {
	if item == nil {
		return nil
	}

	data := &pb.MarketTenantProduct{
		Id:          item.Id,
		TenantId:    item.TenantId,
		ProductId:   item.ProductId,
		Enabled:     common.Enable(item.Enabled),
		AppVisible:  common.Switch(item.AppVisible),
		Sort:        item.Sort,
		Remark:      item.Remark,
		CreateTimes: item.CreateTimes,
		UpdateTimes: item.UpdateTimes,
	}
	if product != nil {
		data.CategoryType = pb.CategoryType(product.CategoryType)
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
	if item.DisplayName != "" {
		data.DisplayName = item.DisplayName
	}
	return data
}

func CategoryTypeCode(categoryType pb.CategoryType) string {
	switch categoryType {
	case pb.CategoryType_CATEGORY_TYPE_FOREX:
		return "forex"
	case pb.CategoryType_CATEGORY_TYPE_CRYPTO:
		return "crypto"
	case pb.CategoryType_CATEGORY_TYPE_STOCK:
		return "stock"
	case pb.CategoryType_CATEGORY_TYPE_FUTURE:
		return "future"
	case pb.CategoryType_CATEGORY_TYPE_INDICES:
		return "indices"
	case pb.CategoryType_CATEGORY_TYPE_FUND:
		return "fund"
	default:
		return ""
	}
}

func StatusMatches(filter int32, actual int64) bool {
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

func KeywordMatches(keyword string, parts ...string) bool {
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
