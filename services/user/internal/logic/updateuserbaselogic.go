package logic

import (
	"context"

	"wklive/proto/user"
	"wklive/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateUserBaseLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateUserBaseLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserBaseLogic {
	return &UpdateUserBaseLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新用户基本信息
func (l *UpdateUserBaseLogic) UpdateUserBase(in *user.UpdateUserBaseReq) (*user.UpdateUserBaseResp, error) {
	// todo: add your logic here and delete this line

	return &user.UpdateUserBaseResp{}, nil
}
