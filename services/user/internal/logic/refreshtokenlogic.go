package logic

import (
	"context"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
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
	claims, err := utils.ParseToken(l.svcCtx.Config.Jwt.AccessSecret, in.RefreshToken)
	if err != nil {
		return &user.RefreshTokenResp{
			Base: helper.GetErrResp(i18n.TokenExpiredOrInvalid, i18n.Translate(i18n.TokenExpiredOrInvalid, l.ctx)),
		}, nil
	}

	// 获取用户信息
	tuser, err := l.svcCtx.UserModel.FindOne(l.ctx, claims.UserId)
	if err != nil {
		return nil, err
	}

	if tuser == nil {
		return &user.RefreshTokenResp{
			Base: helper.GetErrResp(i18n.UserNotFound, i18n.Translate(i18n.UserNotFound, l.ctx)),
		}, nil
	}

	if tuser.Status != 1 {
		return &user.RefreshTokenResp{
			Base: helper.GetErrResp(i18n.AccountDisabled, i18n.Translate(i18n.AccountDisabled, l.ctx)),
		}, nil
	}

	token, err := buildTokenInfo(
		l.svcCtx.Config.Jwt.AccessSecret,
		l.svcCtx.Config.Jwt.AccessExpire,
		tuser.Id, tuser.Username, claims.Expand,
	)
	if err != nil {
		return nil, err
	}

	l.Logger.Infof("用户 %d 刷新Token成功", tuser.Id)

	return &user.RefreshTokenResp{
		Base: helper.OkResp(),
		Data: token,
	}, nil
}
