package logic

import (
	"context"
	"database/sql"
	"errors"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
	"time"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"
	"wklive/services/system/models"
)

type AdminLoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminLoginLogic {
	return &AdminLoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// P0
func (l *AdminLoginLogic) AdminLogin(in *system.AdminLoginReq) (*system.AdminLoginResp, error) {
	// 1️⃣ 查用户
	user, err := l.svcCtx.UserModel.FindOneByUsername(l.ctx, in.Username)
	if err != nil {
		if err == models.ErrNotFound {
			return nil, errors.New(i18n.Translate(i18n.UserNotFound, l.ctx))
		}
		return nil, err
	}

	if user == nil {
		return nil, errors.New(i18n.Translate(i18n.UserNotFound, l.ctx))
	}

	if user.Status != 1 {
		return nil, errors.New(i18n.Translate(i18n.UserDisabledForLogin, l.ctx))
	}

	if user.GoogleEnabled == 1 {
		if in.GoogleCode == "" {
			return nil, errors.New(i18n.Translate(i18n.Google2FACodeRequired, l.ctx))
		}
		if user.GoogleSecret == "" || !utils.VerifyGoogle2FACode(user.GoogleSecret, in.GoogleCode) {
			return nil, errors.New(i18n.Translate(i18n.Google2FACodeInvalid, l.ctx))
		}
	}

	// 2️⃣ 校验密码（bcrypt）
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(in.Password)) != nil {
		return nil, errors.New(i18n.Translate(i18n.PasswordIncorrect, l.ctx))
	}

	// 3️⃣ 更新登录时间
	user.LastLoginIp = sql.NullString{String: in.Ip, Valid: true}
	now := time.Now().UnixMilli()
	user.LastLoginAt = now
	_ = l.svcCtx.UserModel.Update(l.ctx, user)

	// 4️⃣ 写登录日志
	_, _ = l.svcCtx.LoginLogModel.Insert(l.ctx, &models.SysLoginLog{
		UserId:   sql.NullInt64{Int64: user.Id, Valid: true},
		Username: sql.NullString{String: user.Username, Valid: true},
		Ip:       sql.NullString{String: in.Ip, Valid: true},
		Ua:       sql.NullString{String: in.Ua, Valid: true},
		Success:  sql.NullInt64{Int64: 1, Valid: true},
		Msg:      sql.NullString{String: "登录成功", Valid: true},
		LoginAt:  now,
	})

	token, err := utils.GenToken(
		l.svcCtx.Config.Jwt.AccessSecret,
		user.Id,
		user.Username,
		user.PermsVer,
		l.svcCtx.Config.Name,
		time.Duration(l.svcCtx.Config.Jwt.AccessExpire)*time.Second,
	)
	if err != nil {
		return nil, err
	}

	return &system.AdminLoginResp{
		Base:             helper.OkResp(),
		Token:            token,
		Uid:              user.Id,
		Nickname:         user.Nickname,
		Google2FaEnabled: user.GoogleEnabled,
		PermsVer:         user.PermsVer,
	}, nil
}
