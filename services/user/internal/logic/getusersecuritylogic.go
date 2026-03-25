package logic

import (
	"context"

	"wklive/proto/user"
	"wklive/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserSecurityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserSecurityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserSecurityLogic {
	return &GetUserSecurityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取用户安全设置
func (l *GetUserSecurityLogic) GetUserSecurity(in *user.GetUserSecurityReq) (*user.GetUserSecurityResp, error) {
	// todo: add your logic here and delete this line

	return &user.GetUserSecurityResp{}, nil
}
