package logic

import (
	"context"

	"wklive/proto/system"
	"wklive/services/system/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginUserPermsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginUserPermsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginUserPermsLogic {
	return &LoginUserPermsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取登录用户的权限列表
func (l *LoginUserPermsLogic) LoginUserPerms(in *system.LoginUserPermsReq) (*system.LoginUserPermsResp, error) {
	perms, err := l.svcCtx.UserRoleModel.FindLoginUserPerms(l.ctx, in.UserId)
	if err != nil {
		return nil, err
	}
	return &system.LoginUserPermsResp{Perms: perms}, nil
}
