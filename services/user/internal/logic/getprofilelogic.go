package logic

import (
	"context"
	"errors"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/user"
	"wklive/services/user/internal/svc"
	"wklive/services/user/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetProfileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProfileLogic {
	return &GetProfileLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取用户资料
func (l *GetProfileLogic) GetProfile(in *user.GetProfileReq) (*user.GetProfileResp, error) {
	// 查询用户基本信息
	tuser, err := l.svcCtx.UserModel.FindOne(l.ctx, in.UserId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if tuser == nil {
		return &user.GetProfileResp{
			Base: helper.GetErrResp(404, i18n.Translate(i18n.UserNotFound, l.ctx)),
		}, nil
	}

	// 查询用户身份信息
	userIdentity, err := l.svcCtx.UserIdentityModel.FindOneByTenantIdUserId(l.ctx, tuser.TenantId, in.UserId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	// 查询用户安全信息
	userSecurity, err := l.svcCtx.UserSecurityModel.FindOneByTenantIdUserId(l.ctx, tuser.TenantId, in.UserId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	userBase := &user.UserBase{
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
			Birthday:      userIdentity.Birthday,
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
			PayPasswordHash: userSecurity.PayPasswordHash.String,
			GoogleSecret:    userSecurity.GoogleSecret.String,
			GoogleEnabled:   userSecurity.GoogleEnabled,
			LoginErrorCount: userSecurity.LoginErrorCount,
			PayErrorCount:   userSecurity.PayErrorCount,
			LockUntil:       userSecurity.LockUntil,
			RiskLevel:       user.RiskLevel(userSecurity.RiskLevel),
			CreateTimes:     userSecurity.CreateTimes,
			UpdateTimes:     userSecurity.UpdateTimes,
		}
	}

	return &user.GetProfileResp{
		Base: helper.OkResp(),
		Profile: &user.UserProfile{
			Base:     userBase,
			Identity: userIdentityProto,
			Security: userSecurityProto,
		},
	}, nil
}
