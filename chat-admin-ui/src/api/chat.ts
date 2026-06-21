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
  remark?: string;
}

export interface UpdateChatAgentPayload {
  merchantId: number;
  maxSessionCount?: number;
  groupId?: number;
  welcomeMessage?: string;
  remark?: string;
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

export function pageAgents(params: Record<string, unknown>) {
  return getData<ChatAgent[]>("/agents", params);
}

export function createAgent(data: CreateChatAgentPayload) {
  return postData<ChatAgent>("/agents", data);
}

export function updateAgent(id: number, data: UpdateChatAgentPayload) {
  return putData<ChatAgent>(`/agents/${id}`, data);
}

export function updateAgentStatus(id: number, data: {
  merchantId: number;
  status: number;
}) {
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
