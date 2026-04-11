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
		return &trade.AdminCommonResp{Base: helper.GetErrResp(404, i18n.Translate(i18n.BusinessDataNotFound, l.ctx))}, nil
	}
	if err != nil {
		return nil, err
	}
	item.DisplaySymbol = in.DisplaySymbol
	item.Status = int64(in.Status)
	item.PriceScale = int64(in.PriceScale)
	item.QtyScale = int64(in.QtyScale)
	item.MinPrice = mustParseFloat(in.MinPrice)
	item.MaxPrice = mustParseFloat(in.MaxPrice)
	item.PriceTick = mustParseFloat(in.PriceTick)
	item.MinQty = mustParseFloat(in.MinQty)
	item.MaxQty = mustParseFloat(in.MaxQty)
	item.QtyStep = mustParseFloat(in.QtyStep)
	item.MinNotional = mustParseFloat(in.MinNotional)
	item.MaxLeverage = int64(in.MaxLeverage)
	item.OpenTime = in.OpenTime
	item.CloseTime = in.CloseTime
	item.Sort = int64(in.Sort)
	item.Remark = in.Remark
	item.UpdateTimes = utils.NowMillis()
	if err = l.svcCtx.TradeSymbolModel.Update(l.ctx, item); err != nil {
		return nil, err
	}

	return &trade.AdminCommonResp{Base: helper.OkResp()}, nil
}
