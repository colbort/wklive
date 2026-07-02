import type { ChatEventType } from "@/api/constant";

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
  userId: number;
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
  isGuest: boolean;
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

export interface ChatMessageUser {
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

export type ChatQueueAction = number;
export type ChatSenderType = number;
export type ChatMessageStatus = number;
export type ChatMessageOperateType = number;
export type ChatMessageDeleteScope = number;

export interface WsConnectedPayload {
  message: string;
  merchantId: number;
  userId: number;
  nickname?: string;
  avatarUrl?: string;
  isGuest?: boolean;
}

export interface ChatUserStatePayload {
  sessionNo: string;
  userId: number;
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
  agentName?: string;
  agentAvatar?: string;
  agentStatus?: number;
  assignType?: number;
  sessionStatus?: number;
  remark?: string;
  actionTime: number;
}

export interface ChatSystemNoticePayload {
  sessionNo?: string;
  title?: string;
  content: string;
  level?: "info" | "warning" | "error" | string;
  showInChat?: boolean;
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

export interface ChatErrorPayload {
  messageNo?: string;
  errorCode: number;
  errorMessage: string;
  detail?: string;
  retryable?: boolean;
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
  isGuest?: boolean;
}

export interface ChatWsResponse {
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
  evaluation?: ChatEvaluationPayload;
  typing?: ChatTypingPayload;
  receipt?: ChatMessageReceiptPayload;
  messageOperate?: ChatMessageOperatePayload;
  error?: ChatErrorPayload;
}

export type ChatWsEvent = ChatWsResponse;

export interface ChatWsRequestBase {
  eventType: ChatEventType;
  requestId?: string;
  clientTime?: number;
}

export interface SendAgentMessagePayload {
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

export interface AcceptChatSessionPayload {
  merchantId?: number;
  agentId?: number;
  sessionNo: string;
  reason?: string;
  isGuest?: boolean;
}

export interface CloseAgentChatSessionPayload {
  merchantId?: number;
  userId?: number;
  sessionNo: string;
  closeReason?: string;
  isGuest?: boolean;
}

export interface ChatWsRequestPayloadMap {
  message: SendAgentMessagePayload;
  session: CloseAgentChatSessionPayload;
  systemNotice: Record<string, unknown>;
  userState: ChatUserStatePayload;
  queue: ChatQueuePayload;
  agent: AcceptChatSessionPayload;
  transfer: Record<string, unknown>;
  evaluation: Record<string, unknown>;
  typing: Record<string, unknown>;
  receipt: ChatMessageReceiptPayload;
  messageOperate: ChatMessageOperatePayload;
  heartbeat: Record<string, unknown>;
}

export type ChatWsRequestPayloadKey = keyof ChatWsRequestPayloadMap;

type ChatWsRequestOneof<K extends ChatWsRequestPayloadKey> = {
  [P in K]: Pick<ChatWsRequestPayloadMap, P> &
    Partial<Record<Exclude<ChatWsRequestPayloadKey, P>, never>>;
}[K];

export type ChatWsRequest = ChatWsRequestBase &
  ChatWsRequestOneof<ChatWsRequestPayloadKey>;

export function createChatWsRequest<K extends ChatWsRequestPayloadKey>(
  eventType: ChatEventType,
  payloadKey: K,
  payload: ChatWsRequestPayloadMap[K],
  options: Omit<ChatWsRequestBase, "eventType" | "clientTime"> &
    Partial<Pick<ChatWsRequestBase, "clientTime">> = {},
): ChatWsRequest {
  return {
    eventType,
    clientTime: Date.now(),
    ...options,
    [payloadKey]: payload,
  } as ChatWsRequest;
}
