package logic

import (
	"context"
	"errors"
	"time"

	"wklive/common/helper"
	"wklive/common/utils"
	"wklive/proto/user"
	"wklive/services/user/internal/svc"
	"wklive/services/user/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type EnableGoogle2FALogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewEnableGoogle2FALogic(ctx context.Context, svcCtx *svc.ServiceContext) *EnableGoogle2FALogic {
	return &EnableGoogle2FALogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 启用Google 2FA
func (l *EnableGoogle2FALogic) EnableGoogle2FA(in *user.EnableGoogle2FAReq) (*user.AppCommonResp, error) {
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

	// 获取用户安全信息
	userSecurity, err := l.svcCtx.UserSecurityModel.FindOneByTenantIdUserId(l.ctx, tuser.TenantId, in.UserId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if userSecurity == nil {
		return &user.AppCommonResp{
			Base: helper.GetErrResp(404, "安全设置不存在"),
		}, nil
	}

	// 验证Google 2FA code
	if userSecurity.GoogleSecret.String == "" {
		return &user.AppCommonResp{
			Base: helper.GetErrResp(400, "请先初始化Google 2FA"),
		}, nil
	}

	if !utils.VerifyGoogle2FACode(userSecurity.GoogleSecret.String, in.GoogleCode) {
		return &user.AppCommonResp{
			Base: helper.GetErrResp(400, "验证码错误"),
		}, nil
	}

	// 启用Google 2FA
	userSecurity.GoogleEnabled = 1
	userSecurity.UpdateTimes = time.Now().UnixMilli()

	err = l.svcCtx.UserSecurityModel.Update(l.ctx, userSecurity)
	if err != nil {
		return nil, err
	}

	l.Logger.Infof("用户 %d 启用Google 2FA成功", in.UserId)

	return &user.AppCommonResp{
		Base: helper.OkResp(),
	}, nil
}
