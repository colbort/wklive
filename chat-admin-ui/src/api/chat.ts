import {
  deleteBase,
  getData,
  getRaw,
  postBase,
  postData,
  putData,
} from "./request";
import type {
  ChatAgent,
  ChatCategory,
  ChatGroup,
  ChatMessage,
  ChatSession,
  ChatUser,
} from "@/types/chat";

export interface LoginReq {
  username: string;
  password: string;
  googleCode?: string;
}

export interface TokenInfo {
  accessToken: string;
  refreshToken: string;
  expireTime: number;
}

export interface LoginData {
  token: TokenInfo;
  user: ChatUser;
  agent?: ChatAgent;
}

export interface ProfileResp {
  code: number;
  msg: string;
  user: ChatUser;
  agent?: ChatAgent;
}

export interface OptionItem {
  key: string;
  label?: string;
  value: number;
  tagType?: "success" | "info" | "warning" | "danger" | "primary";
}

export interface ChatAdminOptions {
  agentStatuses: OptionItem[];
}

export interface ChatGroupPayload {
  merchantId: number;
  groupCode?: string;
  groupName: string;
  description?: string;
  enabled?: number;
  sort?: number;
  remark?: string;
}

export interface ChatCategoryPayload {
  merchantId: number;
  parentId?: number;
  categoryCode?: string;
  categoryName: string;
  enabled?: number;
  sort?: number;
  remark?: string;
}

export interface CreateChatAgentPayload {
  maxSessionCount?: number;
  groupId?: number;
  welcomeMessage?: string;
  username?: string;
  password?: string;
  nickname?: string;
  mobile?: string;
  email?: string;
  enabled?: number;
  autoOnline?: number;
  remark?: string;
}

export interface UpdateChatAgentPayload {
  merchantId: number;
  maxSessionCount?: number;
  groupId?: number;
  welcomeMessage?: string;
  autoOnline?: number;
  remark?: string;
}

export interface PageChatMessagesParams {
  merchantId: number;
  cursor?: number;
  limit?: number;
  senderType?: number;
}

export interface SendAgentMessagePayload {
  merchantId: number;
  agentId: number;
  messageType: number;
  content?: string;
  mediaUrl?: string;
  mediaName?: string;
  mediaMime?: string;
  mediaSize?: number;
  clientMsgNo?: string;
}

export function login(data: LoginReq) {
  return postData<LoginData>("/login", data);
}

export function logout() {
  return postBase("/logout");
}

export function profile() {
  return getRaw<ProfileResp>("/profile");
}

export function options() {
  return getData<ChatAdminOptions>("/options");
}

export function pageAgents(params: Record<string, unknown>) {
  return getData<ChatAgent[]>("/agents", params);
}

export function createAgent(data: CreateChatAgentPayload) {
  return postData<ChatAgent>("/agents", data);
}

export function updateAgent(id: number, data: UpdateChatAgentPayload) {
  return putData<ChatAgent>(`/agents/${id}`, data);
}

export function updateAgentStatus(
  id: number,
  data: {
    status: number;
  },
) {
  return putData<ChatAgent>(`/agents/${id}/status`, data);
}

export function pageCategories(params: Record<string, unknown>) {
  return getData<ChatCategory[]>("/categories", params);
}

export function createCategory(data: ChatCategoryPayload) {
  return postData<ChatCategory>("/categories", data);
}

export function updateCategory(id: number, data: ChatCategoryPayload) {
  return putData<ChatCategory>(`/categories/${id}`, data);
}

export function deleteCategory(id: number, merchantId: number) {
  return deleteBase(`/categories/${id}`, { merchantId });
}

export function pageGroups(params: Record<string, unknown>) {
  return getData<ChatGroup[]>("/groups", params);
}

export function createGroup(data: ChatGroupPayload) {
  return postData<ChatGroup>("/groups", data);
}

export function updateGroup(id: number, data: ChatGroupPayload) {
  return putData<ChatGroup>(`/groups/${id}`, data);
}

export function deleteGroup(id: number, merchantId: number) {
  return deleteBase(`/groups/${id}`, { merchantId });
}

export function pageSessions(params: Record<string, unknown>) {
  return getData<ChatSession[]>("/sessions", params);
}

export function getSession(sessionNo: string, merchantId: number) {
  return getData<ChatSession>(`/sessions/${encodeURIComponent(sessionNo)}`, {
    merchantId,
  });
}

export function pageMessages(
  sessionNo: string,
  params: PageChatMessagesParams,
) {
  return getData<ChatMessage[]>(
    `/sessions/${encodeURIComponent(sessionNo)}/messages`,
    params,
  );
}

export function sendAgentMessage(
  sessionNo: string,
  data: SendAgentMessagePayload,
) {
  return postData<ChatMessage>(
    `/sessions/${encodeURIComponent(sessionNo)}/messages`,
    data,
  );
}

export function chatAdminWsUrl(params: {
  token: string;
  merchantId?: number;
  agentId?: number;
  sessionNo?: string;
}) {
  const baseURL = import.meta.env.VITE_API_BASE_URL || "/chat/admin";
  const base = String(baseURL).replace(/\/$/, "");
  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  const url = new URL(
    `${base}/ws/messages`,
    `${protocol}//${window.location.host}`,
  );
  url.searchParams.set("token", params.token);
  if (params.merchantId) {
    url.searchParams.set("merchantId", String(params.merchantId));
  }
  if (params.agentId) {
    url.searchParams.set("agentId", String(params.agentId));
  }
  if (params.sessionNo) {
    url.searchParams.set("sessionNo", params.sessionNo);
  }
  return url.toString();
}
