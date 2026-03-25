package logic

import (
	"context"

	"wklive/proto/user"
	"wklive/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UnlockUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUnlockUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnlockUserLogic {
	return &UnlockUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 解锁用户（解除登录锁定）
func (l *UnlockUserLogic) UnlockUser(in *user.UnlockUserReq) (*user.AdminCommonResp, error) {
	// todo: add your logic here and delete this line

	return &user.AdminCommonResp{}, nil
}
