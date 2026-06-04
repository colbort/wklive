package logic

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/system"
	"wklive/proto/user"
	"wklive/services/user/internal/svc"
	"wklive/services/user/models"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
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
	loginIP, _ := utils.GetClientIPFromMd(l.ctx)
	tenant, err := l.svcCtx.SystemCli.SysTenantDetail(l.ctx, &system.SysTenantDetailReq{
		TenantCode: &in.TenantCode,
	})
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if tenant == nil || tenant.Base.Code != 200 {
		return &user.LoginResp{
			Base: helper.GetErrResp(i18n.TenantNotFound, i18n.Translate(i18n.TenantNotFound, l.ctx)),
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
			Base: helper.GetErrResp(i18n.UserNotFoundOrPasswordIncorrect, i18n.Translate(i18n.UserNotFoundOrPasswordIncorrect, l.ctx)),
		}, nil
	}

	if tuser.Status != 1 {
		return &user.LoginResp{
			Base: helper.GetErrResp(i18n.AccountDisabled, i18n.Translate(i18n.AccountDisabled, l.ctx)),
		}, nil
	}

	if bcrypt.CompareHashAndPassword([]byte(tuser.PasswordHash), []byte(in.Password)) != nil {
		return &user.LoginResp{
			Base: helper.GetErrResp(i18n.UserNotFoundOrPasswordIncorrect, i18n.Translate(i18n.UserNotFoundOrPasswordIncorrect, l.ctx)),
		}, nil
	}

	userSecurity, err := l.svcCtx.UserSecurityModel.FindOneByTenantIdUserId(l.ctx, tenant.Data.Id, tuser.Id)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if userSecurity != nil && userSecurity.GoogleEnabled == 1 {
		if in.GoogleCode == "" {
			return &user.LoginResp{
				Base: helper.GetErrResp(i18n.Google2FACodeRequired, i18n.Translate(i18n.Google2FACodeRequired, l.ctx)),
			}, nil
		}
		if !utils.VerifyGoogle2FACode(userSecurity.GoogleSecret.String, in.GoogleCode) {
			return &user.LoginResp{
				Base: helper.GetErrResp(i18n.Google2FACodeInvalid, i18n.Translate(i18n.Google2FACodeInvalid, l.ctx)),
			}, nil
		}
	}

	str := make(map[string]any, 0)
	str["tid"] = tenant.Data.Id
	expand, err := json.Marshal(str)
	if err != nil {
		return nil, err
	}
	token, err := buildTokenInfo(
		l.svcCtx.Config.Jwt.AccessSecret,
		l.svcCtx.Config.Jwt.AccessExpire,
		tuser.Id, tuser.Username, string(expand),
	)
	if err != nil {
		return nil, err
	}

	now := utils.NowMillis()
	if loginIP != "" {
		tuser.LastLoginIp = sql.NullString{String: loginIP, Valid: true}
	}
	tuser.LastLoginTime = now
	tuser.UpdateTimes = now
	_ = l.svcCtx.UserModel.Update(l.ctx, tuser)

	return &user.LoginResp{
		Base:    helper.OkResp(),
		UserId:  tuser.Id,
		Token:   token,
		Profile: toUserProfileProto(tuser, nil, userSecurity),
	}, nil
}
