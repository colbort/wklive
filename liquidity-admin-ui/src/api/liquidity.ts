import { request } from "./request";
import type {
  ListResponse,
  PageQuery,
  ProvisionInternalProviderRequest,
} from "@/types/liquidity";

export type ApiResponse<T> = {
  code: number;
  msg?: string;
  data?: T;
};

export type LoginData = {
  token: string;
  exp: number;
  userId: number;
  nickname: string;
  appScope: number;
};

export type ProfileUser = {
  id: number;
  username: string;
  nickname?: string;
  avatar?: string;
  userType: number;
  isOwner: number;
  google2FaEnabled: number;
  appScope: number;
};

export type MenuNode = {
  id: number;
  parentId: number;
  name: string;
  menuType: number;
  path?: string;
  component?: string;
  icon?: string;
  sort: number;
  visible: number;
  enabled: number;
  perms?: string;
  children?: MenuNode[];
};

export type ProfileData = {
  user: ProfileUser;
  menus: MenuNode[];
  perms: string[];
  roleIds: number[];
};

export const liquidityApi = {
  login: (data: { username: string; password: string; googleCode?: string }) =>
    request.post<unknown, ApiResponse<LoginData>>("/auth/login", data),
  profile: () => request.get<unknown, ApiResponse<ProfileData>>("/auth/profile"),
  dashboard: () => request.get("/dashboard"),
  configOptions: () => request.get("/config-options"),
  providers: (params: PageQuery) =>
    request.get<unknown, ListResponse>("/providers", { params }),
  createProvider: (data: Record<string, unknown>) => request.post("/providers", data),
  provisionInternalProvider: (data: ProvisionInternalProviderRequest) =>
    request.post("/providers/provision", data),
  testProvider: (id: number) => request.post(`/providers/${id}/test`),
  setProviderStatus: (id: number, data: Record<string, unknown>) =>
    request.put(`/providers/${id}/status`, data),
  symbolConfigs: (params: PageQuery) =>
    request.get<unknown, ListResponse>("/symbol-configs", { params }),
  createSymbolConfig: (data: Record<string, unknown>) =>
    request.post("/symbol-configs", data),
  symbolAction: (
    id: number,
    action: "start" | "pause" | "stop",
    version: number,
  ) => request.post(`/symbol-configs/${id}/${action}`, { version }),
  quoteOrders: (params: PageQuery) =>
    request.get<unknown, ListResponse>("/quote-orders", { params }),
  externalOrders: (params: PageQuery) =>
    request.get<unknown, ListResponse>("/external-orders", { params }),
  hedgeTasks: (params: PageQuery) =>
    request.get<unknown, ListResponse>("/hedge-tasks", { params }),
  riskEvents: (params: PageQuery) =>
    request.get<unknown, ListResponse>("/risk-events", { params }),
  reconcileBatches: (params: PageQuery) =>
    request.get<unknown, ListResponse>("/reconcile-batches", { params }),
};
