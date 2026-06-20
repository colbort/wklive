package logic

import (
	"wklive/proto/system"
	"wklive/services/system/models"
)

func sysChatMerchantToProto(v *models.SysChatMerchant) *system.SysChatMerchantItem {
	if v == nil {
		return nil
	}
	return &system.SysChatMerchantItem{
		Id:           v.Id,
		MerchantCode: v.MerchantCode,
		MerchantName: v.MerchantName,
		Enabled:      commonStatusToProto(v.Enabled),
		ExpireTime:   v.ExpireTime,
		ContactName:  v.ContactName.String,
		ContactPhone: v.ContactPhone.String,
		ContactEmail: v.ContactEmail.String,
		Remark:       v.Remark.String,
		CreateBy:     v.CreateBy.String,
		CreateTimes:  v.CreateTimes,
		UpdateBy:     v.UpdateBy.String,
		UpdateTimes:  v.UpdateTimes,
	}
}
