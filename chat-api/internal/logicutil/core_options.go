package logicutil

import (
	"chat-api/internal/types"
	"wklive/proto/chat"
)

func CoreOptions() []types.OptionsGroup {
	options := make([]types.OptionsGroup, 0)
	options = append(options, CommonOptions()...)
	return options
}

func CommonOptions() []types.OptionsGroup {
	return []types.OptionsGroup{
		EnumGroup("chatEventType", "事件类型", chat.ChatEventType_CHAT_EVENT_TYPE_UNSPECIFIED.Descriptor()),
		EnumGroup("chatSessionStatus", "客服会话状态", chat.ChatSessionStatus_CHAT_SESSION_STATUS_UNKNOWN.Descriptor()),
		EnumGroup("chatSenderType", "消息发送方类型", chat.ChatSenderType_CHAT_SENDER_TYPE_UNKNOWN.Descriptor()),
		EnumGroup("chatMessageType", "消息类型", chat.ChatMessageType_CHAT_MESSAGE_TYPE_UNKNOWN.Descriptor()),
		EnumGroup("chatMessageStatus", "消息发送状态", chat.ChatMessageStatus_CHAT_MESSAGE_STATUS_UNKNOWN.Descriptor()),
	}
}
