package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetSymbolLeverageConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetSymbolLeverageConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetSymbolLeverageConfigLogic {
	return &SetSymbolLeverageConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 设置交易对杠杆档位配置
func (l *SetSymbolLeverageConfigLogic) SetSymbolLeverageConfig(in *trade.SetSymbolLeverageConfigReq) (*trade.AdminCommonResp, error) {
	symbol, err := l.svcCtx.TradeSymbolModel.FindOne(l.ctx, in.SymbolId)
	if errors.Is(err, models.ErrNotFound) {
		return &trade.AdminCommonResp{Base: helper.GetErrResp(i18n.BusinessDataNotFound, i18n.Translate(i18n.BusinessDataNotFound, l.ctx))}, nil
	}
	if err != nil {
		return nil, err
	}
	if base, err := adminTenantWriteScopeResp(l.ctx, symbol.TenantId, i18n.BusinessDataNotFound); err != nil {
		return nil, err
	} else if base != nil {
		return &trade.AdminCommonResp{Base: base}, nil
	}
	if !isContractMarket(in.MarketType) || symbol.MarketType != int64(in.MarketType) || in.MarginMode == trade.MarginMode_MARGIN_MODE_UNKNOWN {
		return &trade.AdminCommonResp{Base: helper.GetErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}

	now := utils.NowMillis()
	item, err := l.svcCtx.SymbolLeverageCfgModel.FindOneByTenantIdSymbolIdMarketTypeMarginMode(l.ctx, symbol.TenantId, in.SymbolId, int64(in.MarketType), int64(in.MarginMode))
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	isCreate := item == nil
	if isCreate {
		item = &models.TTradeSymbolLeverageConfig{
			TenantId:    symbol.TenantId,
			SymbolId:    in.SymbolId,
			MarketType:  int64(in.MarketType),
			MarginMode:  int64(in.MarginMode),
			Enabled:     int64(common.Enable_ENABLE_ENABLED),
			Sort:        int64(in.MarginMode),
			CreateTimes: now,
		}
	}

	if isCreate || in.MaxLeverage > 0 || len(in.LeverageValues) > 0 || in.DefaultLeverage > 0 {
		maxLeverage := item.MaxLeverage
		if in.MaxLeverage > 0 {
			maxLeverage = in.MaxLeverage
		}

		leverageSource := parseLeverageValues(item.LeverageValues)
		if len(in.LeverageValues) > 0 {
			leverageSource = in.LeverageValues
		}
		if len(leverageSource) == 0 && in.DefaultLeverage > 0 {
			leverageSource = []int64{in.DefaultLeverage}
		}
		if valueMax := maxLeverageValue(leverageSource); valueMax > maxLeverage {
			maxLeverage = valueMax
		}

		if maxLeverage <= 0 {
			maxLeverage = symbol.MaxLeverage
		}
		if maxLeverage <= 0 {
			maxLeverage = 1
		}

		leverageText, leverageValues := joinLeverageValues(leverageSource, maxLeverage)
		defaultLeverage := item.DefaultLeverage
		if in.DefaultLeverage > 0 {
			defaultLeverage = in.DefaultLeverage
		}
		if !containsLeverage(leverageValues, defaultLeverage) {
			defaultLeverage = leverageValues[0]
		}

		item.LeverageValues = leverageText
		item.DefaultLeverage = defaultLeverage
		item.MaxLeverage = maxLeverage
	}

	if in.Enabled > 0 {
		item.Enabled = enableToModel(in.Enabled, item.Enabled)
	}
	if in.Sort > 0 {
		item.Sort = in.Sort
	}
	if in.Remark != "" {
		item.Remark = in.Remark
	}
	item.UpdateTimes = now

	if item.Id == 0 {
		if _, err = l.svcCtx.SymbolLeverageCfgModel.Insert(l.ctx, item); err != nil {
			return nil, err
		}
	} else if err = l.svcCtx.SymbolLeverageCfgModel.Update(l.ctx, item); err != nil {
		return nil, err
	}

	return &trade.AdminCommonResp{Base: helper.OkResp()}, nil
}

func maxLeverageValue(values []int64) int64 {
	var maxValue int64
	for _, value := range values {
		if value > maxValue {
			maxValue = value
		}
	}
	return maxValue
}
