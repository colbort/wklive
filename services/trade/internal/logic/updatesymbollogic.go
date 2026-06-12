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

type UpdateSymbolLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateSymbolLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateSymbolLogic {
	return &UpdateSymbolLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新交易对信息
func (l *UpdateSymbolLogic) UpdateSymbol(in *trade.UpdateSymbolReq) (*trade.AdminCommonResp, error) {
	item, err := l.svcCtx.TradeSymbolModel.FindOne(l.ctx, in.Id)
	if errors.Is(err, models.ErrNotFound) || (err == nil && item.TenantId != in.TenantId) {
		return &trade.AdminCommonResp{Base: helper.GetErrResp(i18n.BusinessDataNotFound, i18n.Translate(i18n.BusinessDataNotFound, l.ctx))}, nil
	}
	if err != nil {
		return nil, err
	}
	if in.DisplaySymbol != "" {
		item.DisplaySymbol = in.DisplaySymbol
	}
	if in.Status != trade.SymbolStatus_SYMBOL_STATUS_UNKNOWN {
		item.Status = int64(in.Status)
	}
	if in.PriceScale != 0 {
		item.PriceScale = int64(in.PriceScale)
	}
	if in.QtyScale != 0 {
		item.QtyScale = int64(in.QtyScale)
	}
	if in.MinPrice != "" {
		item.MinPrice = mustParseFloat(in.MinPrice)
	}
	if in.MaxPrice != "" {
		item.MaxPrice = mustParseFloat(in.MaxPrice)
	}
	if in.PriceTick != "" {
		item.PriceTick = mustParseFloat(in.PriceTick)
	}
	if in.MinQty != "" {
		item.MinQty = mustParseFloat(in.MinQty)
	}
	if in.MaxQty != "" {
		item.MaxQty = mustParseFloat(in.MaxQty)
	}
	if in.QtyStep != "" {
		item.QtyStep = mustParseFloat(in.QtyStep)
	}
	if in.MinNotional != "" {
		item.MinNotional = mustParseFloat(in.MinNotional)
	}
	if in.MaxLeverage != 0 {
		item.MaxLeverage = int64(in.MaxLeverage)
	}
	if in.OpenTime != 0 {
		item.OpenTime = in.OpenTime
	}
	if in.CloseTime != 0 {
		item.CloseTime = in.CloseTime
	}
	if in.Sort != 0 {
		item.Sort = int64(in.Sort)
	}
	if in.Remark != "" {
		item.Remark = in.Remark
	}
	item.UpdateTimes = utils.NowMillis()
	if err = l.svcCtx.TradeSymbolModel.Update(l.ctx, item); err != nil {
		return nil, err
	}

	return &trade.AdminCommonResp{Base: helper.OkResp()}, nil
}
