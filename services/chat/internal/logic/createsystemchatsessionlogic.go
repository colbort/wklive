package logic

import (
	"context"
	"strings"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateSystemChatSessionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateSystemChatSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateSystemChatSessionLogic {
	return &CreateSystemChatSessionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建系统会话
func (l *CreateSystemChatSessionLogic) CreateSystemChatSession(in *chat.SendSystemMessageReq) (*chat.InternalChatSessionResp, error) {
	session, shouldNotifyQueue, err := ensureOpenSession(l.ctx, l.svcCtx, in.GetMerchantId(), in.GetUserId(), chat.ChatSessionSource_CHAT_SESSION_SOURCE_SYSTEM, in.GetTitle(), in.GetCategory(), chat.ChatSessionPriority_CHAT_SESSION_PRIORITY_NORMAL, nil)
	if err != nil {
		return &chat.InternalChatSessionResp{Base: badBase(err.Error())}, nil
	}
	if shouldNotifyQueue {
		publishQueueEvent(l.ctx, l.svcCtx, session)
		if strings.TrimSpace(in.GetContent()) != "" {
			msg := newMessage(session, chat.ChatSenderType_CHAT_SENDER_TYPE_SYSTEM, 0, "system", "", chat.ChatMessageType_CHAT_MESSAGE_TYPE_TEXT, in.GetContent(), in.GetUrl(), in.GetFileName(), in.GetMimeType(), in.GetFileSize(), nil)
			if _, err := sendMessage(l.ctx, l.svcCtx, session, msg); err != nil {
				return &chat.InternalChatSessionResp{Base: errorBase(err)}, nil
			}
		}
	}
	return &chat.InternalChatSessionResp{Base: okBase(), Data: toProtoSession(session)}, nil
}
