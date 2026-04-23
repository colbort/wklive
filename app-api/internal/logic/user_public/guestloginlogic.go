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

type GuestLoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGuestLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GuestLoginLogic {
	return &GuestLoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GuestLoginLogic) GuestLogin(req *types.GuestLoginReq) (resp *types.GuestLoginResp, err error) {
	result, err := l.svcCtx.UserCli.GuestLogin(l.ctx, &user.GuestLoginReq{
		DeviceId:    req.DeviceId,
		Fingerprint: req.Fingerprint,
		RegisterIp:  req.RegisterIp,
		TenantCode:  req.TenantCode,
	})
	if err != nil {
		return nil, err
	}
	data := types.GuestLogin{}
	if result.Data != nil {
		data = types.GuestLogin{
			Token:    result.Data.Token,
			Uid:      result.Data.Uid,
			DeviceId: result.Data.Uid,
			Username: result.Data.Username,
			IsNew:    result.Data.IsNew,
		}
	}
	return &types.GuestLoginResp{
		RespBase: types.RespBase{
			Code: result.Base.Code,
			Msg:  result.Base.Msg,
		},
		Data: data,
	}, nil
}
