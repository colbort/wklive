package chat

const (
	// 用户端/会话消息事件：chat-api 订阅后推送给 chat-ui
	ChatMessageChannel = "wklive:chat:messages"

	// 后台事件：chat-admin-api 订阅后推送给 chat-admin-ui / 商户管理后台
	ChatAdminEventChannel = "wklive:chat:admin:events"
)
