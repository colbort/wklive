import { computed, ref } from "vue";
import {
  chatWsEvents,
  createChatSocket,
  sendChatSocketUserMessage,
} from "@/api/chat";
import type {
  ChatMessage,
  ChatQueueInfo,
  ChatSessionEvent,
  ChatWsEvent,
  ConnectedPayload,
  SendUserMessagePayload,
} from "@/types/chat";

interface ConnectOptions {
  merchantId: number;
  userId?: string;
  nickname: string;
  avatarUrl: string;
}

interface WsResp<T> {
  code?: number;
  msg?: string;
  base?: {
    code?: number;
    msg?: string;
  };
  data: T;
}

type RawChatMessage = Partial<ChatMessage> & {
  message_no?: string;
  session_no?: string;
  merchant_id?: number;
  user_id?: number;
  agent_id?: number;
  sender_type?: number;
  message_type?: number;
  media_url?: string;
  media_name?: string;
  media_mime?: string;
  media_size?: number;
  create_times?: number;
  update_times?: number;
  read_time?: number;
};

const reconnectDelays = [1000, 2000, 5000, 10000, 15000];

export function useChatSocket() {
  const socket = ref<WebSocket | null>(null);
  const connected = ref<ConnectedPayload | null>(null);
  const messages = ref<ChatMessage[]>([]);
  const error = ref("");
  const queueStatus = ref("");
  const agentAccepted = ref(false);
  const sessionClosed = ref(false);
  const reconnectingIn = ref(0);
  const status = ref<"idle" | "connecting" | "open" | "closed" | "reconnecting">(
    "idle",
  );

  let lastOptions: ConnectOptions | null = null;
  let manualClose = false;
  let suppressNextCloseReconnect = false;
  let reconnectAttempts = 0;
  let reconnectTimer: ReturnType<typeof window.setTimeout> | null = null;
  let reconnectTicker: ReturnType<typeof window.setInterval> | null = null;

  const isOpen = computed(() => status.value === "open");
  const isTemporary = computed(() => Boolean(connected.value?.temporary));
  const reconnectLabel = computed(() => {
    if (status.value !== "reconnecting" || reconnectingIn.value <= 0) {
      return "";
    }
    return `${reconnectingIn.value}s 后重连`;
  });

  function connect(
    merchantId: number,
    user: { userId?: string; nickname?: string; avatarUrl?: string } = {},
  ) {
    manualClose = false;
    lastOptions = {
      merchantId,
      userId: user.userId,
      nickname: user.nickname || "",
      avatarUrl: user.avatarUrl || "",
    };
    reconnectAttempts = 0;
    clearReconnectTimer();
    openSocket(lastOptions);
  }

  function openSocket(options: ConnectOptions) {
    closeSocketOnly(true);
    error.value = "";
    reconnectingIn.value = 0;
    status.value = reconnectAttempts > 0 ? "reconnecting" : "connecting";
    const ws = createChatSocket({
      merchantId: options.merchantId,
      userId: options.userId,
      nickname: options.nickname,
      avatarUrl: options.avatarUrl,
      onOpen: () => {
        reconnectAttempts = 0;
        reconnectingIn.value = 0;
        status.value = "open";
      },
      onEvent: (event) => {
        handleSocketMessage(event.data);
      },
      onError: () => {
        error.value = "连接异常";
      },
      onClose: () => {
        if (socket.value === ws) {
          socket.value = null;
          status.value = "closed";
        }
        if (suppressNextCloseReconnect) {
          suppressNextCloseReconnect = false;
          return;
        }
        scheduleReconnect();
      },
    });
    socket.value = ws;
  }

  function sendText(content: string, nickname: string, avatarUrl = "") {
    const text = content.trim();
    if (
      !socket.value ||
      socket.value.readyState !== WebSocket.OPEN ||
      !text ||
      sessionClosed.value
    ) {
      return;
    }
    const payload: SendUserMessagePayload = {
      messageType: 1,
      content: text,
      senderNickname: nickname.trim(),
      senderAvatarUrl: avatarUrl.trim(),
    };
    sendChatSocketUserMessage(socket.value, payload);
  }

  function close() {
    manualClose = true;
    lastOptions = null;
    clearReconnectTimer();
    closeSocketOnly(false);
    connected.value = null;
    status.value = "idle";
  }

  function resetMessages() {
    messages.value = [];
  }

  function closeSocketOnly(suppressReconnect: boolean) {
    if (socket.value) {
      suppressNextCloseReconnect = suppressReconnect;
      socket.value.close();
    }
    socket.value = null;
  }

  function scheduleReconnect() {
    if (manualClose || !lastOptions) {
      return;
    }
    clearReconnectTimer();

    const delay =
      reconnectDelays[Math.min(reconnectAttempts, reconnectDelays.length - 1)];
    reconnectAttempts += 1;
    reconnectingIn.value = Math.ceil(delay / 1000);
    status.value = "reconnecting";

    reconnectTicker = window.setInterval(() => {
      reconnectingIn.value = Math.max(0, reconnectingIn.value - 1);
    }, 1000);
    reconnectTimer = window.setTimeout(() => {
      clearReconnectTimer();
      if (lastOptions && !manualClose) {
        openSocket(lastOptions);
      }
    }, delay);
  }

  function clearReconnectTimer() {
    if (reconnectTimer) {
      window.clearTimeout(reconnectTimer);
      reconnectTimer = null;
    }
    if (reconnectTicker) {
      window.clearInterval(reconnectTicker);
      reconnectTicker = null;
    }
    reconnectingIn.value = 0;
  }

  function handleSocketMessage(payload: string) {
    try {
      handleEvent(JSON.parse(payload) as ChatWsEvent);
    } catch {
      error.value = "无效的消息格式";
    }
  }

  function handleEvent(event: ChatWsEvent) {
    if (event.type === chatWsEvents.connected) {
      connected.value = event.data as ConnectedPayload;
      agentAccepted.value = false;
      sessionClosed.value = false;
      queueStatus.value =
        queueMessage((event.data as ConnectedPayload).queue) ||
        "请描述您的问题，发送后将进入客服队列。";
      return;
    }
    if (event.type === chatWsEvents.error) {
      const data = event.data as { message?: string };
      error.value = data.message || "消息发送失败";
      return;
    }
    if (event.type === chatWsEvents.sendUserMessageResult) {
      const resp = event.data as WsResp<RawChatMessage>;
      if (wsRespCode(resp) !== 200) {
        error.value = wsRespMsg(resp) || "消息发送失败";
        return;
      }
      pushMessage(normalizeMessage(resp.data));
      if (!agentAccepted.value) {
        queueStatus.value = "正在排队，客服会尽快接入。";
      }
      return;
    }
    if (event.type === chatWsEvents.sessionAccepted) {
      const message = eventMessage(event);
      agentAccepted.value = true;
      queueStatus.value =
        sessionEventMessage(event) || message?.content || "客服已接入。";
      if (message) {
        pushMessage(message);
      }
      return;
    }
    if (event.type === chatWsEvents.sessionClosed) {
      const message = eventMessage(event);
      sessionClosed.value = true;
      queueStatus.value =
        sessionEventMessage(event) || message?.content || "本次会话已结束。";
      if (message) {
        pushMessage(message);
      }
      closeSocketOnly(true);
      status.value = "closed";
      return;
    }
    if (event.type === chatWsEvents.queueUpdated) {
      const message = eventMessage(event);
      queueStatus.value =
        queueMessage(event.queue) ||
        sessionEventMessage(event) ||
        message?.content ||
        "正在排队，客服会尽快接入。";
      return;
    }
    if (event.type === chatWsEvents.message) {
      const message = eventMessage(event);
      if (!message) return;
      if (message.senderType === 2) {
        agentAccepted.value = true;
        queueStatus.value = "";
      }
      pushMessage(message);
    }
  }

  function eventMessage(event: ChatWsEvent): ChatMessage | undefined {
    if (!event.data) return undefined;
    return normalizeMessage(event.data as RawChatMessage);
  }

  function sessionEvent(event: ChatWsEvent): ChatSessionEvent | undefined {
    return event.sessionEvent ?? event.session_event;
  }

  function sessionEventMessage(event: ChatWsEvent) {
    return sessionEvent(event)?.message || "";
  }

  function queueMessage(queue?: ChatQueueInfo) {
    if (!queue) return "";
    if (queue.message) return queue.message;
    if (queue.position > 1) return `正在排队，您前面还有 ${queue.position - 1} 人。`;
    if (queue.position === 1) return "您是当前队列第 1 位，客服即将接入。";
    return "";
  }

  function pushMessage(message: ChatMessage) {
    if (!message?.messageNo) {
      return;
    }
    if (messages.value.some((item) => item.messageNo === message.messageNo)) {
      return;
    }
    messages.value.push(message);
  }

  function wsRespCode(resp: WsResp<unknown>) {
    return resp.code ?? resp.base?.code ?? 0;
  }

  function wsRespMsg(resp: WsResp<unknown>) {
    return resp.msg ?? resp.base?.msg ?? "";
  }

  function normalizeMessage(message: RawChatMessage): ChatMessage {
    return {
      id: message.id,
      messageNo: message.messageNo ?? message.message_no ?? "",
      sessionNo: message.sessionNo ?? message.session_no ?? "",
      merchantId: message.merchantId ?? message.merchant_id ?? 0,
      userId: message.userId ?? message.user_id ?? 0,
      agentId: message.agentId ?? message.agent_id ?? 0,
      senderType: normalizeSenderType(message.senderType ?? message.sender_type),
      sender: message.sender,
      messageType: message.messageType ?? message.message_type ?? 0,
      content: message.content ?? "",
      mediaUrl: message.mediaUrl ?? message.media_url ?? "",
      mediaName: message.mediaName ?? message.media_name ?? "",
      mediaMime: message.mediaMime ?? message.media_mime ?? "",
      mediaSize: message.mediaSize ?? message.media_size ?? 0,
      status: message.status ?? 0,
      readTime: message.readTime ?? message.read_time ?? 0,
      createTimes: message.createTimes ?? message.create_times ?? 0,
      updateTimes: message.updateTimes ?? message.update_times ?? 0,
    };
  }

  function normalizeSenderType(value: unknown) {
    if (typeof value === "number") return value;
    if (value === "CHAT_SENDER_TYPE_USER") return 1;
    if (value === "CHAT_SENDER_TYPE_AGENT") return 2;
    if (value === "CHAT_SENDER_TYPE_SYSTEM") return 3;
    return 0;
  }

  return {
    connected,
    error,
    queueStatus,
    agentAccepted,
    sessionClosed,
    isOpen,
    isTemporary,
    messages,
    reconnectLabel,
    reconnectingIn,
    status,
    close,
    connect,
    resetMessages,
    sendText,
  };
}
