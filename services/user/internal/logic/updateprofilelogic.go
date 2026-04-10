package logic

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"wklive/common/helper"
	"wklive/proto/common"
	"wklive/proto/user"
	"wklive/services/user/internal/svc"
	"wklive/services/user/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateProfileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProfileLogic {
	return &UpdateProfileLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新用户资料
func (l *UpdateProfileLogic) UpdateProfile(in *user.UpdateProfileReq) (*user.UpdateProfileResp, error) {
	// 获取用户信息
	tuser, err := l.svcCtx.UserModel.FindOne(l.ctx, in.UserId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if tuser == nil {
		return &user.UpdateProfileResp{
			Base: &common.RespBase{
				Code: 404,
				Msg:  "用户不存在",
			},
		}, nil
	}

	// 更新用户基本信息
	now := time.Now().UnixMilli()
	tuser.Nickname = sql.NullString{String: in.Nickname, Valid: in.Nickname != ""}
	tuser.Avatar = sql.NullString{String: in.Avatar, Valid: in.Avatar != ""}
	tuser.Language = sql.NullString{String: in.Language, Valid: in.Language != ""}
	tuser.Timezone = sql.NullString{String: in.Timezone, Valid: in.Timezone != ""}
	tuser.Signature = sql.NullString{String: in.Signature, Valid: in.Signature != ""}
	tuser.UpdateTimes = now

	err = l.svcCtx.UserModel.Update(l.ctx, tuser)
	if err != nil {
		return nil, err
	}

	// 更新或创建身份信息
	userIdentity, err := l.svcCtx.UserIdentityModel.FindOneByTenantIdUserId(l.ctx, tuser.TenantId, in.UserId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if userIdentity != nil {
		userIdentity.Gender = int64(in.Gender)
		userIdentity.Birthday = sql.NullTime{
			Time:  parseDate(in.Birthday),
			Valid: in.Birthday != "",
		}
		userIdentity.CountryCode = sql.NullString{String: in.CountryCode, Valid: in.CountryCode != ""}
		userIdentity.Province = sql.NullString{String: in.Province, Valid: in.Province != ""}
		userIdentity.City = sql.NullString{String: in.City, Valid: in.City != ""}
		userIdentity.Address = sql.NullString{String: in.Address, Valid: in.Address != ""}
		userIdentity.UpdateTimes = now

		err = l.svcCtx.UserIdentityModel.Update(l.ctx, userIdentity)
		if err != nil {
			return nil, err
		}
	}

	l.Logger.Infof("用户 %d 更新资料成功", in.UserId)

	// 返回更新后的资料
	return l.buildUpdateProfileResp(tuser, userIdentity)
}

func (l *UpdateProfileLogic) buildUpdateProfileResp(tuser *models.TUser, userIdentity *models.TUserIdentity) (*user.UpdateProfileResp, error) {
	tuser2, _ := l.svcCtx.UserModel.FindOne(l.ctx, tuser.Id)
	userIdentity2, _ := l.svcCtx.UserIdentityModel.FindOneByTenantIdUserId(l.ctx, tuser.TenantId, tuser.Id)
	userSecurity, _ := l.svcCtx.UserSecurityModel.FindOneByTenantIdUserId(l.ctx, tuser.TenantId, tuser.Id)

	userBase := &user.UserBase{
		Id:            tuser2.Id,
		TenantId:      tuser2.TenantId,
		UserNo:        tuser2.UserNo,
		Username:      tuser2.Username,
		Nickname:      tuser2.Nickname.String,
		Avatar:        tuser2.Avatar.String,
		Language:      tuser2.Language.String,
		Timezone:      tuser2.Timezone.String,
		InviteCode:    tuser2.InviteCode.String,
		Signature:     tuser2.Signature.String,
		RegisterType:  user.RegisterType(tuser2.RegisterType),
		Status:        user.UserStatus(tuser2.Status),
		MemberLevel:   int32(tuser2.MemberLevel),
		Source:        tuser2.Source.String,
		ReferrerUserId: tuser2.ReferrerUserId.Int64,
		LastLoginIp:   tuser2.LastLoginIp.String,
		LastLoginTime: tuser2.LastLoginTime,
		RegisterIp:    tuser2.RegisterIp.String,
		RegisterTime:  tuser2.RegisterTime,
		Remark:        tuser2.Remark.String,
		Deleted:       tuser2.Deleted == 1,
		CreateTimes:   tuser2.CreateTimes,
		UpdateTimes:   tuser2.UpdateTimes,
	}

	userIdentityProto := &user.UserIdentity{}
	if userIdentity2 != nil {
		userIdentityProto = &user.UserIdentity{
			Id:            userIdentity2.Id,
			TenantId:      userIdentity2.TenantId,
			UserId:        userIdentity2.UserId,
			Phone:         userIdentity2.Phone.String,
			Email:         userIdentity2.Email.String,
			RealName:      userIdentity2.RealName.String,
			Gender:        user.Gender(userIdentity2.Gender),
			Birthday:      userIdentity2.Birthday.Time.Format("2006-01-02"),
			CountryCode:   userIdentity2.CountryCode.String,
			Province:      userIdentity2.Province.String,
			City:          userIdentity2.City.String,
			Address:       userIdentity2.Address.String,
			IdType:        user.IdType(userIdentity2.IdType),
			IdNo:          userIdentity2.IdNo.String,
			FrontImage:    userIdentity2.FrontImage.String,
			BackImage:     userIdentity2.BackImage.String,
			HandheldImage: userIdentity2.HandheldImage.String,
			KycLevel:      user.KycLevel(userIdentity2.KycLevel),
			VerifyStatus:  user.VerifyStatus(userIdentity2.VerifyStatus),
			RejectReason:  userIdentity2.RejectReason.String,
			SubmitTime:    userIdentity2.SubmitTime,
			VerifyTime:    userIdentity2.VerifyTime,
			VerifyBy:      userIdentity2.VerifyBy.Int64,
			CreateTimes:   userIdentity2.CreateTimes,
			UpdateTimes:   userIdentity2.UpdateTimes,
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

	return &user.UpdateProfileResp{
		Base: helper.OkResp(),
		Profile: &user.UserProfile{
			Base:     userBase,
			Identity: userIdentityProto,
			Security: userSecurityProto,
		},
	}, nil
}

func parseDate(dateStr string) time.Time {
	t, _ := time.Parse("2006-01-02", dateStr)
	return t
}