package logic

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"
	"wklive/services/system/models"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
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
			return nil, i18n.StatusError(l.ctx, i18n.UserNotFound)
		}
		return nil, err
	}

	if user == nil {
		return nil, i18n.StatusError(l.ctx, i18n.UserNotFound)
	}

	if user.Status != 1 {
		return nil, i18n.StatusError(l.ctx, i18n.UserDisabledForLogin)
	}

	if user.GoogleEnabled == 1 {
		if in.GoogleCode == "" {
			return nil, i18n.StatusError(l.ctx, i18n.Google2FACodeRequired)
		}
		if user.GoogleSecret == "" || !utils.VerifyGoogle2FACode(user.GoogleSecret, in.GoogleCode) {
			return nil, i18n.StatusError(l.ctx, i18n.Google2FACodeInvalid)
		}
	}

	// 2️⃣ 校验密码（bcrypt）
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(in.Password)) != nil {
		return nil, i18n.StatusError(l.ctx, i18n.PasswordIncorrect)
	}

	// 3️⃣ 更新登录时间
	user.LastLoginIp = in.Ip
	now := utils.NowMillis()
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

	str := make(map[string]any, 0)
	str["permsVer"] = user.PermsVer
	expand, err := json.Marshal(str)
	if err != nil {
		return nil, err
	}

	token, err := utils.GenToken(
		l.svcCtx.Config.Jwt.AccessSecret,
		user.Id,
		user.Username,
		string(expand),
		l.svcCtx.Config.Name,
		time.Duration(l.svcCtx.Config.Jwt.AccessExpire)*time.Second,
	)
	if err != nil {
		return nil, err
	}

	return &system.AdminLoginResp{
		Base: helper.OkResp(),
		Data: &system.AdminLoginData{
			Token:            token,
			UserId:           user.Id,
			Nickname:         user.Nickname,
			Google2FaEnabled: user.GoogleEnabled,
			PermsVer:         user.PermsVer,
		},
	}, nil
}
