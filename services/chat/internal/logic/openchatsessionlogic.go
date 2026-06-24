package logic

import (
	"context"
	"strings"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/types/known/structpb"
)

type OpenChatSessionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewOpenChatSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OpenChatSessionLogic {
	return &OpenChatSessionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建或获取当前会话
func (l *OpenChatSessionLogic) OpenChatSession(in *chat.OpenChatSessionReq) (*chat.AppChatSessionResp, error) {
	merchantID, userID, base, err := chatAppIdentityFromMetadata(l.ctx)
	if base != nil {
		return &chat.AppChatSessionResp{Base: base}, nil
	}
	if err != nil {
		return &chat.AppChatSessionResp{Base: errorBase(err)}, nil
	}
	title := strings.TrimSpace(in.GetTitle())
	if title == "" {
		title = strings.TrimSpace(in.GetSenderNickname())
	}
	session, shouldNotifyQueue, err := ensureOpenSession(l.ctx, l.svcCtx, merchantID, userID, normalizeSource(in.GetSource()), title, in.GetCategory(), chat.ChatSessionPriority_CHAT_SESSION_PRIORITY_NORMAL, userSnapshotExt(in.GetSenderAvatarUrl()))
	if err != nil {
		return &chat.AppChatSessionResp{Base: badBase(err.Error())}, nil
	}
	if shouldNotifyQueue {
		publishQueueEvent(l.ctx, l.svcCtx, session)
		if strings.TrimSpace(in.GetFirstMessage()) != "" {
			msg := newMessage(session, chat.ChatSenderType_CHAT_SENDER_TYPE_USER, userID, in.GetSenderNickname(), in.GetSenderAvatarUrl(), chat.ChatMessageType_CHAT_MESSAGE_TYPE_TEXT, in.GetFirstMessage(), "", "", "", 0, nil)
			if _, err := sendMessage(l.ctx, l.svcCtx, session, msg); err != nil {
				return &chat.AppChatSessionResp{Base: errorBase(err)}, nil
			}
		}
	}
	return &chat.AppChatSessionResp{Base: okBase(), Data: toProtoSession(session)}, nil
}

func userSnapshotExt(avatarUrl string) *structpb.Struct {
	avatarUrl = strings.TrimSpace(avatarUrl)
	if avatarUrl == "" {
		return nil
	}
	ext, err := structpb.NewStruct(map[string]interface{}{
		"userAvatarUrl": avatarUrl,
	})
	if err != nil {
		return nil
	}
	return ext
}
