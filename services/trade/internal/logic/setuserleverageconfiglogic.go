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

type SetUserLeverageConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetUserLeverageConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetUserLeverageConfigLogic {
	return &SetUserLeverageConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 设置用户杠杆配置
func (l *SetUserLeverageConfigLogic) SetUserLeverageConfig(in *trade.SetUserLeverageConfigReq) (*trade.AdminCommonResp, error) {
	now := utils.NowMillis()
	item, err := l.svcCtx.ContractLeverageCfgModel.FindOneByTenantIdUserIdSymbolIdMarketTypeMarginMode(l.ctx, in.TenantId, in.UserId, in.SymbolId, int64(in.MarketType), int64(in.MarginMode))
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if item == nil {
		item = &models.TContractLeverageConfig{TenantId: in.TenantId, UserId: in.UserId, SymbolId: in.SymbolId, MarketType: int64(in.MarketType), MarginMode: int64(in.MarginMode), CreateTimes: now}
	}
	item.PositionMode = int64(in.PositionMode)
	item.LongLeverage = in.LongLeverage
	item.ShortLeverage = in.ShortLeverage
	item.MaxLeverage = in.MaxLeverage
	item.OperatorId = in.OperatorId
	item.Source = int64(in.Source)
	item.Status = in.Status
	item.Remark = in.Remark
	item.UpdateTimes = now
	if item.Id == 0 {
		_, err = l.svcCtx.ContractLeverageCfgModel.Insert(l.ctx, item)
	} else {
		err = l.svcCtx.ContractLeverageCfgModel.Update(l.ctx, item)
	}
	if err != nil {
		return nil, err
	}
	return &trade.AdminCommonResp{Base: helper.OkResp()}, nil
}
