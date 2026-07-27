package adminlogic

import (
	"context"
	"errors"
	helpers "wklive/services/trade/internal/logic/helpers"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/mr"
)

type GetSymbolDetailAdminLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetSymbolDetailAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSymbolDetailAdminLogic {
	return &GetSymbolDetailAdminLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取交易对详情
func (l *GetSymbolDetailAdminLogic) GetSymbolDetailAdmin(in *trade.GetSymbolDetailAdminReq) (*trade.GetSymbolDetailAdminResp, error) {
	item, err := l.svcCtx.TradeSymbolModel.FindOne(l.ctx, in.Id)
	if errors.Is(err, models.ErrNotFound) || (err == nil && item.TenantId != in.TenantId) {
		return &trade.GetSymbolDetailAdminResp{Base: helper.ErrResp(i18n.BusinessDataNotFound, i18n.Translate(i18n.BusinessDataNotFound, l.ctx))}, nil
	}
	if err != nil {
		return nil, err
	}

	resp := &trade.GetSymbolDetailAdminResp{
		Base: helper.OkResp(),
		Data: &trade.GetSymbolDetailAdminData{
			Symbol: helpers.SymbolToProto(item),
		},
	}
	var spot *models.TTradeSymbolSpot
	var contractCfg *models.TTradeSymbolContract
	var leverageConfigs []*models.TTradeSymbolLeverageConfig
	var secondsConfigs []*models.TTradeSymbolSeconds
	var sessions []*models.TTradeSymbolSession
	err = mr.Finish(
		func() error {
			var queryErr error
			spot, queryErr = l.svcCtx.TradeSymbolSpotModel.FindOneByTenantIdSymbolId(l.ctx, item.TenantId, item.Id)
			if errors.Is(queryErr, models.ErrNotFound) {
				return nil
			}
			return queryErr
		},
		func() error {
			var queryErr error
			contractCfg, queryErr = l.svcCtx.TradeSymbolContractModel.FindOneByTenantIdSymbolId(l.ctx, item.TenantId, item.Id)
			if errors.Is(queryErr, models.ErrNotFound) {
				return nil
			}
			return queryErr
		},
		func() error {
			var queryErr error
			leverageConfigs, _, queryErr = l.svcCtx.SymbolLeverageCfgModel.FindPage(l.ctx, models.TradeSymbolLeverageConfigPageFilter{
				TenantId: item.TenantId,
				SymbolId: item.Id,
			}, 0, 100)
			if errors.Is(queryErr, models.ErrNotFound) {
				return nil
			}
			return queryErr
		},
		func() error {
			var queryErr error
			secondsConfigs, queryErr = l.svcCtx.TradeSymbolSecondsModel.FindAllByTenantIdSymbolId(l.ctx, item.TenantId, item.Id)
			if errors.Is(queryErr, models.ErrNotFound) {
				return nil
			}
			return queryErr
		},
		func() error {
			var queryErr error
			sessions, queryErr = l.svcCtx.TradeSymbolSessionModel.FindAllByTenantIdSymbolId(l.ctx, item.TenantId, item.Id)
			if errors.Is(queryErr, models.ErrNotFound) {
				return nil
			}
			return queryErr
		},
	)
	if err != nil {
		return nil, err
	}

	resp.Data.Spot = helpers.SpotSymbolToProto(spot)
	resp.Data.Contract = helpers.ContractSymbolToProto(contractCfg)
	for _, cfg := range leverageConfigs {
		defaultLeverage, findErr := helpers.FindDefaultLeverage(l.ctx, l.svcCtx.SymbolLeverageDefaultModel, cfg.TenantId, cfg.SymbolId, cfg.MarginMode)
		if findErr != nil {
			return nil, findErr
		}
		resp.Data.LeverageConfigs = append(resp.Data.LeverageConfigs, helpers.SymbolLeverageConfigToProto(cfg, defaultLeverage))
	}
	for _, cfg := range secondsConfigs {
		resp.Data.SecondsConfigs = append(resp.Data.SecondsConfigs, helpers.SecondsSymbolToProto(cfg))
	}
	for _, session := range sessions {
		resp.Data.Sessions = append(resp.Data.Sessions, helpers.SymbolSessionToProto(session))
	}
	return resp, nil
}
