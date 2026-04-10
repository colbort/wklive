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
	if in.Nickname != "" {
		tuser.Nickname = sql.NullString{String: in.Nickname, Valid: true}
	}
	if in.Avatar != "" {
		tuser.Avatar = sql.NullString{String: in.Avatar, Valid: true}
	}
	if in.Language != "" {
		tuser.Language = sql.NullString{String: in.Language, Valid: true}
	}
	if in.Timezone != "" {
		tuser.Timezone = sql.NullString{String: in.Timezone, Valid: true}
	}
	if in.Signature != "" {
		tuser.Signature = sql.NullString{String: in.Signature, Valid: true}
	}
	tuser.UpdateTimes = now

	err = l.svcCtx.UserModel.Update(l.ctx, tuser)
	if err != nil {
		return nil, err
	}

	// 更新或创建身份信息
	identity, err := l.svcCtx.UserIdentityModel.FindOneByTenantIdUserId(l.ctx, tuser.TenantId, in.UserId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if identity != nil {
		if in.Gender != 0 {
			identity.Gender = int64(in.Gender)
		}
		if in.Birthday != "" {
			identity.Birthday = sql.NullTime{
				Time:  parseDate(in.Birthday),
				Valid: true,
			}
		}
		if in.CountryCode != "" {
			identity.CountryCode = sql.NullString{String: in.CountryCode, Valid: true}
		}
		if in.Province != "" {
			identity.Province = sql.NullString{String: in.Province, Valid: true}
		}
		if in.City != "" {
			identity.City = sql.NullString{String: in.City, Valid: true}
		}
		if in.Address != "" {
			identity.Address = sql.NullString{String: in.Address, Valid: true}
		}
		identity.UpdateTimes = now

		err = l.svcCtx.UserIdentityModel.Update(l.ctx, identity)
		if err != nil {
			return nil, err
		}
	}

	l.Logger.Infof("用户 %d 更新资料成功", in.UserId)

	// 返回更新后的资料
	return l.buildUpdateProfileResp(tuser, identity)
}

func (l *UpdateProfileLogic) buildUpdateProfileResp(tuser *models.TUser, _ *models.TUserIdentity) (*user.UpdateProfileResp, error) {
	identity, _ := l.svcCtx.UserIdentityModel.FindOneByTenantIdUserId(l.ctx, tuser.TenantId, tuser.Id)
	security, _ := l.svcCtx.UserSecurityModel.FindOneByTenantIdUserId(l.ctx, tuser.TenantId, tuser.Id)

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
	if identity != nil {
		userIdentityProto = &user.UserIdentity{
			Id:            identity.Id,
			TenantId:      identity.TenantId,
			UserId:        identity.UserId,
			Phone:         identity.Phone.String,
			Email:         identity.Email.String,
			RealName:      identity.RealName.String,
			Gender:        user.Gender(identity.Gender),
			Birthday:      identity.Birthday.Time.Format("2006-01-02"),
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

	userSecurityProto := &user.UserSecurity{}
	if security != nil {
		userSecurityProto = &user.UserSecurity{
			Id:              security.Id,
			TenantId:        security.TenantId,
			UserId:          security.UserId,
			HasPayPassword:  security.PayPasswordHash.Valid && security.PayPasswordHash.String != "",
			GoogleEnabled:   security.GoogleEnabled == 1,
			LoginErrorCount: security.LoginErrorCount,
			PayErrorCount:   security.PayErrorCount,
			LockUntil:       security.LockUntil,
			RiskLevel:       user.RiskLevel(security.RiskLevel),
			CreateTimes:     security.CreateTimes,
			UpdateTimes:     security.UpdateTimes,
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
