package payment

import (
	"wklive/app-api/internal/types"
	"wklive/proto/payment"
)

func respBase(base interface {
	GetCode() int32
	GetMsg() string
	GetTotal() int64
	GetHasNext() bool
	GetHasPrev() bool
	GetNextCursor() int64
	GetPrevCursor() int64
}) types.RespBase {
	return types.RespBase{
		Code:       base.GetCode(),
		Msg:        base.GetMsg(),
		Total:      base.GetTotal(),
		HasNext:    base.GetHasNext(),
		HasPrev:    base.GetHasPrev(),
		NextCursor: base.GetNextCursor(),
		PrevCursor: base.GetPrevCursor(),
	}
}

func cryptoRechargeAddressFromPB(item *payment.CryptoRechargeAddress) types.CryptoRechargeAddress {
	if item == nil {
		return types.CryptoRechargeAddress{}
	}
	return types.CryptoRechargeAddress{
		Id:            item.Id,
		TenantId:      item.TenantId,
		UserId:        item.UserId,
		WalletType:    item.WalletType,
		Coin:          item.Coin,
		ChainCode:     int64(item.ChainCode.Number()),
		Address:       item.Address,
		Memo:          item.Memo,
		AddressSource: int64(item.AddressSource.Number()),
		AddressType:   int64(item.AddressType.Number()),
		Status:        int64(item.Status.Number()),
		LastUsedTime:  item.LastUsedTime,
		CreateTimes:   item.CreateTimes,
		UpdateTimes:   item.UpdateTimes,
	}
}

func rechargeOrderFromPB(item *payment.RechargeOrder) types.RechargeOrder {
	if item == nil {
		return types.RechargeOrder{}
	}
	return types.RechargeOrder{
		Id:           item.Id,
		TenantId:     item.TenantId,
		UserId:       item.UserId,
		OrderNo:      item.OrderNo,
		BizOrderNo:   item.BizOrderNo,
		PlatformId:   item.PlatformId,
		ProductId:    item.ProductId,
		AccountId:    item.AccountId,
		ChannelId:    item.ChannelId,
		Currency:     item.Currency,
		OrderAmount:  item.OrderAmount,
		PayAmount:    item.PayAmount,
		FeeAmount:    item.FeeAmount,
		Subject:      item.Subject,
		Body:         item.Body,
		ClientType:   int64(item.ClientType.Number()),
		ClientIp:     item.ClientIp,
		Status:       int64(item.Status.Number()),
		ThirdTradeNo: item.ThirdTradeNo,
		ThirdOrderNo: item.ThirdOrderNo,
		PayUrl:       item.PayUrl,
		QrContent:    item.QrContent,
		RequestData:  item.RequestData,
		ResponseData: item.ResponseData,
		NotifyData:   item.NotifyData,
		ExpireTime:   item.ExpireTime,
		PaidTime:     item.PaidTime,
		NotifyTime:   item.NotifyTime,
		CloseTime:    item.CloseTime,
		Remark:       item.Remark,
		CreateTimes:  item.CreateTimes,
		UpdateTimes:  item.UpdateTimes,
	}
}

func cryptoRechargeTxFromPB(item *payment.CryptoRechargeTx) types.CryptoRechargeTx {
	if item == nil {
		return types.CryptoRechargeTx{}
	}
	return types.CryptoRechargeTx{
		Id:                   item.Id,
		TenantId:             item.TenantId,
		UserId:               item.UserId,
		OrderId:              item.OrderId,
		OrderNo:              item.OrderNo,
		Coin:                 item.Coin,
		ChainCode:            int64(item.ChainCode.Number()),
		TxHash:               item.TxHash,
		FromAddress:          item.FromAddress,
		ToAddress:            item.ToAddress,
		Memo:                 item.Memo,
		Amount:               item.Amount,
		BlockHeight:          item.BlockHeight,
		ConfirmCount:         item.ConfirmCount,
		RequiredConfirmCount: item.RequiredConfirmCount,
		Status:               int64(item.Status.Number()),
		RawData:              item.RawData,
		CreateTimes:          item.CreateTimes,
		UpdateTimes:          item.UpdateTimes,
	}
}
