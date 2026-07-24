package applogic

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
	"github.com/zeromicro/go-zero/core/mr"
)

type GetSymbolDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetSymbolDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSymbolDetailLogic {
	return &GetSymbolDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取指定交易对详情
func (l *GetSymbolDetailLogic) GetSymbolDetail(in *trade.GetSymbolDetailReq) (*trade.GetSymbolDetailResp, error) {
	tenantId, err := utils.GetTenantIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	return QuerySymbolDetail(l.ctx, l.svcCtx, tenantId, in.SymbolId)
}

// QuerySymbolDetail is shared by the user-facing App service and trusted
// service-to-service callers. A zero tenantId performs a platform-level lookup.
func QuerySymbolDetail(ctx context.Context, svcCtx *svc.ServiceContext, tenantId, symbolId int64) (*trade.GetSymbolDetailResp, error) {
	item, err := svcCtx.TradeSymbolModel.FindOne(ctx, symbolId)
	if errors.Is(err, models.ErrNotFound) || (err == nil && tenantId > 0 && item.TenantId != tenantId) {
		return &trade.GetSymbolDetailResp{Base: helper.ErrResp(i18n.BusinessDataNotFound, i18n.Translate(i18n.BusinessDataNotFound, ctx))}, nil
	}
	if err != nil {
		return nil, err
	}
	configTenantId := item.TenantId

	resp := &trade.GetSymbolDetailResp{
		Base: helper.OkResp(),
		Data: &trade.GetSymbolDetailData{
			Symbol: symbolToProto(item),
		},
	}
	var spot *models.TTradeSymbolSpot
	var contractCfg *models.TTradeSymbolContract
	var configs []*models.TTradeSymbolLeverageConfig
	var secondsConfigs []*models.TTradeSymbolSeconds
	var sessions []*models.TTradeSymbolSession
	err = mr.Finish(
		func() error {
			var queryErr error
			spot, queryErr = svcCtx.TradeSymbolSpotModel.FindOneByTenantIdSymbolId(ctx, configTenantId, symbolId)
			if errors.Is(queryErr, models.ErrNotFound) {
				return nil
			}
			return queryErr
		},
		func() error {
			var queryErr error
			contractCfg, queryErr = svcCtx.TradeSymbolContractModel.FindOneByTenantIdSymbolId(ctx, configTenantId, symbolId)
			if errors.Is(queryErr, models.ErrNotFound) {
				return nil
			}
			return queryErr
		},
		func() error {
			var queryErr error
			configs, _, queryErr = svcCtx.SymbolLeverageCfgModel.FindPage(ctx, models.TradeSymbolLeverageConfigPageFilter{
				TenantId: configTenantId, SymbolId: symbolId, Enabled: 1,
			}, 0, 100)
			if errors.Is(queryErr, models.ErrNotFound) {
				return nil
			}
			return queryErr
		},
		func() error {
			var queryErr error
			secondsConfigs, queryErr = svcCtx.TradeSymbolSecondsModel.FindAllByTenantIdSymbolId(ctx, configTenantId, symbolId)
			if errors.Is(queryErr, models.ErrNotFound) {
				return nil
			}
			return queryErr
		},
		func() error {
			var queryErr error
			sessions, queryErr = svcCtx.TradeSymbolSessionModel.FindAllByTenantIdSymbolId(ctx, configTenantId, symbolId)
			if errors.Is(queryErr, models.ErrNotFound) {
				return nil
			}
			return queryErr
		},
	)
	if err != nil {
		return nil, err
	}
	if spot != nil {
		resp.Data.Spot = spotSymbolToProto(spot)
	}
	if contractCfg != nil {
		resp.Data.Contract = contractSymbolToProto(contractCfg)
	}
	for _, cfg := range configs {
		defaultLeverage, findErr := findDefaultLeverage(ctx, svcCtx.SymbolLeverageDefaultModel, cfg.TenantId, cfg.SymbolId, cfg.MarginMode)
		if findErr != nil {
			return nil, findErr
		}
		resp.Data.LeverageConfigs = append(resp.Data.LeverageConfigs, symbolLeverageConfigToProto(cfg, defaultLeverage))
	}
	for _, cfg := range secondsConfigs {
		resp.Data.SecondsConfigs = append(resp.Data.SecondsConfigs, secondsSymbolToProto(cfg))
	}
	for _, session := range sessions {
		resp.Data.Sessions = append(resp.Data.Sessions, symbolSessionToProto(session))
	}
	return resp, nil
}
