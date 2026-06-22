// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package chat_session

import (
	"context"
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
	if isGuestSession(req.SessionNo) && l.svcCtx.ChatMessageHub != nil {
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

func isGuestSession(sessionNo string) bool {
	return strings.HasPrefix(strings.TrimSpace(sessionNo), "GS")
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
		Id:          item.GetId(),
		MessageNo:   item.GetMessageNo(),
		SessionNo:   item.GetSessionNo(),
		MerchantId:  item.GetMerchantId(),
		UserId:      item.GetUserId(),
		AgentId:     item.GetAgentId(),
		SenderType:  int64(item.GetSenderType()),
		Sender:      protoMessageSenderToType(item.GetSender()),
		MessageType: int64(item.GetMessageType()),
		Content:     item.GetContent(),
		MediaUrl:    item.GetMediaUrl(),
		MediaName:   item.GetMediaName(),
		MediaMime:   item.GetMediaMime(),
		MediaSize:   item.GetMediaSize(),
		Status:      int64(item.GetStatus()),
		ReadTime:    item.GetReadTime(),
		CreateTimes: item.GetCreateTimes(),
		UpdateTimes: item.GetUpdateTimes(),
	}
}

func protoMessageSenderToType(item *chat.ChatMessageSender) types.ChatMessageSender {
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
