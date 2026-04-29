package logic

import (
	"fmt"

	"wklive/common/conv"
	"wklive/proto/common"
	"wklive/proto/payment"
	"wklive/services/payment/models"
)

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
		Status:       payment.CommonStatus(item.Status),
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
		Status:      payment.CommonStatus(item.Status),
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
		Status:      payment.CommonStatus(item.Status),
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
		Status:              payment.CommonStatus(item.Status),
		IsDefault:           item.IsDefault,
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
		Visible:         item.Visible,
		Status:          payment.CommonStatus(item.Status),
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
		Status:               payment.CommonStatus(item.Status),
		SingleAmountMin:      item.SingleAmountMin,
		SingleAmountMax:      item.SingleAmountMax,
		UserTotalRechargeMin: item.UserTotalRechargeMin,
		UserTotalRechargeMax: item.UserTotalRechargeMax,
		MemberLevelMin:       item.MemberLevelMin,
		MemberLevelMax:       item.MemberLevelMax,
		KycLevelMin:          item.KycLevelMin,
		KycLevelMax:          item.KycLevelMax,
		AllowNewUser:         item.AllowNewUser,
		AllowOldUser:         item.AllowOldUser,
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
		WalletType:    item.WalletType,
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
		Status:               toCryptoWalletStatusProto(item.Status),
		IsDefault:            item.IsDefault,
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
	switch status {
	case payment.CryptoRechargeAddressStatus_CRYPTO_RECHARGE_ADDRESS_STATUS_DISABLED:
		return 0
	case payment.CryptoRechargeAddressStatus_CRYPTO_RECHARGE_ADDRESS_STATUS_ENABLED:
		return 1
	case payment.CryptoRechargeAddressStatus_CRYPTO_RECHARGE_ADDRESS_STATUS_FROZEN:
		return 2
	default:
		return defaultValue
	}
}

func toCryptoAddressStatusProto(status int64) payment.CryptoRechargeAddressStatus {
	switch status {
	case 0:
		return payment.CryptoRechargeAddressStatus_CRYPTO_RECHARGE_ADDRESS_STATUS_DISABLED
	case 1:
		return payment.CryptoRechargeAddressStatus_CRYPTO_RECHARGE_ADDRESS_STATUS_ENABLED
	case 2:
		return payment.CryptoRechargeAddressStatus_CRYPTO_RECHARGE_ADDRESS_STATUS_FROZEN
	default:
		return payment.CryptoRechargeAddressStatus_CRYPTO_RECHARGE_ADDRESS_STATUS_UNKNOWN
	}
}

func toCryptoWalletStatusDB(status payment.CommonStatus, defaultValue int64) int64 {
	switch status {
	case payment.CommonStatus_COMMON_STATUS_ENABLED:
		return 1
	case payment.CommonStatus_COMMON_STATUS_DISABLED:
		return 0
	default:
		return defaultValue
	}
}

func toCryptoWalletStatusProto(status int64) payment.CommonStatus {
	switch status {
	case 1:
		return payment.CommonStatus_COMMON_STATUS_ENABLED
	case 0:
		return payment.CommonStatus_COMMON_STATUS_DISABLED
	default:
		return payment.CommonStatus_COMMON_STATUS_UNKNOWN
	}
}
