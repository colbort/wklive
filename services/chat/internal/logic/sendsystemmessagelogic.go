package logic

import (
	"context"
	"strings"

	"wklive/proto/chat"
	"wklive/proto/common"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendSystemMessageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendSystemMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendSystemMessageLogic {
	return &SendSystemMessageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 发送系统消息
func (l *SendSystemMessageLogic) SendSystemMessage(in *chat.SendSystemMessageReq) (*chat.InternalChatMessageResp, error) {
	var session *models.TChatSession
	var base *common.RespBase
	var err error
	if strings.TrimSpace(in.GetSessionNo()) != "" {
		session, base, err = getSession(l.ctx, l.svcCtx, in.GetMerchantId(), in.GetSessionNo())
	} else {
		session, _, err = ensureOpenSession(l.ctx, l.svcCtx, in.GetMerchantId(), in.GetUserId(), chat.ChatSessionSource_CHAT_SESSION_SOURCE_SYSTEM, in.GetTitle(), in.GetCategory(), chat.ChatSessionPriority_CHAT_SESSION_PRIORITY_NORMAL, nil)
	}
	if err != nil {
		return &chat.InternalChatMessageResp{Base: errorBase(err)}, nil
	}
	if base != nil {
		return &chat.InternalChatMessageResp{Base: base}, nil
	}

	msg := newMessage(session, chat.ChatSenderType_CHAT_SENDER_TYPE_SYSTEM, 0, "system", normalizeMessageType(in.GetMessageType()), in.GetContent(), "", "", "", 0, in.GetPayload())
	msg, err = sendMessage(l.ctx, l.svcCtx, session, msg)
	if err != nil {
		return &chat.InternalChatMessageResp{Base: errorBase(err)}, nil
	}
	return &chat.InternalChatMessageResp{Base: okBase(), Data: toProtoMessage(msg)}, nil
}
