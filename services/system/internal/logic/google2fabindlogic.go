package logic

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"
)

type Google2FABindLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGoogle2FABindLogic(ctx context.Context, svcCtx *svc.ServiceContext) *Google2FABindLogic {
	return &Google2FABindLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 绑定Google 2FA
func (l *Google2FABindLogic) Google2FABind(in *system.Google2FABindReq) (*system.RespBase, error) {
	user, err := l.svcCtx.UserModel.FindOne(l.ctx, in.UserId)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return &system.RespBase{
			Base: helper.GetErrResp(1, i18n.Translate(i18n.UserNotFound, l.ctx)),
		}, nil
	}
	// 从 Redis 获取之前存储的 secret
	secret, err := l.svcCtx.UserModel.GetGoogle2FASecret(l.ctx, user.Id)
	if err != nil {
		return &system.RespBase{
			Base: helper.GetErrResp(1, i18n.Translate(i18n.Google2FASecretFetchFailed, l.ctx)+": "+err.Error()),
		}, err
	}

	if secret == "" {
		return &system.RespBase{
			Base: helper.GetErrResp(1, i18n.Translate(i18n.Google2FASecretExpired, l.ctx)),
		}, nil
	}

	if !utils.VerifyGoogle2FACode(secret, in.Code) {
		return &system.RespBase{
			Base: helper.GetErrResp(1, i18n.Translate(i18n.VerificationCodeInvalid, l.ctx)),
		}, nil
	}

	user.GoogleSecret = secret
	user.GoogleEnabled = 1
	if err = l.svcCtx.UserModel.Update(l.ctx, user); err != nil {
		return nil, err
	}

	// 绑定成功后可以选择删除 Redis 中的 secret
	if err := l.svcCtx.UserModel.DeleteGoogle2FASecret(l.ctx, user.Id); err != nil {
		logx.Errorf("删除2FA secret失败: %v", err)
	}

	return &system.RespBase{
		Base: helper.OkResp(),
	}, nil
}
