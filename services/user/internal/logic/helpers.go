package logic

import (
	"wklive/proto/user"
	"wklive/services/user/models"
)

func toUserBaseProto(tuser *models.TUser) *user.UserBase {
	if tuser == nil {
		return &user.UserBase{}
	}

	return &user.UserBase{
		Id:             tuser.Id,
		TenantId:       tuser.TenantId,
		UserNo:         tuser.UserNo,
		Username:       tuser.Username,
		Nickname:       tuser.Nickname.String,
		Avatar:         tuser.Avatar.String,
		PasswordHash:   tuser.PasswordHash,
		Language:       tuser.Language.String,
		Timezone:       tuser.Timezone.String,
		InviteCode:     tuser.InviteCode.String,
		Signature:      tuser.Signature.String,
		RegisterType:   user.RegisterType(tuser.RegisterType),
		Status:         user.UserStatus(tuser.Status),
		MemberLevel:    tuser.MemberLevel,
		Source:         tuser.Source.String,
		ReferrerUserId: tuser.ReferrerUserId.Int64,
		LastLoginIp:    tuser.LastLoginIp.String,
		LastLoginTime:  tuser.LastLoginTime,
		RegisterIp:     tuser.RegisterIp.String,
		RegisterTime:   tuser.RegisterTime,
		IsGuest:        tuser.IsGuest,
		IsRecharge:     tuser.IsRecharge,
		DeviceId:       tuser.DeviceId,
		Fingerprint:    tuser.Fingerprint,
		Remark:         tuser.Remark.String,
		Deleted:        tuser.Deleted,
		CreateTimes:    tuser.CreateTimes,
		UpdateTimes:    tuser.UpdateTimes,
	}
}

func toUserItemProto(tuser *models.TUser) *user.UserItem {
	if tuser == nil {
		return &user.UserItem{}
	}

	return &user.UserItem{
		Id:             tuser.Id,
		TenantId:       tuser.TenantId,
		UserNo:         tuser.UserNo,
		Username:       tuser.Username,
		Nickname:       tuser.Nickname.String,
		Avatar:         tuser.Avatar.String,
		PasswordHash:   tuser.PasswordHash,
		RegisterType:   user.RegisterType(tuser.RegisterType),
		Status:         user.UserStatus(tuser.Status),
		MemberLevel:    tuser.MemberLevel,
		Language:       tuser.Language.String,
		Timezone:       tuser.Timezone.String,
		InviteCode:     tuser.InviteCode.String,
		Signature:      tuser.Signature.String,
		Source:         tuser.Source.String,
		ReferrerUserId: tuser.ReferrerUserId.Int64,
		LastLoginIp:    tuser.LastLoginIp.String,
		LastLoginTime:  tuser.LastLoginTime,
		RegisterIp:     tuser.RegisterIp.String,
		RegisterTime:   tuser.RegisterTime,
		IsGuest:        tuser.IsGuest,
		IsRecharge:     tuser.IsRecharge,
		DeviceId:       tuser.DeviceId,
		Fingerprint:    tuser.Fingerprint,
		Remark:         tuser.Remark.String,
		Deleted:        tuser.Deleted,
		CreateTimes:    tuser.CreateTimes,
		UpdateTimes:    tuser.UpdateTimes,
	}
}

func toUserIdentityProto(identity *models.TUserIdentity) *user.UserIdentity {
	if identity == nil {
		return &user.UserIdentity{}
	}

	return &user.UserIdentity{
		Id:            identity.Id,
		TenantId:      identity.TenantId,
		UserId:        identity.UserId,
		Phone:         identity.Phone.String,
		Email:         identity.Email.String,
		RealName:      identity.RealName.String,
		Gender:        user.Gender(identity.Gender),
		Birthday:      identity.Birthday,
		CountryCode:   identity.CountryCode.String,
		Province:      identity.Province.String,
		City:          identity.City.String,
		Address:       identity.Address.String,
		IdType:        user.IdType(identity.IdType),
		IdNo:          identity.IdNo.String,
		FrontImage:    identity.FrontImage.String,
		BackImage:     identity.BackImage.String,
		HandheldImage: identity.HandheldImage.String,
		KycLevel:      user.KycLevel(identity.KycLevel),
		VerifyStatus:  user.VerifyStatus(identity.VerifyStatus),
		RejectReason:  identity.RejectReason.String,
		SubmitTime:    identity.SubmitTime,
		VerifyTime:    identity.VerifyTime,
		VerifyBy:      identity.VerifyBy.Int64,
		CreateTimes:   identity.CreateTimes,
		UpdateTimes:   identity.UpdateTimes,
	}
}

