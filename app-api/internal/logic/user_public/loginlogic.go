// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user_public

import (
	"context"

	"wklive/app-api/internal/logicutil"
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

func (l *LoginLogic) Login(req *types.LoginReq) (*types.LoginResp, error) {
	result, err := l.svcCtx.UserCli.Login(l.ctx, &user.LoginReq{
		LoginType:  user.LoginType(req.LoginType),
		Account:    req.Account,
		Password:   req.Password,
		GoogleCode: req.GoogleCode,
	})
	if err != nil {
		return logicutil.SystemErrorResp[types.LoginResp](l.ctx, err)
	}

	if result.Base.Code != 200 {
		return &types.LoginResp{
			RespBase: types.RespBase{
				Code: result.Base.Code,
				Msg:  result.Base.Msg,
			},
		}, nil
	} else {
		data := result.Data
		token := data.GetToken()
		return &types.LoginResp{
			RespBase: types.RespBase{
				Code: result.Base.Code,
				Msg:  result.Base.Msg,
			},
			Data: types.LoginData{
				UserId: data.GetUserId(),
				Token: types.TokenInfo{
					AccessToken:  token.GetAccessToken(),
					RefreshToken: token.GetRefreshToken(),
					ExpireTime:   token.GetExpireTime(),
				},
				Profile: types.UserProfile{
					Identity: types.UserIdentity{},
					Security: types.UserSecurity{},
				},
			},
		}, nil
	}
}
