package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserLeverageConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserLeverageConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserLeverageConfigLogic {
	return &GetUserLeverageConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取用户杠杆配置
func (l *GetUserLeverageConfigLogic) GetUserLeverageConfig(in *trade.GetUserLeverageConfigReq) (*trade.GetUserLeverageConfigResp, error) {
	item, err := l.svcCtx.ContractLeverageCfgModel.FindOneByTenantIdUserIdSymbolIdMarketTypeMarginMode(l.ctx, in.TenantId, in.UserId, in.SymbolId, int64(in.MarketType), int64(in.MarginMode))
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	resp := &trade.GetUserLeverageConfigResp{Base: helper.OkResp()}
	if item != nil {
		resp.Data = leverageConfigToProto(item)
	}
	return resp, nil
}
