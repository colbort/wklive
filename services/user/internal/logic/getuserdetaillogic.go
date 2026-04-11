package logic

import (
	"context"
	"errors"
	"github.com/zeromicro/go-zero/core/logx"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/user"
	"wklive/services/user/internal/svc"
	"wklive/services/user/models"
)

type GetUserDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserDetailLogic {
	return &GetUserDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 管理员获取用户详情（聚合查询）
func (l *GetUserDetailLogic) GetUserDetail(in *user.GetUserDetailReq) (*user.GetUserDetailResp, error) {
	// 获取用户基本信息
	tuser, err := l.svcCtx.UserModel.FindOne(l.ctx, in.UserId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if tuser == nil {
		return &user.GetUserDetailResp{
			Base: helper.GetErrResp(404, i18n.Translate(i18n.UserNotFound, l.ctx)),
		}, nil
	}

	// 查询身份信息
	userIdentity, _ := l.svcCtx.UserIdentityModel.FindOneByTenantIdUserId(l.ctx, in.TenantId, in.UserId)

	// 查询安全信息
	userSecurity, _ := l.svcCtx.UserSecurityModel.FindOneByTenantIdUserId(l.ctx, in.TenantId, in.UserId)

	userBanks, _, err := l.svcCtx.UserBankModel.FindPage(l.ctx, in.TenantId, in.UserId, 0, 100)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	bankList := make([]*user.UserBank, 0, len(userBanks))
	for _, bank := range userBanks {
		bankList = append(bankList, &user.UserBank{
			Id:              bank.Id,
			TenantId:        bank.TenantId,
			UserId:          bank.UserId,
			BankName:        bank.BankName,
			BankCode:        bank.BankCode.String,
			AccountName:     bank.AccountName,
			AccountNo:       bank.AccountNo,
			MaskedAccountNo: maskAccountNo(bank.AccountNo),
			BranchName:      bank.BranchName.String,
			CountryCode:     bank.CountryCode.String,
			IsDefault:       bank.IsDefault == 1,
			Status:          user.BankStatus(bank.Status),
			CreateTimes:     bank.CreateTimes,
			UpdateTimes:     bank.UpdateTimes,
		})
	}

	userBase := &user.UserBase{
		Id:             tuser.Id,
		TenantId:       tuser.TenantId,
		UserNo:         tuser.UserNo,
		Username:       tuser.Username,
		Nickname:       tuser.Nickname.String,
		Avatar:         tuser.Avatar.String,
		Language:       tuser.Language.String,
		Timezone:       tuser.Timezone.String,
		InviteCode:     tuser.InviteCode.String,
		Signature:      tuser.Signature.String,
		RegisterType:   user.RegisterType(tuser.RegisterType),
		Status:         user.UserStatus(tuser.Status),
		MemberLevel:    int32(tuser.MemberLevel),
		Source:         tuser.Source.String,
		ReferrerUserId: tuser.ReferrerUserId.Int64,
		LastLoginIp:    tuser.LastLoginIp.String,
		LastLoginTime:  tuser.LastLoginTime,
		RegisterIp:     tuser.RegisterIp.String,
		RegisterTime:   tuser.RegisterTime,
		Remark:         tuser.Remark.String,
		Deleted:        tuser.Deleted == 1,
		CreateTimes:    tuser.CreateTimes,
		UpdateTimes:    tuser.UpdateTimes,
	}

	userIdentityProto := &user.UserIdentity{}
	if userIdentity != nil {
		userIdentityProto = &user.UserIdentity{
			Id:            userIdentity.Id,
			TenantId:      userIdentity.TenantId,
			UserId:        userIdentity.UserId,
			Phone:         userIdentity.Phone.String,
			Email:         userIdentity.Email.String,
			RealName:      userIdentity.RealName.String,
			Gender:        user.Gender(userIdentity.Gender),
			Birthday:      userIdentity.Birthday.Time.Format("2006-01-02"),
			CountryCode:   userIdentity.CountryCode.String,
			Province:      userIdentity.Province.String,
			City:          userIdentity.City.String,
			Address:       userIdentity.Address.String,
			IdType:        user.IdType(userIdentity.IdType),
			IdNo:          userIdentity.IdNo.String,
			FrontImage:    userIdentity.FrontImage.String,
			BackImage:     userIdentity.BackImage.String,
			HandheldImage: userIdentity.HandheldImage.String,
			KycLevel:      user.KycLevel(userIdentity.KycLevel),
			VerifyStatus:  user.VerifyStatus(userIdentity.VerifyStatus),
			RejectReason:  userIdentity.RejectReason.String,
			SubmitTime:    userIdentity.SubmitTime,
			VerifyTime:    userIdentity.VerifyTime,
			VerifyBy:      userIdentity.VerifyBy.Int64,
			CreateTimes:   userIdentity.CreateTimes,
			UpdateTimes:   userIdentity.UpdateTimes,
		}
	}

	userSecurityProto := &user.UserSecurity{}
	if userSecurity != nil {
		userSecurityProto = &user.UserSecurity{
			Id:              userSecurity.Id,
			TenantId:        userSecurity.TenantId,
			UserId:          userSecurity.UserId,
			HasPayPassword:  userSecurity.PayPasswordHash.Valid && userSecurity.PayPasswordHash.String != "",
			GoogleEnabled:   userSecurity.GoogleEnabled == 1,
			LoginErrorCount: userSecurity.LoginErrorCount,
			PayErrorCount:   userSecurity.PayErrorCount,
			LockUntil:       userSecurity.LockUntil,
			RiskLevel:       user.RiskLevel(userSecurity.RiskLevel),
			CreateTimes:     userSecurity.CreateTimes,
			UpdateTimes:     userSecurity.UpdateTimes,
		}
	}

	return &user.GetUserDetailResp{
		Base: helper.OkResp(),
		Detail: &user.UserDetail{
			Base:     userBase,
			Identity: userIdentityProto,
			Security: userSecurityProto,
			Banks:    bankList,
		},
	}, nil
}

func maskAccountNo(accountNo string) string {
	if len(accountNo) <= 4 {
		return accountNo
	}
	return "****" + accountNo[len(accountNo)-4:]
}
