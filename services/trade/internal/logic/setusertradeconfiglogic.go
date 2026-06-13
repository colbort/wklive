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

type SetUserTradeConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetUserTradeConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetUserTradeConfigLogic {
	return &SetUserTradeConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 设置用户交易配置
func (l *SetUserTradeConfigLogic) SetUserTradeConfig(in *trade.SetUserTradeConfigReq) (*trade.AdminCommonResp, error) {
	if base, err := adminTenantWriteScopeResp(l.ctx, in.TenantId, i18n.BusinessDataNotFound); err != nil {
		return nil, err
	} else if base != nil {
		return &trade.AdminCommonResp{Base: base}, nil
	}

	now := utils.NowMillis()
	item, err := l.svcCtx.TradeUserConfigModel.FindOneByTenantIdUserIdMarketTypeSymbolId(l.ctx, in.TenantId, in.UserId, int64(in.MarketType), in.SymbolId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if item == nil {
		item = &models.TTradeUserConfig{
			TenantId:          in.TenantId,
			UserId:            in.UserId,
			MarketType:        int64(in.MarketType),
			SymbolId:          in.SymbolId,
			TradeEnabled:      int64(common.Enable_ENABLE_ENABLED),
			ReduceOnlyEnabled: int64(common.Enable_ENABLE_DISABLED),
			CreateTimes:       now,
		}
	}
	item.PositionMode = int64(in.PositionMode)
	item.MarginMode = int64(in.MarginMode)
	item.DefaultLeverage = in.DefaultLeverage
	item.TradeEnabled = enableToModel(in.TradeEnabled, item.TradeEnabled)
	item.ReduceOnlyEnabled = enableToModel(in.ReduceOnlyEnabled, item.ReduceOnlyEnabled)
	item.UpdateTimes = now
	if item.Id == 0 {
		_, err = l.svcCtx.TradeUserConfigModel.Insert(l.ctx, item)
	} else {
		err = l.svcCtx.TradeUserConfigModel.Update(l.ctx, item)
	}
	if err != nil {
		return nil, err
	}
	return &trade.AdminCommonResp{Base: helper.OkResp()}, nil
}
