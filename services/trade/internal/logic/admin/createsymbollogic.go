package adminlogic

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"wklive/common/conv"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/internal/validation"
	"wklive/services/trade/models"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateSymbolLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateSymbolLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateSymbolLogic {
	return &CreateSymbolLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建交易对
func (l *CreateSymbolLogic) CreateSymbol(in *trade.CreateSymbolReq) (*trade.CommonResp, error) {
	if in.CategoryType == stockCategoryType {
		product, base, err := resolveStockTenantProduct(l.ctx, l.svcCtx.MarketClient, in.TenantId, in.TenantProductId)
		if err != nil {
			return nil, err
		}
		if base != nil {
			return &trade.CommonResp{Base: base}, nil
		}
		applyStockTenantProduct(in, product)
	} else {
		in.TenantProductId = 0
	}
	if in.CategoryType < 1 || in.CategoryType > 6 {
		return &trade.CommonResp{Base: helper.ErrResp(i18n.ParamError, fmt.Sprintf("invalid category type: %d", in.CategoryType))}, nil
	}
	if err := validation.SymbolCategoryConfig(in.CategoryType, in.ProductType, in.ContractType, in.ContractValueType); err != nil {
		return &trade.CommonResp{Base: helper.ErrResp(i18n.ParamError, err.Error())}, nil
	}
	if err := validation.SymbolTradingTimeline(in.ProductType, in.ContractType, in.ListingTime, in.TradingStartTime, in.TradingEndTime); err != nil {
		return &trade.CommonResp{Base: helper.ErrResp(i18n.ParamError, err.Error())}, nil
	}
	market := strings.ToUpper(strings.TrimSpace(in.Market))
	if market == "" {
		return &trade.CommonResp{Base: helper.ErrResp(i18n.ParamError, "market is required")}, nil
	}
	symbol := strings.ToUpper(strings.TrimSpace(in.Symbol))
	if symbol == "" {
		return &trade.CommonResp{Base: helper.ErrResp(i18n.ParamError, "symbol is required")}, nil
	}
	exists, err := l.svcCtx.TradeSymbolModel.FindOneByTenantIdCategoryTypeMarketSymbolProductTypeContractTypeContractValueType(l.ctx, in.TenantId, in.CategoryType, market, symbol, int64(in.ProductType), int64(in.ContractType), int64(in.ContractValueType))
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if exists != nil {
		return &trade.CommonResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	now := utils.NowMillis()
	data := &models.TTradeSymbol{
		TenantId:          in.TenantId,
		TenantProductId:   in.TenantProductId,
		CategoryType:      in.CategoryType,
		Market:            market,
		Symbol:            symbol,
		DisplaySymbol:     in.DisplaySymbol,
		ProductType:       int64(in.ProductType),
		BaseAsset:         in.BaseAsset,
		QuoteAsset:        in.QuoteAsset,
		SettleAsset:       in.SettleAsset,
		ContractType:      int64(in.ContractType),
		ContractValueType: int64(in.ContractValueType),
		MarginAsset:       in.MarginAsset,
		Status:            int64(in.Status),
		PriceScale:        int64(in.PriceScale),
		QtyScale:          int64(in.QtyScale),
		ListingTime:       in.ListingTime,
		TradingStartTime:  in.TradingStartTime,
		TradingEndTime:    in.TradingEndTime,
		Sort:              int64(in.Sort),
		Remark:            in.Remark,
		CreateTimes:       now,
		UpdateTimes:       now,
	}
	decimalFields := []struct {
		raw    string
		target *decimal.Decimal
	}{
		{in.MinPrice, &data.MinPrice},
		{in.MaxPrice, &data.MaxPrice},
		{in.PriceTick, &data.PriceTick},
		{in.MinQty, &data.MinQty},
		{in.MaxQty, &data.MaxQty},
		{in.QtyStep, &data.QtyStep},
		{in.MinNotional, &data.MinNotional},
		{in.MaxNotional, &data.MaxNotional},
	}
	for _, field := range decimalFields {
		value, parseErr := conv.ParseDecimalField(field.raw)
		if parseErr != nil {
			return &trade.CommonResp{Base: helper.ErrResp(i18n.ParamError, parseErr.Error())}, nil
		}
		*field.target = value
	}
	if _, err = l.svcCtx.TradeSymbolModel.Insert(l.ctx, data); err != nil {
		return nil, err
	}

	return &trade.CommonResp{Base: helper.OkResp()}, nil
}
