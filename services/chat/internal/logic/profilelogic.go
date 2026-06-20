package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ProfileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProfileLogic {
	return &ProfileLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 当前登录用户资料
func (l *ProfileLogic) Profile(in *chat.ChatAdminProfileReq) (*chat.ChatAdminProfileResp, error) {
	// todo: add your logic here and delete this line

	return &chat.ChatAdminProfileResp{}, nil
}
