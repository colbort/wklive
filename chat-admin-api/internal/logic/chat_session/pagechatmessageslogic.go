// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package chat_session

import (
	"context"
	"strconv"
	"strings"

	"chat-admin-api/internal/logicutil"

	"chat-admin-api/internal/svc"
	"chat-admin-api/internal/types"
	"wklive/proto/chat"

	"github.com/zeromicro/go-zero/core/logx"
)

type PageChatMessagesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPageChatMessagesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PageChatMessagesLogic {
	return &PageChatMessagesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PageChatMessagesLogic) PageChatMessages(req *types.PageChatMessagesReq) (resp *types.PageChatMessagesResp, err error) {
	if l.svcCtx.ChatMessageHub != nil && l.svcCtx.ChatMessageHub.IsTransientSession(strings.TrimSpace(req.SessionNo)) {
		return &types.PageChatMessagesResp{
			RespBase: types.RespBase{Code: 200, Msg: "OK"},
			Data: protoMessagesToTypes(
				l.svcCtx.ChatMessageHub.ListTransientMessages(
					req.MerchantId,
					req.SessionNo,
					req.SenderType,
					req.Limit,
				),
			),
		}, nil
	}
	return logicutil.Proxy[types.PageChatMessagesResp](l.ctx, req, l.svcCtx.ChatAdminCli.PageChatMessages)
}

func protoMessagesToTypes(list []*chat.ChatMessage) []types.ChatMessage {
	if len(list) == 0 {
		return []types.ChatMessage{}
	}
	resp := make([]types.ChatMessage, 0, len(list))
	for _, item := range list {
		resp = append(resp, protoMessageToType(item))
	}
	return resp
}

func protoMessageToType(item *chat.ChatMessage) types.ChatMessage {
	if item == nil {
		return types.ChatMessage{}
	}
	return types.ChatMessage{
		MessageNo:   item.GetMessageNo(),
		SessionNo:   item.GetSessionNo(),
		AgentId:     int64FromString(item.GetAgentId()),
		SenderType:  int64(protoMessageSenderType(item)),
		Sender:      protoMessageSenderToType(item.GetSender()),
		MessageType: int64(item.GetMessageType()),
		Content:     item.GetContent(),
		MediaUrl:    item.GetUrl(),
		MediaName:   item.GetFileName(),
		MediaMime:   item.GetMimeType(),
		MediaSize:   item.GetFileSize(),
		Status:      int64(item.GetStatus()),
		CreateTimes: item.GetCreateTime(),
		UpdateTimes: item.GetUpdateTime(),
	}
}

func protoMessageSenderToType(item *chat.ChatMessageUser) types.ChatMessageSender {
	if item == nil {
		return types.ChatMessageSender{}
	}
	return types.ChatMessageSender{
		Id:         item.GetId(),
		SenderType: int64(item.GetType()),
		Nickname:   item.GetNickname(),
		AvatarUrl:  item.GetAvatarUrl(),
	}
}

func protoMessageSenderType(item *chat.ChatMessage) chat.ChatSenderType {
	if item == nil || item.GetSender() == nil {
		return chat.ChatSenderType_CHAT_SENDER_TYPE_UNKNOWN
	}
	return item.GetSender().GetType()
}

func int64FromString(value string) int64 {
	id, _ := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	return id
}
