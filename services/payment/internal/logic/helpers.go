package logic

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"wklive/common/conv"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/proto/common"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"
	"wklive/services/payment/models"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func systemAdminWriteScopeResp(ctx context.Context) (*common.RespBase, error) {
	userType, err := utils.GetUserTypeFromMd(ctx)
	if err != nil {
		return nil, i18n.StatusError(ctx, i18n.UserNotFound)
	}
	if userType != utils.SysUserTypeSystemAdmin {
		return helper.GetErrResp(i18n.PermissionDenied, i18n.Translate(i18n.PermissionDenied, ctx)), nil
	}
	return nil, nil
}

func rechargeTypeFromPlatform(item *models.TPayPlatform) payment.RechargeType {
	if item == nil {
		return payment.RechargeType_RECHARGE_TYPE_UNKNOWN
	}
	switch payment.PlatformType(item.PlatformType) {
	case payment.PlatformType_PLATFORM_TYPE_CHAIN:
		return payment.RechargeType_RECHARGE_TYPE_CRYPTO
	case payment.PlatformType_PLATFORM_TYPE_THIRD:
		return payment.RechargeType_RECHARGE_TYPE_THIRD
	case payment.PlatformType_PLATFORM_TYPE_BANK:
		return payment.RechargeType_RECHARGE_TYPE_BANK
	case payment.PlatformType_PLATFORM_TYPE_MANUAL:
		return payment.RechargeType_RECHARGE_TYPE_MANUAL
	default:
		return payment.RechargeType_RECHARGE_TYPE_OTHER
	}
}

func switchToProto(value int64) common.Switch {
	return common.Switch(value)
}

func switchToModel(value common.Switch, defaultValue int64) int64 {
	if value == common.Switch_SWITCH_UNKNOWN {
		return defaultValue
	}
	return int64(value)
}

func markRechargeOrderSuccessAndCredit(ctx context.Context, svcCtx *svc.ServiceContext, order *models.TRechargeOrder, thirdTradeNo string, payAmount int64, remark string) error {
	if order == nil {
		return i18n.StatusError(ctx, i18n.OrderNotFound)
	}

	return svcCtx.DB.TransactCtx(ctx, func(txCtx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		rechargeOrderModel := models.NewTRechargeOrderModel(conn, svcCtx.Config.CacheRedis)

		current, err := rechargeOrderModel.FindOne(txCtx, order.Id)
		if err != nil {
			return err
		}
		if current.Status == int64(payment.PayOrderStatus_PAY_ORDER_STATUS_SUCCESS) {
			return nil
		}
		if current.Status != int64(payment.PayOrderStatus_PAY_ORDER_STATUS_PENDING) &&
			current.Status != int64(payment.PayOrderStatus_PAY_ORDER_STATUS_PAYING) {
			return i18n.StatusError(txCtx, i18n.OnlyPendingPaymentOrdersCanMarkSuccess)
		}

		now := utils.NowMillis()
		current.Status = int64(payment.PayOrderStatus_PAY_ORDER_STATUS_SUCCESS)
		if payAmount > 0 {
			current.PayAmount = payAmount
		} else if current.PayAmount <= 0 {
			current.PayAmount = current.OrderAmount
		}
		if thirdTradeNo != "" {
			current.ThirdTradeNo = sql.NullString{String: thirdTradeNo, Valid: true}
		}
		current.PaidTime = now
		current.UpdateTimes = now

		if err := creditRechargeOrderAsset(txCtx, svcCtx, current, remark); err != nil {
			return err
		}
		return rechargeOrderModel.Update(txCtx, current)
	})
}

