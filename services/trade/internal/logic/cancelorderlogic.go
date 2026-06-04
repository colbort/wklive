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
	"github.com/zeromicro/go-zero/core/stores/sqlx"
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
	userId, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	tenantId, err := utils.GetTenantIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	switch {
	case in.OrderId > 0:
		item, err = l.svcCtx.TradeOrderModel.FindOne(l.ctx, in.OrderId)
	case in.OrderNo != "":
		item, err = l.svcCtx.TradeOrderModel.FindOneByTenantIdOrderNo(l.ctx, tenantId, in.OrderNo)
	case in.ClientOrderId != "":
		item, err = l.svcCtx.TradeOrderModel.FindOneByTenantIdUserIdClientOrderId(l.ctx, tenantId, userId, in.ClientOrderId)
	default:
		return &trade.AppCommonResp{Base: helper.GetErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	if errors.Is(err, models.ErrNotFound) || (err == nil && (item.TenantId != tenantId || item.UserId != userId)) {
		return &trade.AppCommonResp{Base: helper.GetErrResp(i18n.OrderNotFound, i18n.Translate(i18n.OrderNotFound, l.ctx))}, nil
	}
	if err != nil {
		return nil, err
	}
	var canceledOrder *models.TTradeOrder
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		orderModel := models.NewTTradeOrderModel(conn, l.svcCtx.Config.CacheRedis).(models.TradeOrderModel)
		cancelLogModel := models.NewTTradeCancelLogModel(conn, l.svcCtx.Config.CacheRedis).(models.TradeCancelLogModel)

		locked, err := orderModel.FindOneForUpdate(ctx, item.Id)
		if err != nil {
			return err
		}
		if locked.TenantId != tenantId || locked.UserId != userId || !isOpenOrderStatus(locked.Status) {
			return nil
		}
		locked.Status = int64(trade.OrderStatus_ORDER_STATUS_CANCELED)
		locked.CancelReason = orderCancelReason("user")
		locked.UpdateTimes = utils.NowMillis()
		if err := orderModel.Update(ctx, locked); err != nil {
			return err
		}
		_, err = cancelLogModel.Insert(ctx, &models.TTradeCancelLog{
			TenantId:     locked.TenantId,
			OrderId:      locked.Id,
			OrderNo:      locked.OrderNo,
			UserId:       locked.UserId,
			CancelSource: int64(trade.CancelSource_CANCEL_SOURCE_USER),
			CancelReason: locked.CancelReason,
			CreateTimes:  utils.NowMillis(),
		})
		if err != nil {
			return err
		}
		canceledOrder = locked
		return nil
	})
	if err != nil {
		return nil, err
	}
	if canceledOrder != nil {
		if err = removeOrderBookOrder(l.svcCtx, l.ctx, canceledOrder); err != nil {
			return nil, err
		}
		if err = unfreezeRemainingOrderAsset(l.svcCtx, l.ctx, canceledOrder, "trade cancel order unfreeze"); err != nil {
			return nil, err
		}
	}

	return &trade.AppCommonResp{Base: helper.OkResp()}, nil
}
