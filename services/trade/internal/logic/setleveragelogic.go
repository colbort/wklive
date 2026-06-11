package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetLeverageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetLeverageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetLeverageLogic {
	return &SetLeverageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 设置杠杆倍数
func (l *SetLeverageLogic) SetLeverage(in *trade.SetLeverageReq) (*trade.AppCommonResp, error) {
	userId, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	tenantId, err := utils.GetTenantIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	symbol, err := l.svcCtx.TradeSymbolModel.FindOne(l.ctx, in.SymbolId)
	if errors.Is(err, models.ErrNotFound) || (err == nil && symbol.TenantId != tenantId) {
		return &trade.AppCommonResp{Base: helper.GetErrResp(i18n.BusinessDataNotFound, i18n.Translate(i18n.BusinessDataNotFound, l.ctx))}, nil
	}
	if err != nil {
		return nil, err
	}
	longLeverage, ok, err := ensureConfiguredLeverage(l.ctx, l.svcCtx.SymbolLeverageCfgModel, tenantId, symbol, in.MarginMode, in.LongLeverage)
	if err != nil {
		return nil, err
	}
	if !ok {
		return &trade.AppCommonResp{Base: helper.GetErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	shortLeverage, ok, err := ensureConfiguredLeverage(l.ctx, l.svcCtx.SymbolLeverageCfgModel, tenantId, symbol, in.MarginMode, in.ShortLeverage)
	if err != nil {
		return nil, err
	}
	if !ok {
		return &trade.AppCommonResp{Base: helper.GetErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	now := utils.NowMillis()
	cfg, err := l.svcCtx.ContractLeverageCfgModel.FindOneByTenantIdUserIdSymbolIdMarketTypeMarginMode(
		l.ctx, tenantId, userId, in.SymbolId, int64(in.MarketType), int64(in.MarginMode),
	)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if cfg == nil {
		cfg = &models.TContractLeverageConfig{
			TenantId:     tenantId,
			UserId:       userId,
			SymbolId:     in.SymbolId,
			MarketType:   int64(in.MarketType),
			MarginMode:   int64(in.MarginMode),
			CreateTimes:  now,
			Source:       int64(trade.SourceType_SOURCE_TYPE_USER),
			Enabled:      1,
			PositionMode: int64(in.PositionMode),
		}
	}
	cfg.PositionMode = int64(in.PositionMode)
	cfg.LongLeverage = longLeverage
	cfg.ShortLeverage = shortLeverage
	cfg.MaxLeverage = maxLeverageValue([]int64{symbol.MaxLeverage, longLeverage, shortLeverage})
	cfg.OperatorId = userId
	cfg.UpdateTimes = now
	if cfg.Id == 0 {
		if _, err = l.svcCtx.ContractLeverageCfgModel.Insert(l.ctx, cfg); err != nil {
			return nil, err
		}
	} else if err = l.svcCtx.ContractLeverageCfgModel.Update(l.ctx, cfg); err != nil {
		return nil, err
	}

	return &trade.AppCommonResp{Base: helper.OkResp()}, nil
}