func creditRechargeOrderAsset(ctx context.Context, svcCtx *svc.ServiceContext, order *models.TRechargeOrder, remark string) error {
	if order == nil {
		return i18n.StatusError(ctx, i18n.OrderNotFound)
	}
	amount := order.PayAmount
	if amount <= 0 {
		amount = order.OrderAmount
	}
	resp, err := svcCtx.AssetCli.AddAvailable(ctx, &asset.AddAvailableReq{
		TenantId:   order.TenantId,
		UserId:     order.UserId,
		WalletType: rechargeOrderWalletType(order),
		Coin:       order.Currency,
		Amount:     strconv.FormatInt(amount, 10),
		BizType:    asset.BizType_BIZ_TYPE_PAYMENT,
		SceneType:  asset.SceneType_SCENE_TYPE_RECHARGE,
		BizId:      order.Id,
		BizNo:      order.OrderNo,
		Remark:     remark,
	})
	if err != nil {
		return err
	}
	if resp == nil || resp.Base == nil {
		return i18n.StatusError(ctx, i18n.InternalServerError)
	}
	if resp.Base.Code != 200 {
		return i18n.StatusError(ctx, resp.Base.Code)
	}
	return nil
}

func freezeWithdrawOrderAsset(ctx context.Context, svcCtx *svc.ServiceContext, order *models.TWithdrawOrder, remark string) error {
	if order == nil {
		return i18n.StatusError(ctx, i18n.OrderNotFound)
	}
	resp, err := svcCtx.AssetCli.FreezeAsset(ctx, &asset.FreezeAssetReq{
		TenantId:   order.TenantId,
		UserId:     order.UserId,
		WalletType: common.WalletType_WALLET_TYPE_SPOT,
		Coin:       order.Currency,
		Amount:     strconv.FormatInt(order.Amount, 10),
		BizType:    asset.BizType_BIZ_TYPE_PAYMENT,
		SceneType:  asset.SceneType_SCENE_TYPE_WITHDRAW_APPLY,
		BizId:      order.Id,
		BizNo:      order.OrderNo,
		Remark:     remark,
	})
	if err != nil {
		return err
	}
	if resp == nil || resp.Base == nil {
		return i18n.StatusError(ctx, i18n.InternalServerError)
	}
	if resp.Base.Code != 200 {
		return i18n.StatusError(ctx, resp.Base.Code)
	}
	return nil
}

func deductWithdrawOrderFrozenAsset(ctx context.Context, svcCtx *svc.ServiceContext, order *models.TWithdrawOrder, remark string) error {
	if order == nil {
		return i18n.StatusError(ctx, i18n.OrderNotFound)
	}
	resp, err := svcCtx.AssetCli.DeductFrozenAssetByBizNo(ctx, &asset.DeductFrozenAssetByBizNoReq{
		TenantId:      order.TenantId,
		TargetBizType: asset.BizType_BIZ_TYPE_PAYMENT,
		TargetBizNo:   order.OrderNo,
		Amount:        strconv.FormatInt(order.Amount, 10),
		BizType:       asset.BizType_BIZ_TYPE_PAYMENT,
		SceneType:     asset.SceneType_SCENE_TYPE_WITHDRAW_SUCCESS,
		BizId:         order.Id,
		BizNo:         order.OrderNo,
		Remark:        remark,
	})
	if err != nil {
		return err
	}
	if resp == nil || resp.Base == nil {
		return i18n.StatusError(ctx, i18n.InternalServerError)
	}
	if resp.Base.Code != 200 {
		return i18n.StatusError(ctx, resp.Base.Code)
	}
	return nil
}

func unfreezeWithdrawOrderAsset(ctx context.Context, svcCtx *svc.ServiceContext, order *models.TWithdrawOrder, remark string) error {
	if order == nil {
		return i18n.StatusError(ctx, i18n.OrderNotFound)
	}
	resp, err := svcCtx.AssetCli.UnfreezeAssetByBizNo(ctx, &asset.UnfreezeAssetByBizNoReq{
		TenantId:      order.TenantId,
		TargetBizType: asset.BizType_BIZ_TYPE_PAYMENT,
		TargetBizNo:   order.OrderNo,
		Amount:        strconv.FormatInt(order.Amount, 10),
		BizType:       asset.BizType_BIZ_TYPE_PAYMENT,
		SceneType:     asset.SceneType_SCENE_TYPE_WITHDRAW_REJECT,
		BizId:         order.Id,
		BizNo:         order.OrderNo,
		Remark:        remark,
	})
	if err != nil {
		return err
	}
	if resp == nil || resp.Base == nil {
		return i18n.StatusError(ctx, i18n.InternalServerError)
	}
	if resp.Base.Code != 200 {
		return i18n.StatusError(ctx, resp.Base.Code)
	}
	return nil
}

