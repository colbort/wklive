package adminlogic

import (
	"context"
	"errors"
	"wklive/proto/common"

	"wklive/common/conv"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/internal/validation"
	"wklive/services/trade/models"

	"github.com/shopspring/decimal"
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
func (l *UpdateSymbolLogic) UpdateSymbol(in *trade.UpdateSymbolReq) (*trade.CommonResp, error) {
	item, err := l.svcCtx.TradeSymbolModel.FindOne(l.ctx, in.Id)
	if errors.Is(err, models.ErrNotFound) {
		return &trade.CommonResp{Base: helper.ErrResp(i18n.BusinessDataNotFound, i18n.Translate(i18n.BusinessDataNotFound, l.ctx))}, nil
	}
	if err != nil {
		return nil, err
	}
	allowTenantUpdate, allowed, forbidden, err := utils.ResolveAdminTenantWriteScopeFromMd(l.ctx, item.TenantId)
	if err != nil {
		return nil, i18n.StatusError(l.ctx, i18n.UserNotFound)
	}
	if forbidden {
		return &trade.CommonResp{Base: helper.ErrResp(i18n.PermissionDenied, i18n.Translate(i18n.PermissionDenied, l.ctx))}, nil
	}
	if !allowed {
		return &trade.CommonResp{Base: helper.ErrResp(i18n.BusinessDataNotFound, i18n.Translate(i18n.BusinessDataNotFound, l.ctx))}, nil
	}
	if allowTenantUpdate {
		item.TenantId = in.TenantId
	}
	if in.CategoryType != 0 {
		if in.CategoryType < 1 || in.CategoryType > 6 {
			return &trade.CommonResp{Base: helper.ErrResp(i18n.ParamError, "invalid category type")}, nil
		}
		item.CategoryType = in.CategoryType
	}
	if item.CategoryType != 0 {
		if err := validation.SymbolCategoryConfig(item.CategoryType, common.ProductType(item.ProductType), common.ContractType(item.ContractType), trade.ContractValueType(item.ContractValueType)); err != nil {
			return &trade.CommonResp{Base: helper.ErrResp(i18n.ParamError, err.Error())}, nil
		}
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
	decimalFields := []struct {
		raw    string
		target *decimal.Decimal
	}{
		{in.MinPrice, &item.MinPrice},
		{in.MaxPrice, &item.MaxPrice},
		{in.PriceTick, &item.PriceTick},
		{in.MinQty, &item.MinQty},
		{in.MaxQty, &item.MaxQty},
		{in.QtyStep, &item.QtyStep},
		{in.MinNotional, &item.MinNotional},
		{in.MaxNotional, &item.MaxNotional},
	}
	for _, field := range decimalFields {
		if field.raw == "" {
			continue
		}
		value, parseErr := conv.ParseDecimalField(field.raw)
		if parseErr != nil {
			return &trade.CommonResp{Base: helper.ErrResp(i18n.ParamError, parseErr.Error())}, nil
		}
		*field.target = value
	}
	if in.ListingTime != 0 {
		item.ListingTime = in.ListingTime
	}
	if in.TradingStartTime != 0 {
		item.TradingStartTime = in.TradingStartTime
	}
	if in.TradingEndTime != 0 {
		item.TradingEndTime = in.TradingEndTime
	}
	if in.Sort != 0 {
		item.Sort = int64(in.Sort)
	}
	if in.Remark != "" {
		item.Remark = in.Remark
	}
	if err := validation.SymbolTradingTimeline(common.ProductType(item.ProductType), common.ContractType(item.ContractType), item.ListingTime, item.TradingStartTime, item.TradingEndTime); err != nil {
		return &trade.CommonResp{Base: helper.ErrResp(i18n.ParamError, err.Error())}, nil
	}
	item.UpdateTimes = utils.NowMillis()
	if err = l.svcCtx.TradeSymbolModel.Update(l.ctx, item); err != nil {
		return nil, err
	}

	return &trade.CommonResp{Base: helper.OkResp()}, nil
}