func toUserIdentityItemProto(identity *models.TUserIdentity) *user.UserIdentityItem {
	if identity == nil {
		return &user.UserIdentityItem{}
	}

	return &user.UserIdentityItem{
		Id:            identity.Id,
		TenantId:      identity.TenantId,
		UserId:        identity.UserId,
		Phone:         identity.Phone.String,
		Email:         identity.Email.String,
		RealName:      identity.RealName.String,
		Gender:        user.Gender(identity.Gender),
		Birthday:      identity.Birthday,
		CountryCode:   identity.CountryCode.String,
		Province:      identity.Province.String,
		City:          identity.City.String,
		Address:       identity.Address.String,
		IdType:        user.IdType(identity.IdType),
		IdNo:          identity.IdNo.String,
		FrontImage:    identity.FrontImage.String,
		BackImage:     identity.BackImage.String,
		HandheldImage: identity.HandheldImage.String,
		KycLevel:      user.KycLevel(identity.KycLevel),
		VerifyStatus:  user.VerifyStatus(identity.VerifyStatus),
		RejectReason:  identity.RejectReason.String,
		SubmitTime:    identity.SubmitTime,
		VerifyTime:    identity.VerifyTime,
		VerifyBy:      identity.VerifyBy.Int64,
		CreateTimes:   identity.CreateTimes,
		UpdateTimes:   identity.UpdateTimes,
	}
}

func toUserSecurityProto(security *models.TUserSecurity) *user.UserSecurity {
	if security == nil {
		return &user.UserSecurity{}
	}

	return &user.UserSecurity{
		Id:              security.Id,
		TenantId:        security.TenantId,
		UserId:          security.UserId,
		PayPasswordHash: security.PayPasswordHash.String,
		GoogleSecret:    security.GoogleSecret.String,
		GoogleEnabled:   security.GoogleEnabled,
		LoginErrorCount: security.LoginErrorCount,
		PayErrorCount:   security.PayErrorCount,
		LockUntil:       security.LockUntil,
		RiskLevel:       user.RiskLevel(security.RiskLevel),
		CreateTimes:     security.CreateTimes,
		UpdateTimes:     security.UpdateTimes,
	}
}

func toUserBankItemProto(bank *models.TUserBank) *user.UserBankItem {
	if bank == nil {
		return &user.UserBankItem{}
	}

	return &user.UserBankItem{
		Id:          bank.Id,
		TenantId:    bank.TenantId,
		UserId:      bank.UserId,
		BankName:    bank.BankName,
		BankCode:    bank.BankCode.String,
		AccountName: bank.AccountName,
		AccountNo:   bank.AccountNo,
		BranchName:  bank.BranchName.String,
		CountryCode: bank.CountryCode.String,
		IsDefault:   bank.IsDefault,
		Status:      user.BankStatus(bank.Status),
		CreateTimes: bank.CreateTimes,
		UpdateTimes: bank.UpdateTimes,
	}
}

func toUserBankItemListProto(items []*models.TUserBank) []*user.UserBankItem {
	if len(items) == 0 {
		return []*user.UserBankItem{}
	}

	list := make([]*user.UserBankItem, 0, len(items))
	for _, item := range items {
		list = append(list, toUserBankItemProto(item))
	}
	return list
}

func toUserProfileProto(tuser *models.TUser, identity *models.TUserIdentity, security *models.TUserSecurity) *user.UserProfile {
	return &user.UserProfile{
		Base:     toUserBaseProto(tuser),
		Identity: toUserIdentityProto(identity),
		Security: toUserSecurityProto(security),
	}
}

func toUserDetailProto(tuser *models.TUser, identity *models.TUserIdentity, security *models.TUserSecurity, banks []*models.TUserBank) *user.UserDetail {
	return &user.UserDetail{
		Base:     toUserBaseProto(tuser),
		Identity: toUserIdentityProto(identity),
		Security: toUserSecurityProto(security),
		Banks:    toUserBankItemListProto(banks),
	}
}
