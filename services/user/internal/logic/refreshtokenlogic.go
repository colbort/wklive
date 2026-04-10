package logic

import (
	"context"
	"time"

	"wklive/common/helper"
	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/user"
	"wklive/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type RefreshTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRefreshTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshTokenLogic {
	return &RefreshTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 刷新Token
func (l *RefreshTokenLogic) RefreshToken(in *user.RefreshTokenReq) (*user.RefreshTokenResp, error) {
	// 解析旧token获取用户信息
	claims, err := utils.ParseToken(in.RefreshToken, l.svcCtx.Config.Jwt.AccessSecret)
	if err != nil {
		return &user.RefreshTokenResp{
			Base: helper.GetErrResp(401, "Token已过期或无效"),
		}, nil
	}

	// 获取用户信息
	tuser, err := l.svcCtx.UserModel.FindOne(l.ctx, claims.Uid)
	if err != nil {
		return nil, err
	}

	if tuser == nil {
		return &user.RefreshTokenResp{
			Base: helper.GetErrResp(404, "用户不存在"),
		}, nil
	}

	if tuser.Status != 1 {
		return &user.RefreshTokenResp{
			Base: helper.GetErrResp(403, "账户被禁用"),
		}, nil
	}

	// 生成新的accessToken
	accessToken, err := utils.GenToken(
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

	// 生成新的refreshToken
	refreshToken, err := utils.GenToken(
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

	l.Logger.Infof("用户 %d 刷新Token成功", tuser.Id)

	return &user.RefreshTokenResp{
		Base: helper.OkResp(),
		Token: &common.TokenInfo{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
	}, nil
}