func rechargeOrderWalletType(order *models.TRechargeOrder) common.WalletType {
	if order == nil || order.WalletType <= 0 {
		return common.WalletType_WALLET_TYPE_SPOT
	}
	return common.WalletType(order.WalletType)
}

func toPayPlatformProto(item *models.TPayPlatform) *payment.PayPlatform {
	if item == nil {
		return nil
	}
	return &payment.PayPlatform{
		Id:           item.Id,
		PlatformCode: item.PlatformCode,
		PlatformName: item.PlatformName,
		PlatformType: payment.PlatformType(item.PlatformType),
		NotifyUrl:    item.NotifyUrl.String,
		ReturnUrl:    item.ReturnUrl.String,
		Icon:         item.Icon.String,
		Enabled:      common.Enable(item.Enabled),
		Remark:       item.Remark.String,
		CreateTimes:  item.CreateTimes,
		UpdateTimes:  item.UpdateTimes,
	}
}

func toPayProductProto(item *models.TPayProduct) *payment.PayProduct {
	if item == nil {
		return nil
	}
	return &payment.PayProduct{
		Id:          item.Id,
		PlatformId:  item.PlatformId,
		ProductCode: item.ProductCode,
		ProductName: item.ProductName,
		SceneType:   payment.SceneType(item.SceneType),
		Currency:    item.Currency,
		Enabled:     common.Enable(item.Enabled),
		Remark:      item.Remark.String,
		CreateTimes: item.CreateTimes,
		UpdateTimes: item.UpdateTimes,
	}
}

func toTenantPayPlatformProto(item *models.TTenantPayPlatform) *payment.TenantPayPlatform {
	if item == nil {
		return nil
	}
	return &payment.TenantPayPlatform{
		Id:          item.Id,
		TenantId:    item.TenantId,
		PlatformId:  item.PlatformId,
		Enabled:     common.Enable(item.Enabled),
		OpenStatus:  payment.OpenStatus(item.OpenStatus),
		Remark:      item.Remark.String,
		CreateTimes: item.CreateTimes,
		UpdateTimes: item.UpdateTimes,
	}
}

func toTenantPayAccountProto(item *models.TTenantPayAccount) *payment.TenantPayAccount {
	if item == nil {
		return nil
	}
	return &payment.TenantPayAccount{
		Id:                  item.Id,
		TenantId:            item.TenantId,
		TenantPayPlatformId: item.TenantPayPlatformId,
		PlatformId:          item.PlatformId,
		AccountCode:         item.AccountCode,
		AccountName:         item.AccountName,
		AppId:               item.AppId.String,
		MerchantId:          item.MerchantId.String,
		MerchantName:        item.MerchantName.String,
		ApiKeyCipher:        item.ApiKeyCipher.String,
		ApiSecretCipher:     item.ApiSecretCipher.String,
		PrivateKeyCipher:    item.PrivateKeyCipher.String,
		PublicKey:           item.PublicKey.String,
		CertCipher:          item.CertCipher.String,
		ExtConfig:           item.ExtConfig.String,
		Enabled:             common.Enable(item.Enabled),
		IsDefault:           common.YesNo(item.IsDefault),
		Remark:              item.Remark.String,
		CreateTimes:         item.CreateTimes,
		UpdateTimes:         item.UpdateTimes,
	}
}

