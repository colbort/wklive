package logic

import (
	"context"
	"errors"

	"wklive/common/conv"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type CancelOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCancelOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelOrderLogic {
	return &CancelOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 撤销指定订单
func (l *CancelOrderLogic) CancelOrder(in *trade.CancelOrderReq) (*trade.AppCommonResp, error) {
	var (
		item *models.TTradeOrder
		err  error
	)
	switch {
	case in.OrderId > 0:
		item, err = l.svcCtx.TradeOrderModel.FindOne(l.ctx, in.OrderId)
	case in.OrderNo != "":
		item, err = l.svcCtx.TradeOrderModel.FindOneByTenantIdOrderNo(l.ctx, in.TenantId, in.OrderNo)
	case in.ClientOrderId != "":
		item, err = l.svcCtx.TradeOrderModel.FindOneByTenantIdUserIdClientOrderId(l.ctx, in.TenantId, in.UserId, in.ClientOrderId)
	default:
		return &trade.AppCommonResp{Base: helper.GetErrResp(400, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	if errors.Is(err, models.ErrNotFound) || (err == nil && (item.TenantId != in.TenantId || item.UserId != in.UserId)) {
		return &trade.AppCommonResp{Base: helper.GetErrResp(404, i18n.Translate(i18n.OrderNotFound, l.ctx))}, nil
	}
	if err != nil {
		return nil, err
	}
	if item.Status != int64(trade.OrderStatus_ORDER_STATUS_PENDING) && item.Status != int64(trade.OrderStatus_ORDER_STATUS_PART_FILLED) {
		return &trade.AppCommonResp{Base: helper.OkResp()}, nil
	}

	ext, err := parseOrderAssetExt(conv.NullStringValue(item.BizExt))
	if err != nil {
		return nil, err
	}

	if ext.FreezeNo != "" {
		var unfreezeAmount float64
		if item.MarketType == int64(trade.MarketType_MARKET_TYPE_SPOT) {
			spot, findErr := l.svcCtx.TradeOrderSpotModel.FindOneByTenantIdOrderId(l.ctx, item.TenantId, item.Id)
			if findErr != nil && !errors.Is(findErr, models.ErrNotFound) {
				return nil, findErr
			}
			if spot != nil {
				unfreezeAmount = spot.FrozenAmount
			}
		} else {
			contract, findErr := l.svcCtx.TradeOrderContractModel.FindOneByTenantIdOrderId(l.ctx, item.TenantId, item.Id)
			if findErr != nil && !errors.Is(findErr, models.ErrNotFound) {
				return nil, findErr
			}
			if contract != nil {
				unfreezeAmount = contract.MarginAmount
			}
		}
		if err = unfreezeOrderAsset(l.svcCtx, l.ctx, item, ext.FreezeNo, unfreezeAmount, "trade cancel order unfreeze"); err != nil {
			return nil, err
		}
	}

	item.Status = int64(trade.OrderStatus_ORDER_STATUS_CANCELED)
	item.CancelReason = orderCancelReason("user")
	item.UpdateTimes = utils.NowMillis()
	if err = l.svcCtx.TradeOrderModel.Update(l.ctx, item); err != nil {
		return nil, err
	}
	_, err = l.svcCtx.TradeCancelLogModel.Insert(l.ctx, &models.TTradeCancelLog{
		TenantId:     item.TenantId,
		OrderId:      item.Id,
		OrderNo:      item.OrderNo,
		UserId:       item.UserId,
		CancelSource: int64(trade.CancelSource_CANCEL_SOURCE_USER),
		CancelReason: item.CancelReason,
		CreateTimes:  utils.NowMillis(),
	})
	if err != nil {
		return nil, err
	}

	return &trade.AppCommonResp{Base: helper.OkResp()}, nil
}
