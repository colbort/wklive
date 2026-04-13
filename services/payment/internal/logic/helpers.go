package logic

import (
	"fmt"

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
		ExpireTime:   item.ExpireTime.Int64,
		PaidTime:     item.PaidTime.Int64,
		NotifyTime:   item.NotifyTime.Int64,
		CloseTime:    item.CloseTime.Int64,
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
		NotifyTime:    item.NotifyTime.Int64,
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
		NotifyTime:    item.NotifyTime.Int64,
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
		ProcessTime:  item.ProcessTime.Int64,
		NotifyTime:   item.NotifyTime.Int64,
		CloseTime:    item.CloseTime.Int64,
		Remark:       item.Remark.String,
		CreateTimes:  item.CreateTimes,
		UpdateTimes:  item.UpdateTimes,
	}
}