func toTenantPayChannelProto(item *models.TTenantPayChannel) *payment.TenantPayChannel {
	if item == nil {
		return nil
	}
	return &payment.TenantPayChannel{
		Id:              item.Id,
		TenantId:        item.TenantId,
		PlatformId:      item.PlatformId,
		ProductId:       item.ProductId,
		AccountId:       item.AccountId,
		ChannelCode:     item.ChannelCode,
		ChannelName:     item.ChannelName,
		DisplayName:     item.DisplayName.String,
		Icon:            item.Icon.String,
		Currency:        item.Currency,
		Sort:            item.Sort,
		Visible:         switchToProto(item.Visible),
		Enabled:         common.Enable(item.Enabled),
		SingleMinAmount: item.SingleMinAmount,
		SingleMaxAmount: item.SingleMaxAmount,
		DailyMaxAmount:  item.DailyMaxAmount,
		DailyMaxCount:   item.DailyMaxCount,
		FeeType:         payment.FeeType(item.FeeType),
		FeeRate:         fmt.Sprintf("%f", item.FeeRate),
		FeeFixedAmount:  item.FeeFixedAmount,
		ExtConfig:       item.ExtConfig.String,
		Remark:          item.Remark.String,
		CreateTimes:     item.CreateTimes,
		UpdateTimes:     item.UpdateTimes,
	}
}

func toVisiblePayChannelProto(item *models.TTenantPayChannel) *payment.VisiblePayChannel {
	if item == nil {
		return nil
	}
	return &payment.VisiblePayChannel{
		ChannelId:   item.Id,
		ChannelName: item.ChannelName,
	}
}

func toTenantPayChannelRuleProto(item *models.TTenantPayChannelRule) *payment.TenantPayChannelRule {
	if item == nil {
		return nil
	}
	return &payment.TenantPayChannelRule{
		Id:                   item.Id,
		TenantId:             item.TenantId,
		ChannelId:            item.ChannelId,
		RuleName:             item.RuleName,
		Priority:             item.Priority,
		Enabled:              common.Enable(item.Enabled),
		SingleAmountMin:      item.SingleAmountMin,
		SingleAmountMax:      item.SingleAmountMax,
		UserTotalRechargeMin: item.UserTotalRechargeMin,
		UserTotalRechargeMax: item.UserTotalRechargeMax,
		MemberLevelMin:       item.MemberLevelMin,
		MemberLevelMax:       item.MemberLevelMax,
		KycLevelMin:          item.KycLevelMin,
		KycLevelMax:          item.KycLevelMax,
		AllowNewUser:         common.YesNo(item.AllowNewUser),
		AllowOldUser:         common.YesNo(item.AllowOldUser),
		AllowTags:            item.AllowTags.String,
		DenyTags:             item.DenyTags.String,
		Remark:               item.Remark.String,
		CreateTimes:          item.CreateTimes,
		UpdateTimes:          item.UpdateTimes,
	}
}

func toUserRechargeStatProto(item *models.TUserRechargeStat) *payment.UserRechargeStat {
	if item == nil {
		return nil
	}
	return &payment.UserRechargeStat{
		Id:                 item.Id,
		TenantId:           item.TenantId,
		UserId:             item.UserId,
		SuccessOrderCount:  item.SuccessOrderCount,
		SuccessTotalAmount: item.SuccessTotalAmount,
		TodaySuccessAmount: item.TodaySuccessAmount,
		TodaySuccessCount:  item.TodaySuccessCount,
		FirstSuccessTime:   item.FirstSuccessTime.Int64,
		LastSuccessTime:    item.LastSuccessTime.Int64,
		CreateTimes:        item.CreateTimes,
		UpdateTimes:        item.UpdateTimes,
	}
}

