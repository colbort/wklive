// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user_public

import (
	"context"

	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"
	"wklive/proto/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginReq) (resp *types.LoginResp, err error) {
	result, err := l.svcCtx.UserCli.Login(l.ctx, &user.LoginReq{
		TenantCode: req.TenantCode,
		LoginType:  user.LoginType(req.LoginType),
		Account:    req.Account,
		Password:   req.Password,
		GoogleCode: req.GoogleCode,
		LoginIp:    "",
	})
	if err != nil {
		return nil, err
	}

	resp = &types.LoginResp{
		RespBase: types.RespBase{
			Code: result.Base.Code,
			Msg:  result.Base.Msg,
		},
		UserId: result.UserId,
		Token: types.TokenInfo{
			AccessToken:  result.Token.AccessToken,
			RefreshToken: result.Token.RefreshToken,
		},
		Profile: types.UserProfile{
			Identity: types.UserIdentity{},
			Security: types.UserSecurity{},
		},
	}
	return
}
