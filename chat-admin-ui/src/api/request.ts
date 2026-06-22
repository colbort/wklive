import axios from "axios";
import { useAuthStore } from "@/stores/auth";

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

export const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || "/chat/admin",
  timeout: 15000,
});

request.interceptors.request.use((config) => {
  const auth = useAuthStore();
  if (auth.token) {
    config.headers.Authorization = `Bearer ${auth.token}`;
  }
  const merchantId = auth.user?.merchantId || auth.agent?.merchantId || 0;
  if (merchantId) {
    config.headers["x-merchant-id"] = String(merchantId);
  }
  return config;
});

export async function getData<T>(url: string, params?: unknown) {
  const res = await request.get<ApiResp<T>>(url, { params });
  assertOk(res.data);
  return res.data;
}

export async function postData<T>(url: string, data?: unknown) {
  const res = await request.post<ApiResp<T>>(url, data);
  assertOk(res.data);
  return res.data;
}

export async function putData<T>(url: string, data?: unknown) {
  const res = await request.put<ApiResp<T>>(url, data);
  assertOk(res.data);
  return res.data;
}

export async function postBase(url: string, data?: unknown) {
  const res = await request.post<RespBase>(url, data);
  assertOk(res.data);
  return res.data;
}

export async function putBase(url: string, data?: unknown) {
  const res = await request.put<RespBase>(url, data);
  assertOk(res.data);
  return res.data;
}

export async function deleteBase(url: string, params?: unknown) {
  const res = await request.delete<RespBase>(url, { params });
  assertOk(res.data);
  return res.data;
}

export async function getRaw<T extends RespBase>(url: string, params?: unknown) {
  const res = await request.get<T>(url, { params });
  assertOk(res.data);
  return res.data;
}

function assertOk(resp: RespBase) {
  if (resp.code !== 200) {
    throw new Error(resp.msg || "请求失败");
  }
}
