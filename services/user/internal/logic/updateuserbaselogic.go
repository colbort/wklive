package logic

import (
	"context"
	"database/sql"
	"errors"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/user"
	"wklive/services/user/internal/svc"
	"wklive/services/user/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type UpdateUserBaseLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateUserBaseLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserBaseLogic {
	return &UpdateUserBaseLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 管理员更新用户基本信息
func (l *UpdateUserBaseLogic) UpdateUserBase(in *user.UpdateUserBaseReq) (*user.UpdateUserBaseResp, error) {
	// 获取用户信息
	tuser, err := l.svcCtx.UserModel.FindOne(l.ctx, in.UserId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if tuser == nil {
		return &user.UpdateUserBaseResp{
			Base: helper.GetErrResp(404, i18n.Translate(i18n.UserNotFound, l.ctx)),
		}, nil
	}

	now := utils.NowMillis()

	// 更新用户基本信息
	if in.Username != "" {
		tuser.Username = in.Username
	}
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
	if in.Source != "" {
		tuser.Source = sql.NullString{String: in.Source, Valid: true}
	}
	if in.ReferrerUserId > 0 {
		tuser.ReferrerUserId = sql.NullInt64{Int64: in.ReferrerUserId, Valid: true}
	}
	if in.Remark != "" {
		tuser.Remark = sql.NullString{String: in.Remark, Valid: true}
	}

	tuser.UpdateTimes = now

	userIdentity, err := l.svcCtx.UserIdentityModel.FindOneByTenantIdUserId(l.ctx, in.TenantId, in.UserId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if userIdentity != nil {
		if in.Phone != "" {
			userIdentity.Phone = sql.NullString{String: in.Phone, Valid: true}
		}
		if in.Email != "" {
			userIdentity.Email = sql.NullString{String: in.Email, Valid: true}
		}
		userIdentity.UpdateTimes = now

	}

	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		userModel := models.NewTUserModel(conn, l.svcCtx.Config.CacheRedis).(models.UserModel)
		userIdentityModel := models.NewTUserIdentityModel(conn, l.svcCtx.Config.CacheRedis).(models.UserIdentityModel)

		if err := userModel.Update(ctx, tuser); err != nil {
			return err
		}
		if userIdentity != nil {
			if err := userIdentityModel.Update(ctx, userIdentity); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	l.Logger.Infof("管理员更新用户 %d 基本信息成功", in.UserId)

	// 返回完整用户详情
	userSecurity, _ := l.svcCtx.UserSecurityModel.FindOneByTenantIdUserId(l.ctx, in.TenantId, in.UserId)

	userDetailResp := buildUserDetail(tuser, userIdentity, userSecurity)

	return &user.UpdateUserBaseResp{
		Base:   helper.OkResp(),
		Detail: userDetailResp,
	}, nil
}

func buildUserDetail(tuser *models.TUser, userIdentity *models.TUserIdentity, userSecurity *models.TUserSecurity) *user.UserDetail {
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

	return &user.UserDetail{
		Base:     userBase,
		Identity: userIdentityProto,
		Security: userSecurityProto,
		Banks:    []*user.UserBankItem{},
	}
}
