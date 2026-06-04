package logic

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/notify"
	"wklive/common/utils"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"
	"wklive/services/payment/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type CreateCryptoRechargeOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateCryptoRechargeOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCryptoRechargeOrderLogic {
	return &CreateCryptoRechargeOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建链上充值订单
func (l *CreateCryptoRechargeOrderLogic) CreateCryptoRechargeOrder(in *payment.CreateCryptoRechargeOrderReq) (*payment.CreateCryptoRechargeOrderResp, error) {
	if in.RechargeAmount <= 0 {
		return &payment.CreateCryptoRechargeOrderResp{Base: helper.GetErrResp(i18n.AmountMustBePositive, i18n.Translate(i18n.AmountMustBePositive, l.ctx))}, nil
	}
	if in.WalletType <= 0 || in.Coin == "" || in.ChainCode == 0 {
		return &payment.CreateCryptoRechargeOrderResp{Base: helper.GetErrResp(i18n.InvalidCryptoRechargeParams, i18n.Translate(i18n.InvalidCryptoRechargeParams, l.ctx))}, nil
	}

	userId, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	tenantId, err := utils.GetTenantIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	clientIP, _ := utils.GetClientIPFromMd(l.ctx)

	addressItem, err := l.svcCtx.CryptoRechargeAddressModel.FindOneByTenantIdUserIdWalletTypeCoinChainCode(l.ctx, tenantId, userId, in.WalletType, in.Coin, int64(in.ChainCode))
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return &payment.CreateCryptoRechargeOrderResp{Base: helper.GetErrResp(i18n.CryptoRechargeAddressExpired, i18n.Translate(i18n.CryptoRechargeAddressExpired, l.ctx))}, nil
		}
		return nil, err
	}
	reservedByMe, err := cryptoRechargeAddressReservedBy(l.ctx, l.svcCtx, addressItem.Id, tenantId, userId)
	if err != nil {
		return nil, err
	}
	if !reservedByMe {
		return &payment.CreateCryptoRechargeOrderResp{Base: helper.GetErrResp(i18n.CryptoRechargeAddressExpired, i18n.Translate(i18n.CryptoRechargeAddressExpired, l.ctx))}, nil
	}

	now := utils.NowMillis()
	orderNo, err := l.svcCtx.GenerateOrderNo(l.ctx, "CRC")
	if err != nil {
		return nil, err
	}

	requestData, _ := json.Marshal(map[string]any{
		"walletType":     in.WalletType,
		"coin":           in.Coin,
		"chainCode":      int64(in.ChainCode),
		"rechargeAmount": in.RechargeAmount,
		"address":        addressItem.Address,
		"memo":           addressItem.Memo,
		"voucherImage":   in.VoucherImage,
	})
	rechargeOrder := &models.TRechargeOrder{
		TenantId:     tenantId,
		UserId:       userId,
		OrderNo:      orderNo,
		BizOrderNo:   sql.NullString{String: in.BizOrderNo, Valid: in.BizOrderNo != ""},
		RechargeType: int64(payment.RechargeType_RECHARGE_TYPE_CRYPTO),
		WalletType:   in.WalletType,
		Currency:     in.Coin,
		OrderAmount:  in.RechargeAmount,
		PayAmount:    in.RechargeAmount,
		Subject:      sql.NullString{String: "Crypto Recharge", Valid: true},
		Body:         sql.NullString{String: addressItem.Address, Valid: true},
		ClientType:   int64(in.ClientType),
		ClientIp:     sql.NullString{String: clientIP, Valid: clientIP != ""},
		Status:       int64(payment.PayOrderStatus_PAY_ORDER_STATUS_PAYING),
		QrContent:    sql.NullString{String: addressItem.Address, Valid: true},
		VoucherImage: sql.NullString{String: in.VoucherImage, Valid: true},
		RequestData:  sql.NullString{String: string(requestData), Valid: len(requestData) > 0},
		CreateTimes:  now,
		UpdateTimes:  now,
	}

	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		rechargeOrderModel := models.NewTRechargeOrderModel(conn, l.svcCtx.Config.CacheRedis).(models.RechargeOrderModel)
		userRechargeStatModel := models.NewTUserRechargeStatModel(conn, l.svcCtx.Config.CacheRedis).(models.UserRechargeStatModel)

		result, err := rechargeOrderModel.Insert(ctx, rechargeOrder)
		if err != nil {
			return err
		}
		if id, err := result.LastInsertId(); err == nil {
			rechargeOrder.Id = id
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

	l.Logger.Infof("Create recharge order success: %s, user_id: %d", orderNo, userId)
	event := notify.NewEvent(notify.EventTypeRecharge, notify.EventLevelInfo, "用户充值", fmt.Sprintf("用户 %d 发起充值订单 %s", userId, orderNo))
	event.Source = "payment"
	event.TenantID = tenantId
	event.UserID = userId
	event.BizNo = orderNo
	event.Data = map[string]any{
		"amount":   in.RechargeAmount,
		"currency": in.Coin,
	}
	if err := notify.Publish(l.ctx, l.svcCtx.Redis, event); err != nil {
		l.Errorf("publish admin recharge notification failed, orderNo=%s err=%v", orderNo, err)
	}
	releaseCryptoRechargeAddress(l.ctx, l.svcCtx, addressItem.Id)

	return &payment.CreateCryptoRechargeOrderResp{
		Base: helper.OkResp(),
		Data: &payment.CreateCryptoRechargeOrderData{
			Order:   toRechargeOrderProto(rechargeOrder),
			Address: toCryptoRechargeAddressProto(addressItem),
		},
	}, nil
}
