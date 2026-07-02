import type {
  ApiResp,
  ChatOptions,
  ChatMessage,
  ChatWsRequest,
  ListChatMessagesParams,
  RespBase,
} from "@/types/chat";

const apiBaseUrl = import.meta.env.VITE_CHAT_API_BASE_URL || "/chat";
const chatWsProtocol = "wklive-chat";

export interface CreateChatTokenReq {
  apiKey: string;
  apiSecret: string;
  userId: number;
  nickname?: string;
  avatarUrl?: string;
  ttlSeconds?: number;
}

export interface ChatTokenResp {
  chatToken: string;
  expireAt: number;
}

export interface UploadFileData {
  url: string;
  fileName: string;
  fileSize: number;
  mimeType: string;
}

export function createChatToken(
  data: CreateChatTokenReq,
): Promise<ChatTokenResp> {
  return requestData<ChatTokenResp>("/internal/tokens", {
    method: "POST",
    body: data,
  });
}

export function setChatTokenCookie(chatToken: string): Promise<RespBase> {
  return requestBase("/internal/token-cookie", {
    method: "POST",
    body: { chatToken },
  });
}

export function options(): Promise<ChatOptions> {
  return requestData<ChatOptions>("/options", {
    method: "GET",
  });
}

export function listChatMessagesWithMeta(
  params: ListChatMessagesParams,
  chatToken = "",
) {
  return request<ApiResp<ChatMessage[]>>("/session/messages", {
    method: "GET",
    params,
    token: chatToken,
  });
}

export async function uploadChatFile(file: File, chatToken = "") {
  const data = new FormData();
  data.append("file", file, file.name);
  const payload = await requestForm<
    ApiResp<UploadFileData> & { Data?: UploadFileData }
  >("/upload/file", {
    method: "POST",
    body: data,
    token: chatToken,
  });
  const uploaded = payload.data || payload.Data;
  if (!uploaded?.url) {
    throw new Error("图片上传失败");
  }
  return uploaded;
}

export function chatWsUrl(): string {
  const queryWsUrl = new URLSearchParams(window.location.search).get("wsUrl");
  if (queryWsUrl) {
    return queryWsUrl;
  }

  const configured = import.meta.env.VITE_CHAT_WS_URL;
  if (configured) {
    return configured;
  }

  const base = apiBaseUrl.replace(/\/$/, "");
  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  return `${protocol}//${window.location.host}${base}/ws/messages`;
}

export interface CreateChatSocketOptions {
  onOpen?: (event: Event) => void;
  onEvent?: (event: MessageEvent<string>) => void;
  onError?: (event: Event) => void;
  onClose?: (event: CloseEvent) => void;
}

export function createChatSocket(options: CreateChatSocketOptions): WebSocket {
  const protocols = [chatWsProtocol];

  const socket = new WebSocket(chatWsUrl(), protocols);
  if (options.onOpen) {
    socket.addEventListener("open", options.onOpen);
  }
  if (options.onEvent) {
    socket.addEventListener("message", options.onEvent);
  }
  if (options.onError) {
    socket.addEventListener("error", options.onError);
  }
  if (options.onClose) {
    socket.addEventListener("close", options.onClose);
  }
  return socket;
}

export function sendChatSocketTypedEvent(
  socket: WebSocket,
  request: ChatWsRequest,
) {
  socket.send(JSON.stringify(request));
}

export function sendChatSocketEvent(
  socket: WebSocket,
  request: ChatWsRequest,
) {
  sendChatSocketTypedEvent(socket, request);
}

interface RequestOptions {
  method: "GET" | "POST" | "PUT";
  token?: string;
  params?: object;
  body?: unknown;
  headers?: Record<string, string>;
  keepalive?: boolean;
}

interface FormRequestOptions {
  method: "POST";
  token?: string;
  body: FormData;
}

async function requestData<T>(
  path: string,
  options: RequestOptions,
): Promise<T> {
  const payload = await request<ApiResp<T>>(path, options);
  return payload.data;
}

async function requestBase(
  path: string,
  options: RequestOptions,
): Promise<RespBase> {
  return request<RespBase>(path, options);
}

async function request<T extends RespBase>(
  path: string,
  options: RequestOptions,
): Promise<T> {
  const res = await fetch(buildUrl(path, options.params), {
    method: options.method,
    headers: buildHeaders(options),
    body: options.body === undefined ? undefined : JSON.stringify(options.body),
    keepalive: options.keepalive,
    credentials: "include",
    referrerPolicy: "no-referrer",
  });
  if (!res.ok) {
    throw new Error(`HTTP ${res.status}`);
  }

  const payload = (await res.json()) as T;
  if (payload.code !== 200) {
    throw new Error(payload.msg || "请求失败");
  }
  return payload;
}

async function requestForm<T extends RespBase>(
  path: string,
  options: FormRequestOptions,
): Promise<T> {
  const res = await fetch(buildUrl(path), {
    method: options.method,
    headers: buildAuthHeaders(options.token),
    body: options.body,
    credentials: "include",
    referrerPolicy: "no-referrer",
  });
  if (!res.ok) {
    throw new Error(`HTTP ${res.status}`);
  }

  const payload = (await res.json()) as T;
  if (payload.code !== 200) {
    throw new Error(payload.msg || "请求失败");
  }
  return payload;
}

function buildUrl(path: string, params?: object) {
  const url = new URL(`${apiBaseUrl}${path}`, window.location.origin);
  Object.entries(params || {}).forEach(([key, value]) => {
    if (value !== undefined && value !== "") {
      url.searchParams.set(key, String(value));
    }
  });
  if (/^https?:\/\//i.test(apiBaseUrl)) {
    return url.toString();
  }
  return `${url.pathname}${url.search}`;
}

function buildHeaders(options: RequestOptions) {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
  };
  Object.assign(headers, buildAuthHeaders(options.token));
  Object.entries(options.headers || {}).forEach(([key, value]) => {
    if (value !== undefined && value !== "") {
      headers[key] = value;
    }
  });
  return headers;
}

function buildAuthHeaders(token = "") {
  const headers: Record<string, string> = {};
  if (token.trim()) {
    headers.Authorization = `Bearer ${token.trim()}`;
  }
  return headers;
}
