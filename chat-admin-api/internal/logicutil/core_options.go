package logicutil

import (
	"chat-admin-api/internal/types"
	"wklive/proto/chat"
)

func CoreOptions() []types.OptionsGroup {
	options := make([]types.OptionsGroup, 0)
	options = append(options, CommonOptions()...)
	return options
}

func CommonOptions() []types.OptionsGroup {
	return []types.OptionsGroup{
		EnumGroup("chatEventType", "事件类型", chat.ChatEventType_CHAT_EVENT_TYPE_UNKNOWN.Descriptor()),
		EnumGroup("chatAgentStatus", "客服坐席状态", chat.ChatAgentStatus_CHAT_AGENT_STATUS_UNKNOWN.Descriptor()),
	}
}
