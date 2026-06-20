package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateChatGroupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateChatGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateChatGroupLogic {
	return &UpdateChatGroupLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新客服分组
func (l *UpdateChatGroupLogic) UpdateChatGroup(in *chat.UpdateChatGroupReq) (*chat.AdminChatGroupResp, error) {
	// todo: add your logic here and delete this line

	return &chat.AdminChatGroupResp{}, nil
}
