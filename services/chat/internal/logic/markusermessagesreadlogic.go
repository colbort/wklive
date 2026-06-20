package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type MarkUserMessagesReadLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMarkUserMessagesReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MarkUserMessagesReadLogic {
	return &MarkUserMessagesReadLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 标记用户侧已读
func (l *MarkUserMessagesReadLogic) MarkUserMessagesRead(in *chat.MarkUserMessagesReadReq) (*chat.AppMarkMessagesReadResp, error) {
	session, base, err := getSession(l.ctx, l.svcCtx, in.GetMerchantId(), in.GetSessionNo())
	if err != nil {
		return &chat.AppMarkMessagesReadResp{Base: errorBase(err)}, nil
	}
	if base != nil {
		return &chat.AppMarkMessagesReadResp{Base: base}, nil
	}
	if session.UserId != in.GetUserId() {
		return &chat.AppMarkMessagesReadResp{Base: notFoundBase("chat session not found")}, nil
	}
	if err := markRead(l.ctx, l.svcCtx, session, chat.ChatSenderType_CHAT_SENDER_TYPE_USER, in.GetUserId()); err != nil {
		return &chat.AppMarkMessagesReadResp{Base: errorBase(err)}, nil
	}
	return &chat.AppMarkMessagesReadResp{Base: okBase()}, nil
}
