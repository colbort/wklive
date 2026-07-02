import type { ChatEventType } from "@/api/constant";

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

export interface ChatMessageUser {
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
  sender?: ChatMessageUser;
  receiver?: ChatMessageUser;
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

export interface WsConnectedPayload {
  message: string;
  merchantId: number;
  userId: number;
  sessionNo: string;
  nickname?: string;
  avatarUrl?: string;
  isGuest?: boolean;
}

export interface SendUserMessagePayload {
  sessionNo?: string;
  clientMessageId?: string;
  messageType: number;
  content?: string;
  url?: string;
  fileName?: string;
  fileSize?: number;
  mimeType?: string;
  width?: number;
  height?: number;
  duration?: number;
  extra?: string;
  sender: ChatMessageUser;
  receiver?: ChatMessageUser;
  merchantId?: number;
  isGuest?: boolean;   
}

export interface ListChatMessagesParams extends PageReq {}

export interface CloseChatSessionPayload {
  merchantId: number;
  closeReason?: string;
}

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
  userId: number;
  nickname?: string;
  queueAction: ChatQueueAction;
  queuePosition?: number;
  waitingCount?: number;
  estimatedWaitSeconds?: number;
  sessionStatus?: number;
  actionTime: number;
}

export interface ChatAgentPayload {
  sessionNo: string;
  agentId: number;
  agentUserId: number;
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
  userId?: number;
  agentId?: number;
  evaluationId?: number;
  rating?: number;
  tags?: string[];
  comment?: string;
  submitted?: boolean;
  evaluatedAt?: number;
}

export interface ChatTypingPayload {
  sessionNo: string;
  senderId: number;
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
  operatorId: number;
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
  connected?: WsConnectedPayload;
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

export type ChatWsEvent = ChatMessageEvent;

export interface ChatWsRequest<TPayload = unknown> {
  eventType?: ChatEventType;
  data: TPayload;
}

export type ChatUiWsReq =
  ChatWsRequest<SendUserMessagePayload>;
