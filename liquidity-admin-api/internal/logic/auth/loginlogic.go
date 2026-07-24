// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package auth

import (
	"context"

	"wklive/liquidity-admin-api/internal/logicutil"
	"wklive/liquidity-admin-api/internal/svc"
	"wklive/liquidity-admin-api/internal/types"
	"wklive/proto/system"

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

func (l *LoginLogic) Login(req *types.LoginReq, ip, ua string) (resp *types.LoginResp, err error) {
	out, err := l.svcCtx.SystemCli.Login(l.ctx, &system.LoginReq{
		Username: req.Username, Password: req.Password, GoogleCode: req.GoogleCode,
		Ip: ip, Ua: ua, AppScope: system.ApplicationScope_APPLICATION_SCOPE_LIQUIDITY,
	})
	if err != nil {
		return nil, err
	}
	return logicutil.Convert[types.LoginResp](out), nil
}
