package logic

import (
	"wklive/proto/system"
	"wklive/services/system/models"
)

func verificationCodeRecordItem(row *models.SysVerificationCodeRecord) *system.VerificationCodeRecordItem {
	if row == nil {
		return nil
	}
	return &system.VerificationCodeRecordItem{
		Id:           row.Id,
		TenantId:     row.TenantId,
		Channel:      system.VerificationCodeChannel(row.Channel),
		Target:       row.Target,
		Scene:        system.VerificationCodeScene(row.Scene),
		Code:         row.Code,
		Status:       system.VerificationCodeStatus(row.Status),
		Provider:     row.Provider.String,
		ErrorMessage: row.ErrorMessage.String,
		CreateTimes:  row.CreateTimes,
		UpdateTimes:  row.UpdateTimes,
	}
}
