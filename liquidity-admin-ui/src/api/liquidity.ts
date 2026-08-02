import { request } from "./request";
import type {
  ListResponse,
  OptionGroup,
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
  options: () => request.get<unknown, ApiResponse<OptionGroup[]>>("/options"),
  dashboard: () => request.get("/dashboard"),
  configOptions: () => request.get("/config-options"),
  providers: (params: PageQuery) =>
    request.get<unknown, ListResponse>("/providers", { params }),
  createProvider: (data: Record<string, unknown>) => request.post("/providers", data),
  providerDetail: (id: number) => request.get(`/providers/${id}`),
  updateProvider: (id: number, data: Record<string, unknown>) =>
    request.put(`/providers/${id}`, data),
  provisionInternalProvider: (data: ProvisionInternalProviderRequest) =>
    request.post("/providers/provision", data),
  testProvider: (id: number) => request.post(`/providers/${id}/test`),
  setProviderStatus: (id: number, data: Record<string, unknown>) =>
    request.put(`/providers/${id}/status`, data),
  symbolConfigs: (params: PageQuery) =>
    request.get<unknown, ListResponse>("/symbol-configs", { params }),
  symbolConfigDetail: (id: number) =>
    request.get(`/symbol-configs/${id}`),
  createSymbolConfig: (data: Record<string, unknown>) =>
    request.post("/symbol-configs", data),
  updateSymbolConfig: (id: number, data: Record<string, unknown>) =>
    request.put(`/symbol-configs/${id}`, data),
  symbolAction: (
    id: number,
    action: "start" | "pause" | "stop",
    version: number,
  ) => request.post(`/symbol-configs/${id}/${action}`, { version }),
  quoteOrders: (params: PageQuery) =>
    request.get<unknown, ListResponse>("/quote-orders", { params }),
  cancelAllQuoteOrders: (id: number, data: Record<string, unknown>) =>
    request.post(`/symbol-configs/${id}/cancel-quotes`, data),
  quoteCycles: (params: Record<string, unknown>) =>
    request.get<unknown, ListResponse>("/quote-cycles", { params }),
  externalOrders: (params: PageQuery) =>
    request.get<unknown, ListResponse>("/external-orders", { params }),
  externalFills: (params: Record<string, unknown>) =>
    request.get<unknown, ListResponse>("/external-fills", { params }),
  cancelExternalOrder: (id: number, data: Record<string, unknown>) =>
    request.post(`/external-orders/${id}/cancel`, data),
  hedgeTasks: (params: PageQuery) =>
    request.get<unknown, ListResponse>("/hedge-tasks", { params }),
  createManualHedge: (data: Record<string, unknown>) =>
    request.post("/hedge-tasks/manual", data),
  cancelHedgeTask: (id: number, data: Record<string, unknown>) =>
    request.post(`/hedge-tasks/${id}/cancel`, data),
  retryHedgeTask: (id: number, data: Record<string, unknown>) =>
    request.post(`/hedge-tasks/${id}/retry`, data),
  inventorySnapshots: (params: Record<string, unknown>) =>
    request.get<unknown, ListResponse>("/inventory-snapshots", { params }),
  latestInventory: (params: Record<string, unknown>) =>
    request.get("/inventory-snapshots/latest", { params }),
  riskEvents: (params: PageQuery) =>
    request.get<unknown, ListResponse>("/risk-events", { params }),
  resolveRiskEvent: (id: number, data: Record<string, unknown>) =>
    request.post(`/risk-events/${id}/resolve`, data),
  reconcileBatches: (params: PageQuery) =>
    request.get<unknown, ListResponse>("/reconcile-batches", { params }),
  runReconcile: (data: Record<string, unknown>) =>
    request.post("/reconcile-batches/run", data),
  reconcileDetails: (batchId: number, params: Record<string, unknown> = {}) =>
    request.get<unknown, ListResponse>(`/reconcile-batches/${batchId}/details`, {
      params,
    }),
  resolveReconcileDifference: (id: number, data: Record<string, unknown>) =>
    request.post(`/reconcile-differences/${id}/resolve`, data),
};
