export interface RespBase {
  code: number;
  msg: string;
  total?: number;
  hasNext?: boolean;
  hasPrev?: boolean;
  nextCursor?: number;
  prevCursor?: number;
}

export interface ApiResp<T> extends RespBase {
  data: T;
}

export interface PageReq {
  cursor?: number;
  limit?: number;
}

export interface ChatMerchant {
  merchantId: number;
  apiKey: string;
  enabled: number;
  expireTime: number;
  createTimes: number;
  updateTimes: number;
}

export interface ChatSession {
  id: number;
  sessionNo: string;
  merchantId: number;
  userId: number;
  source: number;
  status: number;
  priority: number;
  agentId: number;
  title: string;
  category: string;
  lastMessage: string;
  lastSenderType: number;
  lastMessageTime: number;
  userUnreadCount: number;
  agentUnreadCount: number;
  closeTime: number;
  closeReason: string;
  extJson?: string;
  groupId: number;
  lastMessageNo: string;
  createTimes: number;
  updateTimes: number;
}

export interface ChatMessageSender {
  id: number;
  type: number;
  nickname: string;
  avatarUrl: string;
}

export interface ChatMessage {
  id?: number;
  messageNo: string;
  sessionNo: string;
  merchantId: number;
  userId: number;
  agentId: number;
  senderType: number;
  sender?: ChatMessageSender;
  messageType: number;
  content: string;
  mediaUrl: string;
  mediaName: string;
  mediaMime: string;
  mediaSize: number;
  status: number;
  readTime: number;
  createTimes: number;
  updateTimes: number;
}

export interface ChatWsEvent<T = unknown> {
  type: string;
  data: T;
}

export interface ConnectedPayload {
  message: string;
  merchantId: number;
  userId: number;
  sessionNo: string;
  temporary?: boolean;
}

export interface SendUserMessagePayload {
  messageType: number;
  content?: string;
  mediaUrl?: string;
  mediaName?: string;
  mediaMime?: string;
  mediaSize?: number;
  senderNickname?: string;
  senderAvatarUrl?: string;
}

export interface OpenChatSessionPayload {
  merchantId: number;
  source?: number;
  title?: string;
  category?: string;
  firstMessage?: string;
  senderNickname?: string;
  senderAvatarUrl?: string;
}

export interface SendChatMessagePayload extends SendUserMessagePayload {
  merchantId: number;
  clientMsgNo?: string;
}

export interface ListChatSessionsParams extends PageReq {
  merchantId: number;
  status?: number;
}

export interface ListChatMessagesParams extends PageReq {
  merchantId: number;
}

export interface MarkUserMessagesReadPayload {
  merchantId: number;
  lastMessageId?: number;
  lastMessageNo?: string;
}

export interface CloseChatSessionPayload {
  merchantId: number;
  closeReason?: string;
}

export interface ChatSessionResp extends RespBase {
  data: ChatSession;
}

export interface ListChatSessionsResp extends RespBase {
  data: ChatSession[];
}

export interface ListChatMessagesResp extends RespBase {
  data: ChatMessage[];
}

export interface ChatMessageResp extends RespBase {
  data: ChatMessage;
}
