package chat

const (
	ChatMessageChannel                  = "wklive:chat:messages"
	ChatMessageEventTypeMessage         = "chat.message"
	ChatMessageEventTypeSessionAccepted = "chat.session.accepted"
	ChatMessageEventTypeSessionClosed   = "chat.session.closed"
	ChatMessageEventTypeQueueUpdated    = "chat.queue.updated"
	ChatMessageEventTypeAgentStatus     = "chat.agent.status.updated"
	ChatWsEventConnected                = "connected"
	ChatWsEventError                    = "error"
	ChatWsEventSendUserMessage          = "send_user_message"
	ChatWsEventSendUserMessageResult    = "send_user_message.result"
	ChatWsEventSendAgentMessage         = "send_agent_message"
	ChatWsEventSendAgentMessageResult   = "send_agent_message.result"
	ChatWsEventAcceptChatSession        = "accept_chat_session"
	ChatWsEventAcceptChatSessionResult  = "accept_chat_session.result"
	ChatWsEventCloseChatSession         = "close_chat_session"
	ChatWsEventCloseChatSessionResult   = "close_chat_session.result"
)
