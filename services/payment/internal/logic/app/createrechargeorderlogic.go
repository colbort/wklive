package applogic

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"wklive/common/generate"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/notify"
	"wklive/common/utils"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"
	"wklive/services/payment/models"
)

const (
	defaultRechargeSessionTTL = 15 * time.Minute
	rechargeSessionLockTTL    = 30
)

type CreateRechargeOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateRechargeOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateRechargeOrderLogic {
	return &CreateRechargeOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建充值订单
func (l *CreateRechargeOrderLogic) CreateRechargeOrder(in *payment.CreateRechargeOrderReq) (*payment.CreateRechargeOrderResp, error) {
	userId, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	tenantId, err := utils.GetTenantIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	clientIP, _ := utils.GetClientIPFromMd(l.ctx)
	now := utils.NowMillis()

	// 查询通道信息
	channel, err := l.svcCtx.TenantPayChannelModel.FindOne(l.ctx, in.ChannelId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if channel == nil {
		return &payment.CreateRechargeOrderResp{
			Base: helper.ErrResp(i18n.PaymentChannelNotFound, i18n.Translate(i18n.PaymentChannelNotFound, l.ctx)),
		}, nil
	}

	// 验证通道可用性
	if channel.Enabled != 1 {
		return &payment.CreateRechargeOrderResp{
			Base: helper.ErrResp(i18n.PaymentChannelUnavailable, i18n.Translate(i18n.PaymentChannelUnavailable, l.ctx)),
		}, nil
	}
	platform, err := l.svcCtx.PayPlatformModel.FindOne(l.ctx, channel.PlatformId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	rechargeType := rechargeTypeFromPlatform(platform)

	if platform != nil && platform.PlatformType == int64(payment.PlatformType_PLATFORM_TYPE_THIRD) {
		reusable, err := l.findReusableRechargeOrder(tenantId, userId, channel.Id, in.RechargeAmount, in.Currency)
		if err != nil {
			return nil, err
		}
		if reusable != nil {
			return &payment.CreateRechargeOrderResp{Base: helper.OkResp(), Data: toRechargeOrderProto(reusable)}, nil
		}
		release, acquired, err := l.acquireRechargeSession(tenantId, userId, channel.Id, in.RechargeAmount, in.Currency)
		if err != nil {
			return nil, err
		}
		if !acquired {
			return nil, fmt.Errorf("payment session is being created, please retry")
		}
		defer release()
		// The first request may have completed between the initial lookup and lock acquisition.
		reusable, err = l.findReusableRechargeOrder(tenantId, userId, channel.Id, in.RechargeAmount, in.Currency)
		if err != nil {
			return nil, err
		}
		if reusable != nil {
			return &payment.CreateRechargeOrderResp{Base: helper.OkResp(), Data: toRechargeOrderProto(reusable)}, nil
		}
	}

	// 验证金额限制
	if in.RechargeAmount < channel.SingleMinAmount ||
		(channel.SingleMaxAmount > 0 && in.RechargeAmount > channel.SingleMaxAmount) {
		return &payment.CreateRechargeOrderResp{
			Base: helper.ErrResp(i18n.RechargeAmountOutOfLimit, i18n.Translate(i18n.RechargeAmountOutOfLimit, l.ctx)),
		}, nil
	}

	// 仅在确认不存在可复用支付会话后生成新的本地订单号。
	orderNo, err := generate.GenerateNo(l.svcCtx.Redis, l.ctx, "order_id", "RC", "")
	if err != nil {
		return nil, err
	}

	// 计算手续费
	var feeAmount int64
	if channel.FeeType == int64(payment.FeeType_FEE_TYPE_RATE) {
		// 按比例计算
		feeAmount = decimal.NewFromInt(in.RechargeAmount).Mul(channel.FeeRate).Div(decimal.NewFromInt(100)).IntPart()
	} else if channel.FeeType == int64(payment.FeeType_FEE_TYPE_FIXED) {
		// 固定费用
		feeAmount = channel.FeeFixedAmount
	}

	// 创建充值订单
	rechargeOrder := &models.TRechargeOrder{
		TenantId:     tenantId,
		UserId:       userId,
		OrderNo:      orderNo,
		BizOrderNo:   sql.NullString{String: in.BizOrderNo, Valid: in.BizOrderNo != ""},
		PlatformId:   channel.PlatformId,
		ProductId:    channel.ProductId,
		AccountId:    channel.AccountId,
		ChannelId:    in.ChannelId,
		RechargeType: int64(rechargeType),
		WalletType:   1,
		Currency:     in.Currency,
		OrderAmount:  in.RechargeAmount,
		PayAmount:    in.RechargeAmount,
		FeeAmount:    feeAmount,
		Subject:      sql.NullString{String: in.Subject, Valid: in.Subject != ""},
		Body:         sql.NullString{String: in.Body, Valid: in.Body != ""},
		ClientType:   int64(in.ClientType),
		ClientIp:     sql.NullString{String: clientIP, Valid: clientIP != ""},
		VoucherImage: "",
		Status:       int64(payment.PayOrderStatus_PAY_ORDER_STATUS_PENDING),
		CreateTimes:  now,
		UpdateTimes:  now,
	}

	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		rechargeOrderModel := models.NewTRechargeOrderModel(conn, l.svcCtx.Config.CacheRedis)
		userRechargeStatModel := models.NewTUserRechargeStatModel(conn, l.svcCtx.Config.CacheRedis)

		result, err := rechargeOrderModel.Insert(ctx, rechargeOrder)
		if err != nil {
			return err
		}
		rechargeOrder.Id, err = result.LastInsertId()
		if err != nil {
			return err
		}

		stat, err := userRechargeStatModel.FindOneByTenantIdUserId(ctx, tenantId, userId)
		if err != nil && !errors.Is(err, models.ErrNotFound) {
			return err
		}
		if stat == nil {
			_, err = userRechargeStatModel.Insert(ctx, &models.TUserRechargeStat{
				TenantId:    tenantId,
				UserId:      userId,
				CreateTimes: now,
				UpdateTimes: now,
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 外部网络调用在本地事务提交之后执行，并由支付编排层保证只提交一次。
	if err := l.createThirdPartyPayment(platform, rechargeOrder); err != nil {
		return nil, err
	}
	if err := l.cacheRechargeSession(rechargeOrder); err != nil {
		l.Errorf("cache recharge payment session failed, orderNo=%s err=%v", rechargeOrder.OrderNo, err)
	}

	l.Logger.Infof("Create recharge order success: %s, user_id: %d", orderNo, userId)
	event := notify.NewEvent(notify.EventTypeRecharge, notify.EventLevelInfo, "用户充值", fmt.Sprintf("用户 %d 发起充值订单 %s", userId, orderNo))
	event.Source = "payment"
	event.TenantID = tenantId
	event.UserID = userId
	event.BizNo = orderNo
	event.Data = map[string]any{
		"amount":   in.RechargeAmount,
		"currency": in.Currency,
	}
	if err := notify.Publish(l.ctx, l.svcCtx.MQPublisher, event); err != nil {
		l.Errorf("publish admin recharge notification failed, orderNo=%s err=%v", orderNo, err)
	}

	return &payment.CreateRechargeOrderResp{
		Base: helper.OkResp(),
		Data: toRechargeOrderProto(rechargeOrder),
	}, nil
}

func rechargeSessionKey(tenantID, userID, channelID, amount int64, currency string) string {
	return fmt.Sprintf("payment:recharge:session:%d:%d:%d:%s:%d",
		tenantID, userID, channelID, strings.ToUpper(strings.TrimSpace(currency)), amount)
}

func (l *CreateRechargeOrderLogic) findReusableRechargeOrder(
	tenantID, userID, channelID, amount int64,
	currency string,
) (*models.TRechargeOrder, error) {
	key := rechargeSessionKey(tenantID, userID, channelID, amount, currency)
	orderNo, err := l.svcCtx.Redis.GetCtx(l.ctx, key)
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(orderNo) == "" {
		return nil, nil
	}
	order, err := l.svcCtx.RechargeOrderModel.FindOneByOrderNo(l.ctx, orderNo)
	if err != nil {
		_, _ = l.svcCtx.Redis.DelCtx(l.ctx, key)
		if errors.Is(err, models.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	now := time.Now().UnixMilli()
	reusable := order.TenantId == tenantID &&
		order.UserId == userID &&
		order.ChannelId == channelID &&
		order.OrderAmount == amount &&
		strings.EqualFold(order.Currency, currency) &&
		(order.Status == int64(payment.PayOrderStatus_PAY_ORDER_STATUS_PENDING) ||
			order.Status == int64(payment.PayOrderStatus_PAY_ORDER_STATUS_PAYING)) &&
		(order.ExpireTime == 0 || order.ExpireTime > now) &&
		(order.PayUrl.Valid || order.QrContent.Valid)
	if !reusable {
		_, _ = l.svcCtx.Redis.DelCtx(l.ctx, key)
		return nil, nil
	}
	return order, nil
}

func (l *CreateRechargeOrderLogic) acquireRechargeSession(
	tenantID, userID, channelID, amount int64,
	currency string,
) (func(), bool, error) {
	key := rechargeSessionKey(tenantID, userID, channelID, amount, currency) + ":lock"
	token := strconv.FormatInt(time.Now().UnixNano(), 10)
	ok, err := l.svcCtx.Redis.SetnxExCtx(l.ctx, key, token, rechargeSessionLockTTL)
	if err != nil {
		return nil, false, err
	}
	release := func() {
		current, getErr := l.svcCtx.Redis.GetCtx(context.Background(), key)
		if getErr == nil && current == token {
			_, _ = l.svcCtx.Redis.DelCtx(context.Background(), key)
		}
	}
	return release, ok, nil
}

func (l *CreateRechargeOrderLogic) cacheRechargeSession(order *models.TRechargeOrder) error {
	if order == nil || (!order.PayUrl.Valid && !order.QrContent.Valid) {
		return nil
	}
	ttl := defaultRechargeSessionTTL
	if order.ExpireTime > 0 {
		ttl = time.Until(time.UnixMilli(order.ExpireTime))
	}
	if ttl <= 0 {
		return nil
	}
	key := rechargeSessionKey(order.TenantId, order.UserId, order.ChannelId, order.OrderAmount, order.Currency)
	return l.svcCtx.Redis.SetexCtx(l.ctx, key, order.OrderNo, int(ttl.Seconds()))
}

func (l *CreateRechargeOrderLogic) createThirdPartyPayment(
	platform *models.TPayPlatform,
	order *models.TRechargeOrder,
) error {
	if platform == nil || order == nil {
		return fmt.Errorf("payment platform and recharge order are required")
	}
	if platform.PlatformType != int64(payment.PlatformType_PLATFORM_TYPE_THIRD) {
		return nil
	}
	if order.ThirdOrderNo.Valid && order.ThirdOrderNo.String != "" {
		return nil
	}

	requestNo := order.OrderNo + "_CREATE"
	existing, err := l.svcCtx.PayRequestLogModel.FindOneByRequestNo(l.ctx, requestNo)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return err
	}
	if existing != nil {
		switch payment.PayRequestStatus(existing.Status) {
		case payment.PayRequestStatus_PAY_REQUEST_STATUS_SUCCESS,
			payment.PayRequestStatus_PAY_REQUEST_STATUS_PROCESSING,
			payment.PayRequestStatus_PAY_REQUEST_STATUS_UNCERTAIN:
			if current, findErr := l.svcCtx.RechargeOrderModel.FindOne(l.ctx, order.Id); findErr == nil {
				*order = *current
			}
			return nil
		}
	}

	account, err := l.svcCtx.TenantPayAccountModel.FindOne(l.ctx, order.AccountId)
	if err != nil {
		return err
	}
	adapter, err := l.svcCtx.PaymentAdapters.Get(platform.PlatformCode)
	if err != nil {
		return err
	}

	startedAt := utils.NowMillis()
	requestLog := &models.TPayRequestLog{
		TenantId: order.TenantId, OrderType: 1, OrderId: order.Id,
		OrderNo: order.OrderNo, PlatformId: order.PlatformId, AccountId: order.AccountId,
		RequestType: 1, RequestNo: requestNo,
		Status:      int64(payment.PayRequestStatus_PAY_REQUEST_STATUS_PROCESSING),
		CreateTimes: startedAt, UpdateTimes: startedAt,
	}
	if existing == nil {
		if _, err := l.svcCtx.PayRequestLogModel.Insert(l.ctx, requestLog); err != nil {
			if current, findErr := l.svcCtx.PayRequestLogModel.FindOneByRequestNo(l.ctx, requestNo); findErr == nil && current != nil {
				if currentOrder, orderErr := l.svcCtx.RechargeOrderModel.FindOne(l.ctx, order.Id); orderErr == nil {
					*order = *currentOrder
				}
				return nil
			}
			return err
		}
	} else {
		requestLog = existing
		requestLog.Status = int64(payment.PayRequestStatus_PAY_REQUEST_STATUS_PROCESSING)
		requestLog.UpdateTimes = startedAt
		if err := l.svcCtx.PayRequestLogModel.Update(l.ctx, requestLog); err != nil {
			return err
		}
	}

	result, createErr := adapter.CreatePayment(l.ctx, account, order)
	finishedAt := utils.NowMillis()
	requestLog.DurationMs = finishedAt - startedAt
	requestLog.UpdateTimes = finishedAt
	if createErr != nil {
		requestLog.Status = int64(payment.PayRequestStatus_PAY_REQUEST_STATUS_UNCERTAIN)
		requestLog.ThirdMessage = createErr.Error()
		_ = l.svcCtx.PayRequestLogModel.Update(l.ctx, requestLog)
		return createErr
	}

	requestLog.Status = int64(payment.PayRequestStatus_PAY_REQUEST_STATUS_SUCCESS)
	requestLog.RequestData = sql.NullString{String: result.RawRequest, Valid: result.RawRequest != ""}
	requestLog.ResponseData = sql.NullString{String: result.RawResponse, Valid: result.RawResponse != ""}
	if err := l.svcCtx.PayRequestLogModel.Update(l.ctx, requestLog); err != nil {
		return err
	}

	order.ThirdOrderNo = sql.NullString{String: result.ThirdOrderNo, Valid: result.ThirdOrderNo != ""}
	order.PayUrl = sql.NullString{String: result.PayURL, Valid: result.PayURL != ""}
	order.QrContent = sql.NullString{String: result.QRContent, Valid: result.QRContent != ""}
	order.ExpireTime = result.ExpireTime
	order.Status = int64(payment.PayOrderStatus_PAY_ORDER_STATUS_PAYING)
	order.UpdateTimes = finishedAt
	responseJSON, _ := json.Marshal(result)
	order.ResponseData = sql.NullString{String: string(responseJSON), Valid: true}
	return l.svcCtx.RechargeOrderModel.Update(l.ctx, order)
}