func toRechargeOrderProto(item *models.TRechargeOrder) *payment.RechargeOrder {
	if item == nil {
		return nil
	}
	return &payment.RechargeOrder{
		Id:           item.Id,
		TenantId:     item.TenantId,
		UserId:       item.UserId,
		OrderNo:      item.OrderNo,
		BizOrderNo:   item.BizOrderNo.String,
		PlatformId:   item.PlatformId,
		ProductId:    item.ProductId,
		AccountId:    item.AccountId,
		ChannelId:    item.ChannelId,
		RechargeType: payment.RechargeType(item.RechargeType),
		WalletType:   common.WalletType(item.WalletType),
		Currency:     item.Currency,
		OrderAmount:  item.OrderAmount,
		PayAmount:    item.PayAmount,
		FeeAmount:    item.FeeAmount,
		Subject:      item.Subject.String,
		Body:         item.Body.String,
		ClientType:   payment.ClientType(item.ClientType),
		ClientIp:     item.ClientIp.String,
		Status:       payment.PayOrderStatus(item.Status),
		ThirdTradeNo: item.ThirdTradeNo.String,
		ThirdOrderNo: item.ThirdOrderNo.String,
		PayUrl:       item.PayUrl.String,
		QrContent:    item.QrContent.String,
		VoucherImage: item.VoucherImage,
		RequestData:  item.RequestData.String,
		ResponseData: item.ResponseData.String,
		NotifyData:   item.NotifyData.String,
		ExpireTime:   item.ExpireTime,
		PaidTime:     item.PaidTime,
		NotifyTime:   item.NotifyTime,
		CloseTime:    item.CloseTime,
		Remark:       item.Remark.String,
		CreateTimes:  item.CreateTimes,
		UpdateTimes:  item.UpdateTimes,
	}
}

func toRechargeNotifyLogProto(item *models.TRechargeNotifyLog) *payment.PayNotifyLog {
	if item == nil {
		return nil
	}
	return &payment.PayNotifyLog{
		Id:            item.Id,
		TenantId:      item.TenantId,
		OrderId:       item.OrderId.Int64,
		OrderNo:       item.OrderNo.String,
		PlatformId:    item.PlatformId,
		ChannelId:     item.ChannelId.Int64,
		NotifyStatus:  payment.NotifyProcessStatus(item.NotifyStatus),
		NotifyBody:    item.NotifyBody.String,
		SignResult:    payment.SignResult(item.SignResult),
		ProcessResult: item.ProcessResult.String,
		ErrorMessage:  item.ErrorMessage.String,
		NotifyTime:    item.NotifyTime,
		CreateTimes:   item.CreateTimes,
	}
}

func toWithdrawNotifyLogProto(item *models.TWithdrawNotifyLog) *payment.PayNotifyLog {
	if item == nil {
		return nil
	}
	return &payment.PayNotifyLog{
		Id:            item.Id,
		TenantId:      item.TenantId,
		OrderId:       item.OrderId.Int64,
		OrderNo:       item.OrderNo.String,
		PlatformId:    item.PlatformId,
		ChannelId:     item.ChannelId.Int64,
		NotifyStatus:  payment.NotifyProcessStatus(item.NotifyStatus),
		NotifyBody:    item.NotifyBody.String,
		SignResult:    payment.SignResult(item.SignResult),
		ProcessResult: item.ProcessResult.String,
		ErrorMessage:  item.ErrorMessage.String,
		NotifyTime:    item.NotifyTime,
		CreateTimes:   item.CreateTimes,
	}
}

func toWithdrawOrderProto(item *models.TWithdrawOrder) *payment.WithdrawOrder {
	if item == nil {
		return nil
	}
	return &payment.WithdrawOrder{
		Id:           item.Id,
		TenantId:     item.TenantId,
		UserId:       item.UserId,
		OrderNo:      item.OrderNo,
		BizOrderNo:   item.BizOrderNo.String,
		PlatformId:   item.PlatformId,
		ProductId:    item.ProductId,
		AccountId:    item.AccountId,
		ChannelId:    item.ChannelId,
		Currency:     item.Currency,
		Amount:       item.Amount,
		FeeAmount:    item.FeeAmount,
		ActualAmount: item.ActualAmount,
		ClientType:   payment.ClientType(item.ClientType),
		ClientIp:     item.ClientIp.String,
		Status:       payment.PayOrderStatus(item.Status),
		ThirdTradeNo: item.ThirdTradeNo.String,
		ThirdOrderNo: item.ThirdOrderNo.String,
		RequestData:  item.RequestData.String,
		ResponseData: item.ResponseData.String,
		NotifyData:   item.NotifyData.String,
		ProcessTime:  item.ProcessTime,
		NotifyTime:   item.NotifyTime,
		CloseTime:    item.CloseTime,
		Remark:       item.Remark.String,
		CreateTimes:  item.CreateTimes,
		UpdateTimes:  item.UpdateTimes,
	}
}

