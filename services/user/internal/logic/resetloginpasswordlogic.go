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
)

type ResetLoginPasswordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewResetLoginPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetLoginPasswordLogic {
	return &ResetLoginPasswordLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 管理员重置登录密码
func (l *ResetLoginPasswordLogic) ResetLoginPassword(in *user.ResetLoginPasswordReq) (*user.AdminCommonResp, error) {
	// 获取用户信息
	tuser, err := l.svcCtx.UserModel.FindOne(l.ctx, in.UserId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if tuser == nil {
		return &user.AdminCommonResp{
			Base: helper.GetErrResp(404, i18n.Translate(i18n.UserNotFound, l.ctx)),
		}, nil
	}

	// 重置登录密码
	tuser.PasswordHash = in.NewPassword
	tuser.UpdateTimes = utils.NowMillis()

	err = l.svcCtx.UserModel.Update(l.ctx, tuser)
	if err != nil {
		return nil, err
	}

	l.Logger.Infof("管理员为用户 %d 重置登录密码成功", in.UserId)

	return &user.AdminCommonResp{
		Base: helper.OkResp(),
	}, nil
}
