import { request } from "./request";
import type { ListResponse, PageQuery } from "@/types/liquidity";

export const liquidityApi = {
  login: (data: { username: string; password: string }) =>
    request.post<unknown, { token: string; name?: string }>("/auth/login", data),
  dashboard: () => request.get("/dashboard"),
  providers: (params: PageQuery) =>
    request.get<unknown, ListResponse>("/providers", { params }),
  createProvider: (data: Record<string, unknown>) => request.post("/providers", data),
  testProvider: (id: number) => request.post(`/providers/${id}/test`),
  setProviderStatus: (id: number, data: Record<string, unknown>) =>
    request.put(`/providers/${id}/status`, data),
  symbolConfigs: (params: PageQuery) =>
    request.get<unknown, ListResponse>("/symbol-configs", { params }),
  createSymbolConfig: (data: Record<string, unknown>) =>
    request.post("/symbol-configs", data),
  symbolAction: (id: number, action: "start" | "pause" | "stop") =>
    request.post(`/symbol-configs/${id}/${action}`),
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
