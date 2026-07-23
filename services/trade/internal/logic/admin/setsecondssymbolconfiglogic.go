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

type SetSecondsSymbolConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetSecondsSymbolConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetSecondsSymbolConfigLogic {
	return &SetSecondsSymbolConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 设置秒合约产品配置
func (l *SetSecondsSymbolConfigLogic) SetSecondsSymbolConfig(in *trade.SetSecondsSymbolConfigReq) (*trade.CommonResp, error) {
	symbol, err := l.svcCtx.TradeSymbolModel.FindOne(l.ctx, in.SymbolId)
	if errors.Is(err, models.ErrNotFound) || (err == nil && symbol.ProductType != int64(trade.ProductType_PRODUCT_TYPE_SECONDS)) {
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
	if err := validation.AuthoritativeQuoteSources("start_price_source", in.StartPriceSource); err != nil {
		return &trade.CommonResp{Base: helper.ErrResp(i18n.ParamError, err.Error())}, nil
	}
	if err := validation.AuthoritativeQuoteSources("settlement_price_source", in.SettlementPriceSource); err != nil {
		return &trade.CommonResp{Base: helper.ErrResp(i18n.ParamError, err.Error())}, nil
	}
	item, err := l.svcCtx.TradeSymbolSecondsModel.FindOneByTenantIdSymbolIdDurationSeconds(l.ctx, symbol.TenantId, in.SymbolId, in.DurationSeconds)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	now := utils.NowMillis()
	if item == nil {
		item = &models.TTradeSymbolSeconds{TenantId: symbol.TenantId, SymbolId: in.SymbolId, DurationSeconds: in.DurationSeconds, UpEnabled: int64(common.Enable_ENABLE_ENABLED), DownEnabled: int64(common.Enable_ENABLE_ENABLED), CreateTimes: now}
	}
	item.PayoutRate, item.DrawRule = mustParseFloat(in.PayoutRate), int64(in.DrawRule)
	item.FeeRate = mustParseFloat(in.FeeRate)
	item.StartPriceSource, item.SettlementPriceSource = in.StartPriceSource, in.SettlementPriceSource
	item.QuoteValidityMs = in.QuoteValidityMs
	item.SettlementWindowMs = in.SettlementWindowMs
	item.SettlementPriceAlgorithm = in.SettlementPriceAlgorithm
	item.DrawTolerance = mustParseFloat(in.DrawTolerance)
	item.MaxExposureAmount = mustParseFloat(in.MaxExposureAmount)
	item.MinStake, item.MaxStake = mustParseFloat(in.MinStake), mustParseFloat(in.MaxStake)
	item.UpEnabled = enableToModel(in.UpEnabled, item.UpEnabled)
	item.DownEnabled = enableToModel(in.DownEnabled, item.DownEnabled)
	item.UpdateTimes = now
	if item.Id == 0 {
		_, err = l.svcCtx.TradeSymbolSecondsModel.Insert(l.ctx, item)
	} else {
		err = l.svcCtx.TradeSymbolSecondsModel.Update(l.ctx, item)
	}
	if err != nil {
		return nil, err
	}
	return &trade.CommonResp{Base: helper.OkResp()}, nil
}
