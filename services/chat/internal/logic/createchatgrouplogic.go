package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateChatGroupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateChatGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateChatGroupLogic {
	return &CreateChatGroupLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建客服分组
func (l *CreateChatGroupLogic) CreateChatGroup(in *chat.CreateChatGroupReq) (*chat.AdminChatGroupResp, error) {
	// todo: add your logic here and delete this line

	return &chat.AdminChatGroupResp{}, nil
}
