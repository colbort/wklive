package logic

import (
	"context"

	"wklive/proto/user"
	"wklive/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateUserLevelLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateUserLevelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserLevelLogic {
	return &UpdateUserLevelLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新用户会员等级
func (l *UpdateUserLevelLogic) UpdateUserLevel(in *user.UpdateUserLevelReq) (*user.AdminCommonResp, error) {
	// todo: add your logic here and delete this line

	return &user.AdminCommonResp{}, nil
}