func toCryptoRechargeAddressProto(item *models.TCryptoRechargeAddress) *payment.CryptoRechargeAddress {
	if item == nil {
		return nil
	}
	return &payment.CryptoRechargeAddress{
		Id:            item.Id,
		TenantId:      item.TenantId,
		UserId:        item.UserId,
		WalletType:    common.WalletType(item.WalletType),
		Coin:          item.Coin,
		ChainCode:     common.ChainCode(item.ChainCode),
		Address:       item.Address,
		Memo:          item.Memo,
		AddressSource: payment.CryptoRechargeAddressSource(item.AddressSource),
		AddressType:   payment.CryptoRechargeAddressType(item.AddressType),
		Status:        toCryptoAddressStatusProto(item.Status),
		LastUsedTime:  item.LastUsedTime,
		CreateTimes:   item.CreateTimes,
		UpdateTimes:   item.UpdateTimes,
	}
}

func toCryptoWalletAccountProto(item *models.TCryptoWalletAccount) *payment.CryptoWalletAccount {
	if item == nil {
		return nil
	}
	return &payment.CryptoWalletAccount{
		Id:                   item.Id,
		TenantId:             item.TenantId,
		AccountCode:          item.AccountCode,
		AccountName:          item.AccountName,
		Provider:             item.Provider,
		ApiKeyCipher:         item.ApiKeyCipher.String,
		ApiSecretCipher:      item.ApiSecretCipher.String,
		CallbackSecretCipher: item.CallbackSecretCipher.String,
		ExtConfig:            item.ExtConfig.String,
		Enabled:              toCryptoWalletStatusProto(item.Enabled),
		IsDefault:            common.YesNo(item.IsDefault),
		CreateTimes:          item.CreateTimes,
		UpdateTimes:          item.UpdateTimes,
	}
}

func toCryptoRechargeTxProto(item *models.TCryptoRechargeTx) *payment.CryptoRechargeTx {
	if item == nil {
		return nil
	}
	return &payment.CryptoRechargeTx{
		Id:                   item.Id,
		TenantId:             item.TenantId,
		UserId:               item.UserId,
		OrderId:              item.OrderId,
		OrderNo:              item.OrderNo,
		Coin:                 item.Coin,
		ChainCode:            common.ChainCode(item.ChainCode),
		TxHash:               item.TxHash,
		FromAddress:          item.FromAddress,
		ToAddress:            item.ToAddress,
		Memo:                 item.Memo,
		Amount:               conv.FloatString(item.Amount),
		BlockHeight:          item.BlockHeight,
		ConfirmCount:         item.ConfirmCount,
		RequiredConfirmCount: item.RequiredConfirmCount,
		Status:               payment.CryptoRechargeTxStatus(item.Status),
		RawData:              item.RawData.String,
		CreateTimes:          item.CreateTimes,
		UpdateTimes:          item.UpdateTimes,
	}
}

func toCryptoAddressStatusDB(status payment.CryptoRechargeAddressStatus, defaultValue int64) int64 {
	if status == payment.CryptoRechargeAddressStatus_CRYPTO_RECHARGE_ADDRESS_STATUS_UNKNOWN {
		return defaultValue
	}
	return int64(status)
}

func toCryptoAddressStatusProto(status int64) payment.CryptoRechargeAddressStatus {
	return payment.CryptoRechargeAddressStatus(status)
}

func toCryptoWalletStatusDB(status common.Enable, defaultValue int64) int64 {
	if status == common.Enable_ENABLE_UNKNOWN {
		return defaultValue
	}
	return int64(status)
}

func toCryptoWalletStatusProto(status int64) common.Enable {
	return common.Enable(status)
}

func enableToModel(enabled common.Enable, defaultValue int64) int64 {
	if enabled == common.Enable_ENABLE_UNKNOWN {
		return defaultValue
	}
	return int64(enabled)
}
