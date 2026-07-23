package adminlogic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/trade"
	"wklive/services/trade/internal/authz"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/internal/validation"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetContractSymbolConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetContractSymbolConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetContractSymbolConfigLogic {
	return &SetContractSymbolConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 设置合约交易对配置
func (l *SetContractSymbolConfigLogic) SetContractSymbolConfig(in *trade.SetContractSymbolConfigReq) (*trade.CommonResp, error) {
	symbol, err := l.svcCtx.TradeSymbolModel.FindOne(l.ctx, in.SymbolId)
	if errors.Is(err, models.ErrNotFound) {
		return &trade.CommonResp{Base: helper.ErrResp(i18n.BusinessDataNotFound, i18n.Translate(i18n.BusinessDataNotFound, l.ctx))}, nil
	}
	if err != nil {
		return nil, err
	}
	if base, err := authz.AdminTenantWriteScopeResp(l.ctx, symbol.TenantId, i18n.BusinessDataNotFound); err != nil {
		return nil, err
	} else if base != nil {
		return &trade.CommonResp{Base: base}, nil
	}
	if err := validation.ContractTradingTimeline(symbol, in.DeliveryTime, in.OpenCutoffTime, in.MatchingStopTime); err != nil {
		return &trade.CommonResp{Base: helper.ErrResp(i18n.ParamError, err.Error())}, nil
	}
	if err := validation.ContractMarginModes(in.SupportCross, in.SupportIsolated); err != nil {
		return &trade.CommonResp{Base: helper.ErrResp(i18n.ParamError, err.Error())}, nil
	}
	now := utils.NowMillis()
	cfg, err := l.svcCtx.TradeSymbolContractModel.FindOneByTenantIdSymbolId(l.ctx, symbol.TenantId, in.SymbolId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if cfg == nil {
		cfg = &models.TTradeSymbolContract{
			TenantId:          symbol.TenantId,
			SymbolId:          in.SymbolId,
			OpenLongEnabled:   int64(common.Enable_ENABLE_ENABLED),
			OpenShortEnabled:  int64(common.Enable_ENABLE_ENABLED),
			CloseLongEnabled:  int64(common.Enable_ENABLE_ENABLED),
			CloseShortEnabled: int64(common.Enable_ENABLE_ENABLED),
			CreateTimes:       now,
		}
	}
	cfg.ContractSize = mustParseFloat(in.ContractSize)
	cfg.Multiplier = mustParseFloat(in.Multiplier)
	cfg.MaintenanceMarginRate = mustParseFloat(in.MaintenanceMarginRate)
	cfg.InitialMarginRate = mustParseFloat(in.InitialMarginRate)
	cfg.MakerFeeRate = mustParseFloat(in.MakerFeeRate)
	cfg.TakerFeeRate = mustParseFloat(in.TakerFeeRate)
	cfg.FundingIntervalMinutes = int64(in.FundingIntervalMinutes)
	cfg.FundingRateCap = mustParseFloat(in.FundingRateCap)
	cfg.FundingRateFloor = mustParseFloat(in.FundingRateFloor)
	cfg.FundingRateSource = in.FundingRateSource
	cfg.IndexSymbol = in.IndexSymbol
	cfg.MarkPriceSource = in.MarkPriceSource
	cfg.SettlementPriceSource = in.SettlementPriceSource
	cfg.DeliveryTime = in.DeliveryTime
	cfg.OpenCutoffTime = in.OpenCutoffTime
	cfg.MatchingStopTime = in.MatchingStopTime
	cfg.SettlementWindowSeconds = in.SettlementWindowSeconds
	cfg.SettlementPriceAlgorithm = in.SettlementPriceAlgorithm
	cfg.DeliveryFeeRate = mustParseFloat(in.DeliveryFeeRate)
	cfg.LiquidationFeeRate = mustParseFloat(in.LiquidationFeeRate)
	cfg.SupportCross = in.SupportCross
	cfg.SupportIsolated = in.SupportIsolated
	cfg.OpenLongEnabled = enableToModel(in.OpenLongEnabled, cfg.OpenLongEnabled)
	cfg.OpenShortEnabled = enableToModel(in.OpenShortEnabled, cfg.OpenShortEnabled)
	cfg.CloseLongEnabled = enableToModel(in.CloseLongEnabled, cfg.CloseLongEnabled)
	cfg.CloseShortEnabled = enableToModel(in.CloseShortEnabled, cfg.CloseShortEnabled)
	cfg.UpdateTimes = now
	if cfg.Id == 0 {
		if _, err = l.svcCtx.TradeSymbolContractModel.Insert(l.ctx, cfg); err != nil {
			return nil, err
		}
	} else if err = l.svcCtx.TradeSymbolContractModel.Update(l.ctx, cfg); err != nil {
		return nil, err
	}
	if in.SupportCross == 0 {
		if err := l.svcCtx.SymbolLeverageCfgModel.DisableGroup(l.ctx, symbol.TenantId, symbol.Id, int64(trade.MarginMode_MARGIN_MODE_CROSS), now); err != nil {
			return nil, err
		}
	}
	if in.SupportIsolated == 0 {
		if err := l.svcCtx.SymbolLeverageCfgModel.DisableGroup(l.ctx, symbol.TenantId, symbol.Id, int64(trade.MarginMode_MARGIN_MODE_ISOLATED), now); err != nil {
			return nil, err
		}
	}

	return &trade.CommonResp{Base: helper.OkResp()}, nil
}
