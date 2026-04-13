package logic

import (
	"context"
	"errors"
	"time"
	"wklive/common/helper"
	"wklive/common/i18n"
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
	tenant, err := l.svcCtx.SystemCli.SysTenantDetail(l.ctx, &system.SysTenantDetailReq{
		TenantCode: &in.TenantCode,
	})
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if tenant == nil || tenant.Base.Code != 200 {
		return &user.LoginResp{
			Base: helper.GetErrResp(401, i18n.Translate(i18n.TenantNotFound, l.ctx)),
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
			Base: helper.GetErrResp(401, i18n.Translate(i18n.UserNotFoundOrPasswordIncorrect, l.ctx)),
		}, nil
	}

	if tuser.Status != 1 {
		return &user.LoginResp{
			Base: helper.GetErrResp(403, i18n.Translate(i18n.AccountDisabled, l.ctx)),
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
		LastLoginTime: utils.NowMillis(),
		UpdateTimes:   utils.NowMillis(),
	})

	return &user.LoginResp{
		Base:   helper.OkResp(),
		UserId: tuser.Id,
		Token: &common.TokenInfo{
			AccessToken: token,
		},
		Profile: toUserProfileProto(tuser, nil, nil),
	}, nil
}
