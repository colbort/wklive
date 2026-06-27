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
  userNickname: string;
  userAvatarUrl: string;
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
  extJson?: string | Record<string, unknown>;
  isGuest?: boolean;
  avatarUrl?: string;
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
