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
)

type InitGoogle2FALogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewInitGoogle2FALogic(ctx context.Context, svcCtx *svc.ServiceContext) *InitGoogle2FALogic {
	return &InitGoogle2FALogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 初始化Google 2FA
func (l *InitGoogle2FALogic) InitGoogle2FA(in *user.InitGoogle2FAReq) (*user.InitGoogle2FAResp, error) {
	userId, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	// 获取用户信息
	tuser, err := l.svcCtx.UserModel.FindOne(l.ctx, userId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if tuser == nil {
		return &user.InitGoogle2FAResp{
			Base: helper.GetErrResp(i18n.UserNotFound, i18n.Translate(i18n.UserNotFound, l.ctx)),
		}, nil
	}

	// 生成Google 2FA密钥
	secret, _, qrCodeUrl, err := utils.GenerateGoogle2FA(tuser.Username, "AVE", 100)
	if err != nil {
		return nil, err
	}
	if secret == "" {
		return &user.InitGoogle2FAResp{
			Base: helper.GetErrResp(i18n.SecretGenerationFailed, i18n.Translate(i18n.SecretGenerationFailed, l.ctx)),
		}, nil
	}

	now := utils.NowMillis()
	userSecurity, err := l.svcCtx.UserSecurityModel.FindOneByTenantIdUserId(l.ctx, tuser.TenantId, userId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if userSecurity != nil {
		userSecurity.GoogleSecret = sql.NullString{String: secret, Valid: true}
		userSecurity.UpdateTimes = now

		err = l.svcCtx.UserSecurityModel.Update(l.ctx, userSecurity)
		if err != nil {
			return nil, err
		}
	} else {
		userSecurity = &models.TUserSecurity{
			Id:           l.svcCtx.Node.Generate().Int64(),
			TenantId:     tuser.TenantId,
			UserId:       userId,
			GoogleSecret: sql.NullString{String: secret, Valid: true},
			CreateTimes:  now,
			UpdateTimes:  now,
		}

		_, err = l.svcCtx.UserSecurityModel.Insert(l.ctx, userSecurity)
		if err != nil {
			return nil, err
		}
	}

	l.Logger.Infof("用户 %d 初始化Google 2FA成功", userId)

	return &user.InitGoogle2FAResp{
		Base:      helper.OkResp(),
		Secret:    secret,
		QrCodeUrl: qrCodeUrl,
	}, nil
}
