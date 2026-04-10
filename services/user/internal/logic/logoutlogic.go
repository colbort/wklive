package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/proto/user"
	"wklive/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type LogoutLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLogoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LogoutLogic {
	return &LogoutLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 用户登出
func (l *LogoutLogic) Logout(in *user.LogoutReq) (*user.AppCommonResp, error) {
	// 可以在这里执行登出逻辑，如删除 session 或 token 黑名单
	return &user.AppCommonResp{Base: helper.OkResp()}, nil
}
