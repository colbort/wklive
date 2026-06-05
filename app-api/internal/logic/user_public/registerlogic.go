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

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req *types.RegisterReq) (resp *types.RegisterResp, err error) {
	return logicutil.Proxy[types.RegisterResp](l.ctx, &user.RegisterReq{
		RegisterType:    user.RegisterType(req.RegisterType),
		Username:        req.Username,
		Phone:           req.Phone,
		Email:           req.Email,
		Password:        req.Password,
		ConfirmPassword: req.ConfirmPassword,
		InviteCode:      req.InviteCode,
		Source:          req.Source,
		DeviceId:        req.DeviceId,
		Fingerprint:     req.Fingerprint,
	}, l.svcCtx.UserCli.Register)
}
