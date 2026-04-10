package logic

import (
	"context"
	"errors"
	"time"

	"wklive/common/helper"
	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/user"
	"wklive/services/user/internal/svc"
	"wklive/services/user/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type DisableGoogle2FALogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDisableGoogle2FALogic(ctx context.Context, svcCtx *svc.ServiceContext) *DisableGoogle2FALogic {
	return &DisableGoogle2FALogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 禁用Google 2FA
func (l *DisableGoogle2FALogic) DisableGoogle2FA(in *user.DisableGoogle2FAReq) (*user.AppCommonResp, error) {
	// 获取用户信息
	tuser, err := l.svcCtx.UserModel.FindOne(l.ctx, in.UserId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if tuser == nil {
		return &user.AppCommonResp{
			Base: &common.RespBase{
				Code: 404,
				Msg:  "用户不存在",
			},
		}, nil
	}

	// 获取用户安全信息
	userSecurity, err := l.svcCtx.UserSecurityModel.FindOneByTenantIdUserId(l.ctx, tuser.TenantId, in.UserId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if userSecurity == nil {
		return &user.AppCommonResp{
			Base: &common.RespBase{
				Code: 404,
				Msg:  "安全设置不存在",
			},
		}, nil
	}

	// 验证Google 2FA code
	if !utils.VerifyGoogle2FACode(userSecurity.GoogleSecret.String, in.GoogleCode) {
		return &user.AppCommonResp{
			Base: &common.RespBase{
				Code: 400,
				Msg:  "验证码错误",
			},
		}, nil
	}

	// 禁用Google 2FA
	userSecurity.GoogleEnabled = 0
	userSecurity.UpdateTimes = time.Now().UnixMilli()

	err = l.svcCtx.UserSecurityModel.Update(l.ctx, userSecurity)
	if err != nil {
		return nil, err
	}

	l.Logger.Infof("用户 %d 禁用Google 2FA成功", in.UserId)

	return &user.AppCommonResp{
		Base: helper.OkResp(),
	}, nil
}