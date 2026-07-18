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
	item, err := l.svcCtx.TradeSymbolModel.FindOne(l.ctx, in.SymbolId)
	if errors.Is(err, models.ErrNotFound) || (err == nil && item.TenantId != tenantId) {
		return &trade.GetSymbolDetailResp{Base: helper.ErrResp(i18n.BusinessDataNotFound, i18n.Translate(i18n.BusinessDataNotFound, l.ctx))}, nil
	}
	if err != nil {
		return nil, err
	}

	resp := &trade.GetSymbolDetailResp{
		Base: helper.OkResp(),
		Data: &trade.GetSymbolDetailData{
			Symbol: symbolToProto(item),
		},
	}
	var spot *models.TTradeSymbolSpot
	var contractCfg *models.TTradeSymbolContract
	var configs []*models.TTradeSymbolLeverageConfig
	err = mr.Finish(
		func() error {
			var queryErr error
			spot, queryErr = l.svcCtx.TradeSymbolSpotModel.FindOneByTenantIdSymbolId(l.ctx, tenantId, in.SymbolId)
			if errors.Is(queryErr, models.ErrNotFound) {
				return nil
			}
			return queryErr
		},
		func() error {
			var queryErr error
			contractCfg, queryErr = l.svcCtx.TradeSymbolContractModel.FindOneByTenantIdSymbolId(l.ctx, tenantId, in.SymbolId)
			if errors.Is(queryErr, models.ErrNotFound) {
				return nil
			}
			return queryErr
		},
		func() error {
			var queryErr error
			configs, _, queryErr = l.svcCtx.SymbolLeverageCfgModel.FindPage(l.ctx, models.TradeSymbolLeverageConfigPageFilter{
				TenantId: tenantId, SymbolId: in.SymbolId, MarketType: item.MarketType, Enabled: 1,
			}, 0, 100)
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
		resp.Data.LeverageConfigs = append(resp.Data.LeverageConfigs, symbolLeverageConfigToProto(cfg))
	}
	return resp, nil
}
