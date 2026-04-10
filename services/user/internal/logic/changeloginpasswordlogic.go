package logic

import (
	"context"
	"errors"
	"time"

	"wklive/common/helper"
	"wklive/proto/user"
	"wklive/services/user/internal/svc"
	"wklive/services/user/models"

	"github.com/zeromicro/go-zero/core/logx"
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
	// 获取用户信息
	tuser, err := l.svcCtx.UserModel.FindOne(l.ctx, in.UserId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if tuser == nil {
		return &user.AppCommonResp{
			Base: helper.GetErrResp(404, "用户不存在"),
		}, nil
	}

	// 验证密码是否一致
	if in.NewPassword != in.ConfirmPassword {
		return &user.AppCommonResp{
			Base: helper.GetErrResp(400, "两次密码输入不一致"),
		}, nil
	}

	// TODO: 验证旧密码是否正确
	// 在实际项目中需要对密码进行验证

	// 更新密码
	tuser.PasswordHash = in.NewPassword
	tuser.UpdateTimes = time.Now().UnixMilli()

	err = l.svcCtx.UserModel.Update(l.ctx, tuser)
	if err != nil {
		return nil, err
	}

	l.Logger.Infof("用户 %d 修改登录密码成功", in.UserId)

	return &user.AppCommonResp{
		Base: helper.OkResp(),
	}, nil
}
