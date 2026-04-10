package logic

import (
	"context"
	"errors"
	"time"

	"wklive/common/helper"
	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/system"
	"wklive/proto/user"
	"wklive/services/user/internal/svc"
	"wklive/services/user/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 用户登录
func (l *LoginLogic) Login(in *user.LoginReq) (*user.LoginResp, error) {
	tenant, err := l.svcCtx.SystemCli.SysTenantByCode(l.ctx, &system.SysTenantByCodeReq{
		TenantCode: in.TenantCode,
	})
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if tenant == nil || tenant.Base.Code != 200 {
		return &user.LoginResp{
			Base: helper.GetErrResp(401, "租户不存在"),
		}, nil
	}

	var tuser *models.TUser
	switch in.LoginType {
	case user.LoginType_LOGIN_TYPE_EMAIL:
		identity, err := l.svcCtx.UserIdentityModel.FindByEmail(l.ctx, tenant.Data.Id, in.Account)
		if err != nil && !errors.Is(err, models.ErrNotFound) {
			return nil, err
		}
		if identity != nil {
			tuser, err = l.svcCtx.UserModel.FindOne(l.ctx, identity.UserId)
			if err != nil && !errors.Is(err, models.ErrNotFound) {
				return nil, err
			}
		}
	case user.LoginType_LOGIN_TYPE_PHONE:
		identity, err := l.svcCtx.UserIdentityModel.FindByPhone(l.ctx, tenant.Data.Id, in.Account)
		if err != nil && !errors.Is(err, models.ErrNotFound) {
			return nil, err
		}
		if identity != nil {
			tuser, err = l.svcCtx.UserModel.FindOne(l.ctx, identity.UserId)
			if err != nil && !errors.Is(err, models.ErrNotFound) {
				return nil, err
			}
		}
	case user.LoginType_LOGIN_TYPE_USERNAME:
		tuser, err = l.svcCtx.UserModel.FindByUsername(l.ctx, in.TenantCode, in.Account)
		if err != nil && !errors.Is(err, models.ErrNotFound) {
			return nil, err
		}
	}

	if tuser == nil {
		return &user.LoginResp{
			Base: helper.GetErrResp(401, "用户不存在或密码错误"),
		}, nil
	}

	if tuser.Status != 1 {
		return &user.LoginResp{
			Base: helper.GetErrResp(403, "账户被禁用"),
		}, nil
	}

	// TODO: verify password
	// TODO: verify google 2FA if enabled

	token, err := utils.GenToken(
		l.svcCtx.Config.Jwt.AccessSecret,
		tuser.Id,
		tuser.Username,
		0,
		"",
		time.Duration(l.svcCtx.Config.Jwt.AccessExpire)*time.Second,
	)
	if err != nil {
		return nil, err
	}

	_ = l.svcCtx.UserModel.Update(l.ctx, &models.TUser{
		Id:            tuser.Id,
		LastLoginIp:   tuser.LastLoginIp,
		LastLoginTime: time.Now().UnixMilli(),
		UpdateTimes:   time.Now().UnixMilli(),
	})

	// userIdentity, _ := l.svcCtx.UserIdentityModel.FindOne(l.ctx, tuser.Id)
	// userSecurity, _ := l.svcCtx.UserSecurityModel.FindOne(l.ctx, tuser.Id)

	return &user.LoginResp{
		Base:   helper.OkResp(),
		UserId: tuser.Id,
		Token: &common.TokenInfo{
			AccessToken: token,
		},
		Profile: &user.UserProfile{
			Base: &user.UserBase{
				Id:            tuser.Id,
				TenantId:      tuser.TenantId,
				Username:      tuser.Username,
				Nickname:      tuser.Nickname.String,
				Avatar:        tuser.Avatar.String,
				Status:        user.UserStatus(tuser.Status),
				MemberLevel:   int32(tuser.MemberLevel),
				RegisterTime:  tuser.CreateTimes,
				LastLoginTime: tuser.LastLoginTime,
			},
			Identity: &user.UserIdentity{},
			Security: &user.UserSecurity{},
		},
	}, nil
}
