package logic

import (
	"wklive/proto/itick"
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
