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
func (l *SetUserTradeConfigLogic) SetUserTradeConfig(in *trade.SetUserTradeConfigReq) (*trade.CommonResp, error) {
	if base, err := authz.AdminTenantWriteScopeResp(l.ctx, in.TenantId, i18n.BusinessDataNotFound); err != nil {
		return nil, err
	} else if base != nil {
		return &trade.CommonResp{Base: base}, nil
	}

	now := utils.NowMillis()
	item, err := l.svcCtx.TradeUserConfigModel.FindOneByTenantIdUserIdProductTypeSymbolId(l.ctx, in.TenantId, in.UserId, int64(in.ProductType), in.SymbolId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if item == nil {
		item = &models.TTradeUserConfig{
			TenantId:     in.TenantId,
			UserId:       in.UserId,
			ProductType:  int64(in.ProductType),
			SymbolId:     in.SymbolId,
			TradeEnabled: int64(common.Enable_ENABLE_ENABLED),
			CreateTimes:  now,
		}
	}
	item.TradeEnabled = enableToModel(in.TradeEnabled, item.TradeEnabled)
	item.UpdateTimes = now
	if item.Id == 0 {
		_, err = l.svcCtx.TradeUserConfigModel.Insert(l.ctx, item)
	} else {
		err = l.svcCtx.TradeUserConfigModel.Update(l.ctx, item)
	}
	if err != nil {
		return nil, err
	}
	return &trade.CommonResp{Base: helper.OkResp()}, nil
}
