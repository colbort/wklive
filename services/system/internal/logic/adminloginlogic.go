package logic

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"wklive/common/utils"
	"wklive/rpc/system"
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
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}

	// 2️⃣ 校验密码（bcrypt）
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(in.Password)) != nil {
		return nil, errors.New("密码错误")
	}

	// 3️⃣ 更新登录时间
	user.LastLoginIp = sql.NullString{String: in.Ip, Valid: true}
	now := time.Now()
	user.LastLoginAt = sql.NullTime{Time: now, Valid: true}
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
		Token:            token,
		Uid:              user.Id,
		Nickname:         user.Nickname,
		Google2FaEnabled: int32(user.GoogleEnabled),
		PermsVer:         user.PermsVer,
	}, nil
}
