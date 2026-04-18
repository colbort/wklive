// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package auth_public

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
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

func (l *LoginLogic) Login(req *types.LoginReq, ip string) (resp *types.LoginResp, err error) {
	result, err := l.svcCtx.SystemCli.AdminLogin(l.ctx, &system.AdminLoginReq{
		Username:   req.Username,
		Password:   req.Password,
		GoogleCode: req.GoogleCode,
		Ip:         ip,
		Ua:         "",
	})
	if err != nil {
		return nil, err
	}

	resp = new(types.LoginResp)
	resp.Code = result.Base.Code
	resp.Msg = result.Base.Msg
	resp.Data.Token = result.Token
	resp.Data.Exp = result.Exp
	return
}
