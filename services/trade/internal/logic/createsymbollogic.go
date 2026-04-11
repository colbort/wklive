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

type CreateSymbolLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateSymbolLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateSymbolLogic {
	return &CreateSymbolLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建交易对
func (l *CreateSymbolLogic) CreateSymbol(in *trade.CreateSymbolReq) (*trade.AdminCommonResp, error) {
	exists, err := l.svcCtx.TradeSymbolModel.FindOneByTenantIdSymbolMarketType(l.ctx, in.TenantId, in.Symbol, int64(in.MarketType))
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if exists != nil {
		return &trade.AdminCommonResp{Base: helper.GetErrResp(400, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	now := utils.NowMillis()
	data := &models.TTradeSymbol{
		TenantId:      in.TenantId,
		Symbol:        in.Symbol,
		DisplaySymbol: in.DisplaySymbol,
		MarketType:    int64(in.MarketType),
		BaseAsset:     in.BaseAsset,
		QuoteAsset:    in.QuoteAsset,
		SettleAsset:   in.SettleAsset,
		ContractType:  int64(in.ContractType),
		Status:        int64(in.Status),
		PriceScale:    int64(in.PriceScale),
		QtyScale:      int64(in.QtyScale),
		MinPrice:      mustParseFloat(in.MinPrice),
		MaxPrice:      mustParseFloat(in.MaxPrice),
		PriceTick:     mustParseFloat(in.PriceTick),
		MinQty:        mustParseFloat(in.MinQty),
		MaxQty:        mustParseFloat(in.MaxQty),
		QtyStep:       mustParseFloat(in.QtyStep),
		MinNotional:   mustParseFloat(in.MinNotional),
		MaxLeverage:   int64(in.MaxLeverage),
		OpenTime:      in.OpenTime,
		CloseTime:     in.CloseTime,
		Sort:          int64(in.Sort),
		Remark:        in.Remark,
		CreateTimes:   now,
		UpdateTimes:   now,
	}
	if _, err = l.svcCtx.TradeSymbolModel.Insert(l.ctx, data); err != nil {
		return nil, err
	}

	return &trade.AdminCommonResp{Base: helper.OkResp()}, nil
}
