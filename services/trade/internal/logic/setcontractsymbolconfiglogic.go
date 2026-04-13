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

type SetContractSymbolConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetContractSymbolConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetContractSymbolConfigLogic {
	return &SetContractSymbolConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 设置合约交易对配置
func (l *SetContractSymbolConfigLogic) SetContractSymbolConfig(in *trade.SetContractSymbolConfigReq) (*trade.AdminCommonResp, error) {
	symbol, err := l.svcCtx.TradeSymbolModel.FindOne(l.ctx, in.SymbolId)
	if errors.Is(err, models.ErrNotFound) || (err == nil && symbol.TenantId != in.TenantId) {
		return &trade.AdminCommonResp{Base: helper.GetErrResp(404, i18n.Translate(i18n.BusinessDataNotFound, l.ctx))}, nil
	}
	if err != nil {
		return nil, err
	}
	now := utils.NowMillis()
	cfg, err := l.svcCtx.TradeSymbolContractModel.FindOneByTenantIdSymbolId(l.ctx, in.TenantId, in.SymbolId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if cfg == nil {
		cfg = &models.TTradeSymbolContract{
			TenantId:    in.TenantId,
			SymbolId:    in.SymbolId,
			CreateTimes: now,
		}
	}
	cfg.ContractSize = mustParseFloat(in.ContractSize)
	cfg.Multiplier = mustParseFloat(in.Multiplier)
	cfg.MaintenanceMarginRate = mustParseFloat(in.MaintenanceMarginRate)
	cfg.InitialMarginRate = mustParseFloat(in.InitialMarginRate)
	cfg.MakerFeeRate = mustParseFloat(in.MakerFeeRate)
	cfg.TakerFeeRate = mustParseFloat(in.TakerFeeRate)
	cfg.FundingIntervalMinutes = int64(in.FundingIntervalMinutes)
	cfg.DeliveryTime = in.DeliveryTime
	cfg.SupportCross = in.SupportCross
	cfg.SupportIsolated = in.SupportIsolated
	cfg.BuyEnabled = in.BuyEnabled
	cfg.SellEnabled = in.SellEnabled
	cfg.UpdateTimes = now
	if cfg.Id == 0 {
		if _, err = l.svcCtx.TradeSymbolContractModel.Insert(l.ctx, cfg); err != nil {
			return nil, err
		}
	} else if err = l.svcCtx.TradeSymbolContractModel.Update(l.ctx, cfg); err != nil {
		return nil, err
	}

	return &trade.AdminCommonResp{Base: helper.OkResp()}, nil
}
