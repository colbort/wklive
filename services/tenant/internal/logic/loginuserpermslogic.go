package logic

import (
	"context"

	"wklive/proto/tenant"
	"wklive/services/tenant/internal/svc"

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

// 获取登录租户的权限列表
func (l *LoginUserPermsLogic) LoginUserPerms(in *tenant.LoginUserPermsReq) (*tenant.LoginUserPermsResp, error) {
	// todo: add your logic here and delete this line

	return &tenant.LoginUserPermsResp{}, nil
}
