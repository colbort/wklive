package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
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
		return &trade.GetSymbolDetailAdminResp{Base: helper.GetErrResp(i18n.BusinessDataNotFound, i18n.Translate(i18n.BusinessDataNotFound, l.ctx))}, nil
	}
	if err != nil {
		return nil, err
	}

	resp := &trade.GetSymbolDetailAdminResp{
		Base: helper.OkResp(),
		Data: &trade.GetSymbolDetailAdminData{
			Data: symbolToProto(item),
		},
	}
	configs, _, err := l.svcCtx.SymbolLeverageCfgModel.FindPage(l.ctx, models.TradeSymbolLeverageConfigPageFilter{
		TenantId:   in.TenantId,
		SymbolId:   in.Id,
		MarketType: item.MarketType,
	}, 0, 100)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	for _, cfg := range configs {
		resp.Data.LeverageConfigs = append(resp.Data.LeverageConfigs, symbolLeverageConfigToProto(cfg))
	}
	return resp, nil
}
