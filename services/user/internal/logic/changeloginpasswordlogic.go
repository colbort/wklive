package logic

import (
	"context"
	"errors"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/user"
	"wklive/services/user/internal/svc"
	"wklive/services/user/models"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

type ChangeLoginPasswordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewChangeLoginPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChangeLoginPasswordLogic {
	return &ChangeLoginPasswordLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 修改登录密码
func (l *ChangeLoginPasswordLogic) ChangeLoginPassword(in *user.ChangeLoginPasswordReq) (*user.AppCommonResp, error) {
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
		return &user.AppCommonResp{
			Base: helper.ErrResp(i18n.UserNotFound, i18n.Translate(i18n.UserNotFound, l.ctx)),
		}, nil
	}

	// 验证密码是否一致
	if in.NewPassword != in.ConfirmPassword {
		return &user.AppCommonResp{
			Base: helper.ErrResp(i18n.PasswordsDoNotMatch, i18n.Translate(i18n.PasswordsDoNotMatch, l.ctx)),
		}, nil
	}

	if in.NewPassword == "" {
		return nil, i18n.StatusError(l.ctx, i18n.ParamError)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(in.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// TODO: 验证旧密码是否正确
	// 在实际项目中需要对密码进行验证

	// 更新密码
	tuser.PasswordHash = string(hashedPassword)
	tuser.UpdateTimes = utils.NowMillis()

	err = l.svcCtx.UserModel.Update(l.ctx, tuser)
	if err != nil {
		return nil, err
	}

	l.Logger.Infof("用户 %d 修改登录密码成功", userId)

	return &user.AppCommonResp{
		Base: helper.OkResp(),
	}, nil
}
