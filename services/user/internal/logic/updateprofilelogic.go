package logic

import (
	"context"
	"database/sql"
	"errors"
	"time"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/user"
	"wklive/services/user/internal/svc"
	"wklive/services/user/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
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
			Base: helper.GetErrResp(404, i18n.Translate(i18n.UserNotFound, l.ctx)),
		}, nil
	}

	// 更新用户基本信息
	now := utils.NowMillis()
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

	identity, err := l.svcCtx.UserIdentityModel.FindOneByTenantIdUserId(l.ctx, tuser.TenantId, in.UserId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if identity != nil {
		if in.Gender != 0 {
			identity.Gender = int64(in.Gender)
		}
		if in.Birthday != 0 {
			identity.Birthday = in.Birthday
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

	}

	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		userModel := models.NewTUserModel(conn, l.svcCtx.Config.CacheRedis)
		userIdentityModel := models.NewTUserIdentityModel(conn, l.svcCtx.Config.CacheRedis)

		if err := userModel.Update(ctx, tuser); err != nil {
			return err
		}
		if identity != nil {
			if err := userIdentityModel.Update(ctx, identity); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	l.Logger.Infof("用户 %d 更新资料成功", in.UserId)

	// 返回更新后的资料
	return l.buildUpdateProfileResp(tuser, identity)
}

func (l *UpdateProfileLogic) buildUpdateProfileResp(tuser *models.TUser, _ *models.TUserIdentity) (*user.UpdateProfileResp, error) {
	identity, err := l.svcCtx.UserIdentityModel.FindOneByTenantIdUserId(l.ctx, tuser.TenantId, tuser.Id)
	if err != nil {
		return nil, err
	}
	security, err := l.svcCtx.UserSecurityModel.FindOneByTenantIdUserId(l.ctx, tuser.TenantId, tuser.Id)
	if err != nil {
		return nil, err
	}

	return &user.UpdateProfileResp{
		Base: helper.OkResp(),
		Profile: toUserProfileProto(tuser, identity, security),
	}, nil
}

func parseDate(dateStr string) time.Time {
	t, _ := time.Parse("2006-01-02", dateStr)
	return t
}
