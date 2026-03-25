package logic

import (
	"context"

	"wklive/proto/user"
	"wklive/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChangeLoginPasswordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewChangeLoginPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChangeLoginPasswordLogic {
	return &ChangeLoginPasswordLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 修改登录密码
func (l *ChangeLoginPasswordLogic) ChangeLoginPassword(in *user.ChangeLoginPasswordReq) (*user.AppCommonResp, error) {
	// todo: add your logic here and delete this line

	return &user.AppCommonResp{}, nil
}
