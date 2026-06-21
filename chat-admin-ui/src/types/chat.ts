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
  status: number;
  priority: number;
  agentId: number;
  groupId: number;
  title: string;
  category: string;
  lastMessage: string;
  lastMessageTime: number;
  userUnreadCount: number;
}

export interface ChatMessage {
  id: number;
  messageNo: string;
  sessionNo: string;
  senderType: number;
  content: string;
  createTimes: number;
}
