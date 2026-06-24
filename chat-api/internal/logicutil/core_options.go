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
		EnumGroup("eventType", "事件类型", chat.ChatEventType_CHAT_EVENT_TYPE_UNKNOWN.Descriptor()),
	}
}
