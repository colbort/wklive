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

export interface OptionItem {
  value: number;
  code: string;
}

export interface OptionGroup {
  key: string;
  label: string;
  options: OptionItem[];
}

export interface ChatOptions {
  options: OptionGroup[];
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
  extJson?: string | Record<string, unknown>;
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
  senderType: number;
  sender?: ChatMessageSender;
  receiver?: ChatMessageSender;
  messageType: number;
  content: string;
  url: string;
  fileName: string;
  fileSize: number;
  mimeType: string;
  width: number;
  height: number;
  duration: number;
  status: number;
  extra?: string;
  readTime: number;
  createTime: number;
  updateTime: number;
}

export interface ChatWsEvent<T = unknown> {
  eventType?: string | number;
  data?: T;
  session?: ChatSession;
  sessionEvent?: ChatSessionEvent;
  session_event?: ChatSessionEvent;
  queue?: ChatQueueInfo;
  base?: RespBase;
}

export interface ConnectedPayload {
  message: string;
  merchantId: number;
  userId: number;
  sessionNo: string;
  temporary?: boolean;
  session?: ChatSession;
  queue?: ChatQueueInfo;
}

export interface ChatQueueInfo {
  merchantId?: number;
  sessionNo: string;
  userId: number | string;
  groupId?: number;
  position?: number;
  queuePosition?: number;
  waitingCount?: number;
  estimateWaitSeconds?: number;
  estimatedWaitSeconds?: number;
  message?: string;
  updateTimes?: number;
}

export interface ChatSessionEvent {
  sessionNo: string;
  merchantId: number;
  userId: number;
  agentId: number;
  operatorId: number;
  status: number;
  assignType: number;
  reason: string;
  message: string;
  session?: ChatSession;
  queue?: ChatQueueInfo;
  createdAt: number;
}

