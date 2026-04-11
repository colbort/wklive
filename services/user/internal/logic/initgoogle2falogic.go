package logic

import (
	"context"
	"errors"
	"github.com/zeromicro/go-zero/core/logx"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/user"
	"wklive/services/user/internal/svc"
	"wklive/services/user/models"
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
	// 获取用户信息
	tuser, err := l.svcCtx.UserModel.FindOne(l.ctx, in.UserId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if tuser == nil {
		return &user.InitGoogle2FAResp{
			Base: helper.GetErrResp(404, i18n.Translate(i18n.UserNotFound, l.ctx)),
		}, nil
	}

	// 生成Google 2FA密钥
	secret, _, qrCodeUrl, err := utils.GenerateGoogle2FA(tuser.Username, "", 100)
	if err != nil {
		return nil, err
	}
	if secret == "" {
		return &user.InitGoogle2FAResp{
			Base: helper.GetErrResp(500, i18n.Translate(i18n.SecretGenerationFailed, l.ctx)),
		}, nil
	}

	l.Logger.Infof("用户 %d 初始化Google 2FA成功", in.UserId)

	return &user.InitGoogle2FAResp{
		Base:      helper.OkResp(),
		Secret:    secret,
		QrCodeUrl: qrCodeUrl,
	}, nil
}
