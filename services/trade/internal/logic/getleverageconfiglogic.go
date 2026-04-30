package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/common/utils"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetLeverageConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetLeverageConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLeverageConfigLogic {
	return &GetLeverageConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取当前杠杆配置
func (l *GetLeverageConfigLogic) GetLeverageConfig(in *trade.GetLeverageConfigReq) (*trade.GetLeverageConfigResp, error) {
	userId, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	tenantId, err := utils.GetTenantIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	cfg, err := l.svcCtx.ContractLeverageCfgModel.FindOneByTenantIdUserIdSymbolIdMarketTypeMarginMode(
		l.ctx, tenantId, userId, in.SymbolId, int64(in.MarketType), int64(in.MarginMode),
	)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	resp := &trade.GetLeverageConfigResp{Base: helper.OkResp()}
	if cfg != nil {
		resp.Data = leverageConfigToProto(cfg)
	}

	return resp, nil
}
