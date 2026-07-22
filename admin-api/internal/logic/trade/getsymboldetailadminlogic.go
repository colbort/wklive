// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package trade

import (
	"context"
	"errors"

	"wklive/admin-api/internal/logicutil"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/common/utils"
	"wklive/proto/trade"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSymbolDetailAdminLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetSymbolDetailAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSymbolDetailAdminLogic {
	return &GetSymbolDetailAdminLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSymbolDetailAdminLogic) GetSymbolDetailAdmin(req *types.GetSymbolDetailAdminReq) (resp *types.GetSymbolDetailAdminResp, err error) {
	resp, err = logicutil.Proxy[types.GetSymbolDetailAdminResp](l.ctx, req, l.svcCtx.TradeCli.GetSymbolDetailAdmin)
	if err != nil || resp == nil || resp.Code != 200 {
		if err != nil {
			return logicutil.SystemErrorResp[types.GetSymbolDetailAdminResp](l.ctx, err)
		}
		if resp == nil {
			return logicutil.SystemErrorResp[types.GetSymbolDetailAdminResp](l.ctx, errors.New("empty symbol detail response"))
		}
		return resp, nil
	}

	tenantId := req.TenantId
	if tenantId == 0 {
		tenantId = resp.Data.Symbol.TenantId
	}
	ctx := context.WithValue(l.ctx, utils.CtxKeyTenantId, tenantId)
	detail, err := logicutil.Proxy[trade.GetSymbolDetailAdminResp](ctx, &types.GetSymbolDetailAdminReq{Id: req.Id}, l.svcCtx.TradeCli.GetSymbolDetailAdmin)
	if err != nil || detail == nil {
		if err != nil {
			return logicutil.SystemErrorResp[types.GetSymbolDetailAdminResp](l.ctx, err)
		}
		if detail == nil || detail.Base == nil {
			return logicutil.SystemErrorResp[types.GetSymbolDetailAdminResp](l.ctx, errors.New("empty app symbol detail response"))
		}
		resp.Code = detail.GetBase().GetCode()
		resp.Msg = detail.GetBase().GetMsg()
		return resp, nil
	}
	if detail.GetData().GetSpot() != nil {
		resp.Data.Spot = tradeSymbolSpotToTypes(detail.GetData().GetSpot())
	}
	if detail.GetData().GetContract() != nil {
		resp.Data.Contract = tradeSymbolContractToTypes(detail.GetData().GetContract())
	}
	if detail.GetData().GetSymbol() != nil {
		resp.Data.Symbol = tradeSymbolToTypes(detail.GetData().GetSymbol())
	}
	if detail.GetData().GetLeverageConfigs() != nil {
		resp.Data.LeverageConfigs = make([]types.TradeSymbolLeverageConfig, 0, len(detail.GetData().GetLeverageConfigs()))
		for _, item := range detail.GetData().GetLeverageConfigs() {
			if item != nil {
				resp.Data.LeverageConfigs = append(resp.Data.LeverageConfigs, tradeSymbolLeverageConfigToTypes(item))
			}
		}
	}
	if detail.GetData().GetSecondsConfigs() != nil {
		resp.Data.SecondsConfigs = make([]types.TradeSymbolSeconds, 0, len(detail.GetData().GetSecondsConfigs()))
		for _, item := range detail.GetData().GetSecondsConfigs() {
			if item != nil {
				resp.Data.SecondsConfigs = append(resp.Data.SecondsConfigs, tradeSymbolSecondsToTypes(item))
			}
		}
	}
	return resp, nil
}

func tradeSymbolToTypes(symbol *trade.TradeSymbol) types.TradeSymbol {
	return types.TradeSymbol{
		Id:                symbol.GetId(),
		TenantId:          symbol.GetTenantId(),
		Symbol:            symbol.GetSymbol(),
		DisplaySymbol:     symbol.GetDisplaySymbol(),
		ProductType:       int64(symbol.GetProductType()),
		BaseAsset:         symbol.GetBaseAsset(),
		QuoteAsset:        symbol.GetQuoteAsset(),
		SettleAsset:       symbol.GetSettleAsset(),
		ContractType:      int64(symbol.GetContractType()),
		ContractValueType: int64(symbol.GetContractValueType()),
		MarginAsset:       symbol.GetMarginAsset(),
		Status:            int64(symbol.GetStatus()),
		PriceScale:        symbol.GetPriceScale(),
		QtyScale:          symbol.GetQtyScale(),
		MinPrice:          symbol.GetMinPrice(),
		MaxPrice:          symbol.GetMaxPrice(),
		PriceTick:         symbol.GetPriceTick(),
		MinQty:            symbol.GetMinQty(),
		MaxQty:            symbol.GetMaxQty(),
		QtyStep:           symbol.GetQtyStep(),
		MinNotional:       symbol.GetMinNotional(),
		MaxNotional:       symbol.GetMaxNotional(),
		ListingTime:       symbol.GetListingTime(),
		TradingStartTime:  symbol.GetTradingStartTime(),
		TradingEndTime:    symbol.GetTradingEndTime(),
		Sort:              symbol.GetSort(),
		Remark:            symbol.GetRemark(),
		CreateTimes:       symbol.GetCreateTimes(),
		UpdateTimes:       symbol.GetUpdateTimes(),
	}
}

func tradeSymbolSpotToTypes(item *trade.TradeSymbolSpot) types.TradeSymbolSpot {
	return types.TradeSymbolSpot{
		Id:           item.GetId(),
		TenantId:     item.GetTenantId(),
		SymbolId:     item.GetSymbolId(),
		MakerFeeRate: item.GetMakerFeeRate(),
		TakerFeeRate: item.GetTakerFeeRate(),
		BuyEnabled:   int64(item.GetBuyEnabled()),
		SellEnabled:  int64(item.GetSellEnabled()),
		CreateTimes:  item.GetCreateTimes(),
		UpdateTimes:  item.GetUpdateTimes(),
	}
}

func tradeSymbolContractToTypes(item *trade.TradeSymbolContract) types.TradeSymbolContract {
	return types.TradeSymbolContract{
		Id:                     item.GetId(),
		TenantId:               item.GetTenantId(),
		SymbolId:               item.GetSymbolId(),
		ContractSize:           item.GetContractSize(),
		Multiplier:             item.GetMultiplier(),
		MaintenanceMarginRate:  item.GetMaintenanceMarginRate(),
		InitialMarginRate:      item.GetInitialMarginRate(),
		MakerFeeRate:           item.GetMakerFeeRate(),
		TakerFeeRate:           item.GetTakerFeeRate(),
		FundingIntervalMinutes: item.GetFundingIntervalMinutes(),
		DeliveryTime:           item.GetDeliveryTime(),
		SupportCross:           item.GetSupportCross(),
		SupportIsolated:        item.GetSupportIsolated(),
		FundingRateCap:         item.GetFundingRateCap(),
		FundingRateFloor:       item.GetFundingRateFloor(),
		IndexSymbol:            item.GetIndexSymbol(),
		MarkPriceSource:        item.GetMarkPriceSource(),
		SettlementPriceSource:  item.GetSettlementPriceSource(),
		OpenLongEnabled:        int64(item.GetOpenLongEnabled()),
		OpenShortEnabled:       int64(item.GetOpenShortEnabled()),
		CloseLongEnabled:       int64(item.GetCloseLongEnabled()),
		CloseShortEnabled:      int64(item.GetCloseShortEnabled()),
		CreateTimes:            item.GetCreateTimes(),
		UpdateTimes:            item.GetUpdateTimes(),
	}
}

func tradeSymbolLeverageConfigToTypes(item *trade.TradeSymbolLeverageConfig) types.TradeSymbolLeverageConfig {
	return types.TradeSymbolLeverageConfig{
		Id:          item.GetId(),
		TenantId:    item.GetTenantId(),
		SymbolId:    item.GetSymbolId(),
		MarginMode:  int64(item.GetMarginMode()),
		Leverage:    item.GetLeverage(),
		IsDefault:   int64(item.GetIsDefault()),
		Enabled:     int64(item.GetEnabled()),
		Sort:        item.GetSort(),
		Remark:      item.GetRemark(),
		CreateTimes: item.GetCreateTimes(),
		UpdateTimes: item.GetUpdateTimes(),
	}
}

func tradeSymbolSecondsToTypes(item *trade.TradeSymbolSeconds) types.TradeSymbolSeconds {
	return types.TradeSymbolSeconds{
		Id:                       item.GetId(),
		TenantId:                 item.GetTenantId(),
		SymbolId:                 item.GetSymbolId(),
		DurationSeconds:          item.GetDurationSeconds(),
		PayoutRate:               item.GetPayoutRate(),
		FeeRate:                  item.GetFeeRate(),
		DrawRule:                 int64(item.GetDrawRule()),
		StartPriceSource:         item.GetStartPriceSource(),
		SettlementPriceSource:    item.GetSettlementPriceSource(),
		QuoteValidityMs:          item.GetQuoteValidityMs(),
		SettlementWindowMs:       item.GetSettlementWindowMs(),
		SettlementPriceAlgorithm: item.GetSettlementPriceAlgorithm(),
		DrawTolerance:            item.GetDrawTolerance(),
		MaxExposureAmount:        item.GetMaxExposureAmount(),
		MinStake:                 item.GetMinStake(),
		MaxStake:                 item.GetMaxStake(),
		UpEnabled:                int64(item.GetUpEnabled()),
		DownEnabled:              int64(item.GetDownEnabled()),
		CreateTimes:              item.GetCreateTimes(),
		UpdateTimes:              item.GetUpdateTimes(),
	}
}
