import type {
  ApiResp,
  ChatMerchant,
  ChatMessage,
  ChatSession,
  CloseChatSessionPayload,
  ListChatMessagesParams,
  ListChatSessionsParams,
  MarkUserMessagesReadPayload,
  OpenChatSessionPayload,
  RespBase,
  SendChatMessagePayload,
  SendUserMessagePayload,
} from "@/types/chat";

const apiBaseUrl = import.meta.env.VITE_CHAT_API_BASE_URL || "/chat";
const chatWsProtocol = "wklive-chat";

export const chatWsEvents = {
  connected: "connected",
  error: "error",
  message: "chat.message",
  sendUserMessage: "send_user_message",
  sendUserMessageResult: "send_user_message.result",
} as const;

export interface ChatAuthReq {
  apiKey: string;
  apiSecret: string;
}

export async function authChatMerchant(data: ChatAuthReq): Promise<ChatMerchant> {
  return requestData<ChatMerchant>("/auth", {
    method: "POST",
    body: data,
  });
}

export function openChatSession(token: string, data: OpenChatSessionPayload) {
  return requestData<ChatSession>("/sessions", {
    method: "POST",
    token,
    body: data,
  });
}

export function listChatSessions(token: string, params: ListChatSessionsParams) {
  return requestData<ChatSession[]>("/sessions", {
    method: "GET",
    token,
    params,
  });
}

export function getChatSession(token: string, sessionNo: string, merchantId: number) {
  return requestData<ChatSession>(`/sessions/${encodeURIComponent(sessionNo)}`, {
    method: "GET",
    token,
    params: { merchantId },
  });
}

export function sendUserMessage(
  token: string,
  sessionNo: string,
  data: SendChatMessagePayload,
) {
  return requestData<ChatMessage>(
    `/sessions/${encodeURIComponent(sessionNo)}/messages`,
    {
      method: "POST",
      token,
      body: data,
    },
  );
}

export function listChatMessages(
  token: string,
  sessionNo: string,
  params: ListChatMessagesParams,
) {
  return requestData<ChatMessage[]>(
    `/sessions/${encodeURIComponent(sessionNo)}/messages`,
    {
      method: "GET",
      token,
      params,
    },
  );
}

export function markUserMessagesRead(
  token: string,
  sessionNo: string,
  data: MarkUserMessagesReadPayload,
) {
  return requestBase(`/sessions/${encodeURIComponent(sessionNo)}/read`, {
    method: "PUT",
    token,
    body: data,
  });
}

export function closeChatSession(
  token: string,
  sessionNo: string,
  data: CloseChatSessionPayload,
) {
  return requestData<ChatSession>(`/sessions/${encodeURIComponent(sessionNo)}/close`, {
    method: "PUT",
    token,
    body: data,
  });
}

export function chatWsUrl(): string {
  const configured = import.meta.env.VITE_CHAT_WS_URL;
  if (configured) {
    return configured;
  }

  const base = apiBaseUrl.replace(/\/$/, "");
  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  return `${protocol}//${window.location.host}${base}/ws/messages`;
}

export interface CreateChatSocketOptions {
  merchantId: number;
  token?: string;
  onOpen?: (event: Event) => void;
  onEvent?: (event: MessageEvent<string>) => void;
  onError?: (event: Event) => void;
  onClose?: (event: CloseEvent) => void;
}

export function createChatSocket(options: CreateChatSocketOptions): WebSocket {
  const protocols = [chatWsProtocol, `merchant.${options.merchantId}`];
  const token = options.token?.trim();
  if (token) {
    protocols.push(`bearer.${token}`);
  }

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

export function sendChatSocketEvent(
  socket: WebSocket,
  type: string,
  data: unknown,
) {
  socket.send(JSON.stringify({ type, data }));
}

export function sendChatSocketUserMessage(
  socket: WebSocket,
  data: SendUserMessagePayload,
) {
  sendChatSocketEvent(socket, chatWsEvents.sendUserMessage, data);
}

interface RequestOptions {
  method: "GET" | "POST" | "PUT";
  token?: string;
  params?: object;
  body?: unknown;
}

async function requestData<T>(path: string, options: RequestOptions): Promise<T> {
  const payload = await request<ApiResp<T>>(path, options);
  return payload.data;
}

async function requestBase(path: string, options: RequestOptions): Promise<RespBase> {
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

function buildUrl(
  path: string,
  params?: object,
) {
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
  if (options.token?.trim()) {
    headers.Authorization = `Bearer ${options.token.trim()}`;
  }
  return headers;
}
