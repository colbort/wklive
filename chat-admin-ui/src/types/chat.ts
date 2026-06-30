export type UserType = 1 | 2;

export interface ChatUser {
  id: number;
  merchantId: number;
  userType: UserType;
  isOwner: number;
  username: string;
  nickname: string;
  avatarUrl: string;
  mobile: string;
  email: string;
  enabled: number;
  remark: string;
}

export interface ChatAgent {
  id: number;
  merchantId: number;
  chatUserId: number;
  agentNo: string;
  welcomeMessage: string;
  status: number;
  autoOnline: number;
  maxSessionCount: number;
  currentSessionCount: number;
  lastActiveTime: number;
  groupId: number;
  remark: string;
  createTimes: number;
  updateTimes: number;
}

export interface ChatGroup {
  id: number;
  merchantId: number;
  groupCode: string;
  groupName: string;
  description: string;
  enabled: number;
  sort: number;
  remark: string;
  createTimes: number;
  updateTimes: number;
}

export interface ChatCategory {
  id: number;
  merchantId: number;
  parentId: number;
  categoryCode: string;
  categoryName: string;
  groupId: number;
  enabled: number;
  sort: number;
  remark: string;
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
  groupId: number;
  title: string;
  category: string;
  lastMessage: string;
  lastSenderType: number;
  lastMessageTime: number;
  userUnreadCount: number;
  agentUnreadCount: number;
  closeTime: number;
  closeReason: string;
  extJson: ChatSessionExtJson;
  isGuest?: boolean;
  avatarUrl?: string;
  lastMessageNo: string;
  createTimes: number;
  updateTimes: number;
}

export interface ChatSessionExtJson {
  nickname?: string;
  avatarUrl?: string;
  [key: string]: unknown;
}

export interface ChatMessageSender {
  id: number;
  type: number;
  nickname: string;
  avatarUrl: string;
}

export interface ChatMessage {
  id: number;
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

export interface ChatQueueInfo {
  merchantId: number;
  sessionNo: string;
  userId: number;
  groupId: number;
  position: number;
  waitingCount: number;
  estimateWaitSeconds: number;
  message: string;
  updateTimes: number;
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
  userId: string;
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

export interface SendAgentMessagePayload {
  merchantId?: number;
  agentId?: number;
  userId?: number;
  sessionNo?: string;
  messageType: number;
  content?: string;
  url?: string;
  fileName?: string;
  mimeType?: string;
  fileSize?: number;
  width?: number;
  height?: number;
  duration?: number;
  extra?: string;
}

export interface AcceptChatSessionPayload {
  merchantId?: number;
  agentId?: number;
  userId?: number;
  sessionNo: string;
  reason?: string;
}

export interface CloseAgentChatSessionPayload {
  merchantId?: number;
  userId?: number;
  sessionNo: string;
  closeReason?: string;
}

export type ChatAdminUiWsReq =
  | ChatWsRequest<AcceptChatSessionPayload>
  | ChatWsRequest<SendAgentMessagePayload>
  | ChatWsRequest<ChatMessageReceiptPayload>
  | ChatWsRequest<ChatTypingPayload>
  | ChatWsRequest<ChatEvaluationPayload>
  | ChatWsRequest<CloseAgentChatSessionPayload>
  | ChatWsRequest<ChatTransferPayload>
  | ChatWsRequest<ChatAgentPayload>
  | ChatWsRequest<ChatMessageOperatePayload>
  | ChatWsRequest<ChatSystemNoticePayload>
  | ChatWsRequest<ChatHeartbeatPayload>;

export type ChatAdminUiWsResp = ChatMessageEvent & {
  eventType:
    | "CHAT_EVENT_TYPE_SYSTEM_NOTICE"
    | "CHAT_EVENT_TYPE_USER_JOIN"
    | "CHAT_EVENT_TYPE_USER_LEAVE"
    | "CHAT_EVENT_TYPE_QUEUE_UPDATE"
    | "CHAT_EVENT_TYPE_AGENT_ACCEPTED"
    | "CHAT_EVENT_TYPE_AGENT_LEAVE"
    | "CHAT_EVENT_TYPE_MESSAGE"
    | "CHAT_EVENT_TYPE_MESSAGE_DELIVERED"
    | "CHAT_EVENT_TYPE_MESSAGE_READ"
    | "CHAT_EVENT_TYPE_TYPING"
    | "CHAT_EVENT_TYPE_EVALUATION_SUBMIT"
    | "CHAT_EVENT_TYPE_SESSION_CLOSE"
    | "CHAT_EVENT_TYPE_TRANSFER_REQUEST"
    | "CHAT_EVENT_TYPE_TRANSFER_ACCEPT"
    | "CHAT_EVENT_TYPE_TRANSFER_REJECT"
    | "CHAT_EVENT_TYPE_MESSAGE_RECALL"
    | "CHAT_EVENT_TYPE_MESSAGE_DELETE"
    | "CHAT_EVENT_TYPE_HEARTBEAT"
    | "CHAT_EVENT_TYPE_ERROR";
};
