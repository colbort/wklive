// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package chat_ws

import (
	"context"

	"chat-admin-api/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type MessagesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMessagesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MessagesLogic {
	return &MessagesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MessagesLogic) Messages() error {
	return nil
}