export interface SendUserMessagePayload {
  messageType: number;
  content?: string;
  url?: string;
  fileName?: string;
  mimeType?: string;
  fileSize?: number;
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

export interface ListChatMessagesParams extends PageReq {}

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

export type ChatEventType =
  | "CHAT_EVENT_TYPE_UNSPECIFIED"
  | "CHAT_EVENT_TYPE_MESSAGE"
  | "CHAT_EVENT_TYPE_SYSTEM_NOTICE"
  | "CHAT_EVENT_TYPE_USER_JOIN"
  | "CHAT_EVENT_TYPE_USER_LEAVE"
  | "CHAT_EVENT_TYPE_QUEUE_UPDATE"
  | "CHAT_EVENT_TYPE_AGENT_ACCEPTED"
  | "CHAT_EVENT_TYPE_AGENT_LEAVE"
  | "CHAT_EVENT_TYPE_TRANSFER_REQUEST"
  | "CHAT_EVENT_TYPE_TRANSFER_ACCEPT"
  | "CHAT_EVENT_TYPE_TRANSFER_REJECT"
  | "CHAT_EVENT_TYPE_SESSION_CLOSE"
  | "CHAT_EVENT_TYPE_EVALUATION_INVITE"
  | "CHAT_EVENT_TYPE_EVALUATION_SUBMIT"
  | "CHAT_EVENT_TYPE_TYPING"
  | "CHAT_EVENT_TYPE_MESSAGE_DELIVERED"
  | "CHAT_EVENT_TYPE_MESSAGE_READ"
  | "CHAT_EVENT_TYPE_MESSAGE_RECALL"
  | "CHAT_EVENT_TYPE_MESSAGE_DELETE"
  | "CHAT_EVENT_TYPE_HEARTBEAT"
  | "CHAT_EVENT_TYPE_ERROR";

export type ChatSenderType = number;
export type ChatQueueAction = number;
export type ChatMessageStatus = number;
export type ChatMessageOperateType = number;
export type ChatMessageDeleteScope = number;

export interface ChatSystemNoticePayload {
  sessionNo?: string;
  title?: string;
  content: string;
  level?: "info" | "warning" | "error" | string;
  showInChat?: boolean;
}

export interface ChatUserStatePayload {
  sessionNo: string;
  userId: string;
  userName?: string;
  avatar?: string;
  online: boolean;
  source?: number;
}

export interface ChatQueuePayload {
  sessionNo: string;
  userId: number | string;
  userName?: string;
  queueAction: ChatQueueAction;
  queuePosition?: number;
  waitingCount?: number;
  estimatedWaitSeconds?: number;
  sessionStatus?: number;
  actionTime: number;
}

export interface ChatAgentPayload {
  sessionNo: string;
  agentId: string;
  agentName?: string;
  agentAvatar?: string;
  agentStatus?: number;
  assignType?: number;
  sessionStatus?: number;
  remark?: string;
  actionTime: number;
}

export interface ChatTransferPayload {
  sessionNo: string;
  fromAgentId?: string;
  fromAgentName?: string;
  toAgentId?: string;
  toAgentName?: string;
  reason?: string;
  rejectReason?: string;
  actionTime?: number;
}

export interface ChatEvaluationPayload {
  sessionNo: string;
  userId?: string;
  agentId?: string;
  evaluationId?: string;
  rating?: number;
  tags?: string[];
  comment?: string;
  submitted?: boolean;
  evaluatedAt?: number;
}

export interface ChatTypingPayload {
  sessionNo: string;
  senderId: string;
  senderType: ChatSenderType;
  text?: string;
  actionTime: number;
}

export interface ChatMessageReceiptPayload {
  sessionNo: string;
  messageNo: string;
  senderId?: number;
  operatorId: number;
  operatorType: ChatSenderType;
  messageStatus: ChatMessageStatus;
  receiptTime: number;
}

export interface ChatMessageOperatePayload {
  sessionNo: string;
  messageNo: string;
  operateType: ChatMessageOperateType;
  operatorId: string;
  operatorType: ChatSenderType;
  deleteScope?: ChatMessageDeleteScope;
  reason?: string;
  operatedAt: number;
}

export interface ChatHeartbeatPayload {
  connectionId?: string;
  uid?: string;
  senderType?: ChatSenderType;
  clientTime?: number;
  serverTime?: number;
}

export interface ChatErrorPayload {
  sessionNo?: string;
  messageNo?: string;
  errorCode: number;
  errorMessage: string;
  detail?: string;
  retryable?: boolean;
}

export interface ChatMessageEvent {
  code?: number;
  msg?: string;
  eventType: ChatEventType;
  createdAt: number;
  message?: ChatMessage;
  session?: ChatSession;
  systemNotice?: ChatSystemNoticePayload;
  userState?: ChatUserStatePayload;
  queue?: ChatQueuePayload;
  agent?: ChatAgentPayload;
  transfer?: ChatTransferPayload;
  evaluation?: ChatEvaluationPayload;
  typing?: ChatTypingPayload;
  receipt?: ChatMessageReceiptPayload;
  messageOperate?: ChatMessageOperatePayload;
  heartbeat?: ChatHeartbeatPayload;
  error?: ChatErrorPayload;
}

export interface ChatWsRequest<TPayload = unknown> {
  eventType?: ChatEventType;
  data: TPayload;
}

export type ChatUiWsReq =
  | ChatWsRequest<SendUserMessagePayload>
  | ChatWsRequest<ChatMessageReceiptPayload>
  | ChatWsRequest<ChatTypingPayload>
  | ChatWsRequest<ChatEvaluationPayload>
  | ChatWsRequest<CloseChatSessionPayload>
  | ChatWsRequest<ChatMessageOperatePayload>
  | ChatWsRequest<ChatHeartbeatPayload>;

export type ChatUiWsResp = ChatMessageEvent & {
  eventType:
    | "CHAT_EVENT_TYPE_SYSTEM_NOTICE"
    | "CHAT_EVENT_TYPE_QUEUE_UPDATE"
    | "CHAT_EVENT_TYPE_AGENT_ACCEPTED"
    | "CHAT_EVENT_TYPE_AGENT_LEAVE"
    | "CHAT_EVENT_TYPE_MESSAGE"
    | "CHAT_EVENT_TYPE_MESSAGE_DELIVERED"
    | "CHAT_EVENT_TYPE_MESSAGE_READ"
    | "CHAT_EVENT_TYPE_TYPING"
    | "CHAT_EVENT_TYPE_EVALUATION_INVITE"
    | "CHAT_EVENT_TYPE_SESSION_CLOSE"
    | "CHAT_EVENT_TYPE_MESSAGE_RECALL"
    | "CHAT_EVENT_TYPE_MESSAGE_DELETE"
    | "CHAT_EVENT_TYPE_HEARTBEAT"
    | "CHAT_EVENT_TYPE_ERROR";
};
