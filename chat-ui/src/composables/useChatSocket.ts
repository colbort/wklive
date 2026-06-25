import { computed, ref } from "vue";
import {
  chatWsEvents,
  closeMyChatSession,
  createChatSocket,
  listChatMessagesWithMeta,
  options as loadOptions,
  sendChatSocketUserMessage,
} from "@/api/chat";
import type {
  ChatMessage,
  OptionGroup,
  ChatQueueInfo,
  ChatSessionEvent,
  ChatWsEvent,
  ConnectedPayload,
  SendUserMessagePayload,
} from "@/types/chat";

interface ConnectOptions {
  chatToken: string;
}

interface WsResp<T> {
  code?: number;
  msg?: string;
  hasNext?: boolean;
  nextCursor?: number;
  base?: {
    code?: number;
    msg?: string;
    hasNext?: boolean;
    nextCursor?: number;
  };
  data?: T;
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
  extra?: string;
};

const reconnectDelays = [1000, 2000, 5000, 10000, 15000];

export function useChatSocket() {
  const socket = ref<WebSocket | null>(null);
  const connected = ref<ConnectedPayload | null>(null);
  const messages = ref<ChatMessage[]>([]);
  const error = ref("");
  const queueStatus = ref("");
  const optionGroups = ref<OptionGroup[]>([]);
  const agentAccepted = ref(false);
  const activeAgentName = ref("");
  const sessionClosed = ref(false);
  const historyLoading = ref(false);
  const historyHasMore = ref(false);
  const historyNextCursor = ref(0);
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

  function connect(chatToken: string) {
    manualClose = false;
    void loadChatOptions();
    lastOptions = {
      chatToken,
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
      chatToken: options.chatToken,
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
    resetHistoryState();
    status.value = "idle";
  }

  async function endSession(closeReason = "user_closed", keepalive = false) {
    const current = connected.value;
    const token = lastOptions?.chatToken || "";
    const canCloseSession = Boolean(
      current?.sessionNo &&
        token &&
        !current.temporary &&
        !sessionClosed.value,
    );

    if (canCloseSession) {
      try {
        await closeMyChatSession(token, closeReason, keepalive);
      } catch (err) {
        if (!keepalive) {
          error.value = err instanceof Error ? err.message : "结束会话失败";
          return false;
        }
      }
    }

    sessionClosed.value = true;
    close();
    return true;
  }

  function resetMessages() {
    messages.value = [];
    resetHistoryState();
  }

  async function loadHistory(initial = false) {
    const current = connected.value;
    if (
      !current?.sessionNo ||
      current.temporary ||
      historyLoading.value ||
      (!initial && !historyHasMore.value)
    ) {
      return false;
    }

    historyLoading.value = true;
    try {
      const resp = await listChatMessagesWithMeta(
        {
          cursor: initial ? 0 : historyNextCursor.value,
          limit: 20,
        },
        lastOptions?.chatToken || "",
      );
      const list = Array.isArray(resp.data) ? resp.data : [];
      prependMessages(list.map(normalizeMessage).reverse());
      historyHasMore.value = Boolean(resp.hasNext);
      historyNextCursor.value = Number(resp.nextCursor || 0);
      return list.length > 0;
    } catch (err) {
      error.value = err instanceof Error ? err.message : "历史消息加载失败";
      return false;
    } finally {
      historyLoading.value = false;
    }
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
    const eventType = normalizeWsEventType(event.type);
    if (eventType === chatWsEvents.system) {
      ensureConnectedFromEvent(event);
      return;
    }
    if (eventType === chatWsEvents.userJoin) {
      ensureConnectedFromEvent(event);
      return;
    }
    if (eventType === chatWsEvents.queueJoin) {
      ensureConnectedFromEvent(event);
      if (agentAccepted.value) {
        queueStatus.value = serviceStatusMessage(activeAgentName.value);
        return;
      }
      queueStatus.value =
        queueMessage(event.queue) ||
        sessionEventMessage(event) ||
        eventMessage(event)?.content ||
        "正在排队，客服会尽快接入。";
      return;
    }
    if (eventType === chatWsEvents.error) {
      const data = event.data as { message?: string };
      error.value = data?.message || "消息发送失败";
      return;
    }
    if (eventType === chatWsEvents.delivered) {
      const message = eventMessage(event);
      if (message) pushMessage(message);
      return;
    }
    if (eventType === chatWsEvents.evaluationInvite) {
      queueStatus.value =
        sessionEventMessage(event) ||
        eventMessage(event)?.content ||
        "请对本次服务进行评价";
      return;
    }
    if (eventType === chatWsEvents.evaluationSubmit) {
      const resp = event.data as WsResp<RawChatMessage> | undefined;
      if (resp && wsRespCode(resp) && wsRespCode(resp) !== 200) {
        error.value = wsRespMsg(resp) || "评价提交失败";
        return;
      }
      queueStatus.value = "评价已提交";
      return;
    }
    if (eventType === chatWsEvents.typing) {
      if (eventMessage(event)?.senderType === 3) return;
      queueStatus.value =
        sessionEventMessage(event) || eventMessage(event)?.content || "客服正在输入";
      return;
    }
    if (eventType === chatWsEvents.stopTyping) {
      if (agentAccepted.value) {
        queueStatus.value = serviceStatusMessage(activeAgentName.value);
      }
      return;
    }
    if (eventType === chatWsEvents.queueUpdated) {
      ensureConnectedFromEvent(event);
      const message = eventMessage(event);
      if (agentAccepted.value) {
        queueStatus.value = serviceStatusMessage(activeAgentName.value);
        return;
      }
      queueStatus.value =
        queueMessage(event.queue) ||
        sessionEventMessage(event) ||
        message?.content ||
        "正在排队，客服会尽快接入。";
      return;
    }
    if (eventType === chatWsEvents.agentAssigned) {
      ensureConnectedFromEvent(event);
      const message = eventMessage(event);
      agentAccepted.value = true;
      activeAgentName.value = agentNameFromMessage(message) || activeAgentName.value;
      queueStatus.value = serviceStatusMessage(
        activeAgentName.value,
        sessionEventMessage(event) || message?.content || "",
      );
      return;
    }
    if (eventType === chatWsEvents.sessionClosed) {
      ensureConnectedFromEvent(event);
      const message = eventMessage(event);
      sessionClosed.value = true;
      queueStatus.value =
        sessionEventMessage(event) || message?.content || "本次会话已结束。";
      closeSocketOnly(true);
      status.value = "closed";
      return;
    }
    if (eventType === chatWsEvents.message) {
      ensureConnectedFromEvent(event);
      const message = eventMessage(event);
      if (!message) return;
      if (message.senderType === 2) {
        agentAccepted.value = true;
        activeAgentName.value = agentNameFromMessage(message) || activeAgentName.value;
        queueStatus.value = serviceStatusMessage(activeAgentName.value);
      }
      pushMessage(message);
      return;
    }
    if (eventType === 0 && event.data && isConnectedPayload(event.data)) {
      connected.value = event.data;
      agentAccepted.value = false;
      activeAgentName.value = "";
      sessionClosed.value = false;
      resetHistoryState();
      queueStatus.value =
        queueMessage(event.data.queue) ||
        "请描述您的问题，发送后将进入客服队列。";
      if (!event.data.temporary) {
        void loadHistory(true);
      }
      return;
    }
  }

  function eventMessage(event: ChatWsEvent): ChatMessage | undefined {
    if (!event.data) return undefined;
    return normalizeMessage(event.data as RawChatMessage);
  }

  function normalizeWsEventType(type: string | number) {
    const eventMap = optionCodeMap("chatEventType", {
      CHAT_EVENT_TYPE_MESSAGE: chatWsEvents.message,
      CHAT_EVENT_TYPE_SYSTEM: chatWsEvents.system,
      CHAT_EVENT_TYPE_USER_JOIN: chatWsEvents.userJoin,
      CHAT_EVENT_TYPE_USER_LEAVE: chatWsEvents.userLeave,
      CHAT_EVENT_TYPE_QUEUE_JOIN: chatWsEvents.queueJoin,
      CHAT_EVENT_TYPE_QUEUE_UPDATE: chatWsEvents.queueUpdated,
      CHAT_EVENT_TYPE_QUEUE_LEAVE: chatWsEvents.queueLeave,
      CHAT_EVENT_TYPE_AGENT_ASSIGNED: chatWsEvents.agentAssigned,
      CHAT_EVENT_TYPE_AGENT_JOIN: chatWsEvents.agentJoin,
      CHAT_EVENT_TYPE_AGENT_LEAVE: chatWsEvents.agentLeave,
      CHAT_EVENT_TYPE_TRANSFER: chatWsEvents.transfer,
      CHAT_EVENT_TYPE_SESSION_START: chatWsEvents.sessionStart,
      CHAT_EVENT_TYPE_SESSION_CLOSE: chatWsEvents.sessionClosed,
      CHAT_EVENT_TYPE_EVALUATION_INVITE: chatWsEvents.evaluationInvite,
      CHAT_EVENT_TYPE_EVALUATION_SUBMIT: chatWsEvents.evaluationSubmit,
      CHAT_EVENT_TYPE_TYPING: chatWsEvents.typing,
      CHAT_EVENT_TYPE_STOP_TYPING: chatWsEvents.stopTyping,
      CHAT_EVENT_TYPE_DELIVERED: chatWsEvents.delivered,
      CHAT_EVENT_TYPE_READ: chatWsEvents.read,
      CHAT_EVENT_TYPE_RECALL: chatWsEvents.recall,
      CHAT_EVENT_TYPE_HEARTBEAT: chatWsEvents.heartbeat,
      CHAT_EVENT_TYPE_ERROR: chatWsEvents.error,
      CHAT_EVENT_TYPE_NO_AGENT_ONLINE: chatWsEvents.noAgentOnline,
      CHAT_EVENT_TYPE_SESSION_TIMEOUT: chatWsEvents.sessionTimeout,
      CHAT_EVENT_TYPE_DELETE: chatWsEvents.delete,
      "chat.message": chatWsEvents.message,
      "chat.session.accepted": chatWsEvents.agentAssigned,
      "chat.session.closed": chatWsEvents.sessionClosed,
      "chat.queue.updated": chatWsEvents.queueUpdated,
      connected: chatWsEvents.system,
      error: chatWsEvents.error,
    });
    if (typeof type === "number") {
      return type;
    }
    return eventMap[type] || 0;
  }

  async function loadChatOptions() {
    if (optionGroups.value.length) return;
    try {
      const resp = await loadOptions();
      optionGroups.value = resp.options || [];
    } catch {
      optionGroups.value = [];
    }
  }

  function optionCodeMap(groupKey: string, fallback: Record<string, number>) {
    const group = optionGroups.value.find((item) => item.key === groupKey);
    if (!group?.options.length) return fallback;
    return group.options.reduce<Record<string, number>>((map, item) => {
      if (fallback[item.code] !== undefined) {
        map[item.code] = fallback[item.code];
      } else {
        map[item.code] = item.value;
      }
      return map;
    }, { ...fallback });
  }

  function optionValueMap(groupKey: string, fallback: Record<string, number>) {
    const group = optionGroups.value.find((item) => item.key === groupKey);
    if (!group?.options.length) return fallback;
    return group.options.reduce<Record<string, number>>((map, item) => {
      map[item.code] = item.value;
      return map;
    }, { ...fallback });
  }

  function sessionEvent(event: ChatWsEvent): ChatSessionEvent | undefined {
    return event.sessionEvent ?? event.session_event;
  }

  function sessionEventMessage(event: ChatWsEvent) {
    return sessionEvent(event)?.message || "";
  }

  function ensureConnectedFromEvent(event: ChatWsEvent) {
    const session = event.session || sessionEvent(event)?.session;
    const queue = event.queue || sessionEvent(event)?.queue;
    const sessionNo = session?.sessionNo || queue?.sessionNo || eventMessage(event)?.sessionNo || "";
    if (!sessionNo) return;
    const temporary = session ? sessionIsGuest(session) : true;
    if (connected.value?.sessionNo) {
      connected.value = {
        ...connected.value,
        message:
          queueMessage(queue) ||
          sessionEventMessage(event) ||
          connected.value.message,
        merchantId: session?.merchantId || queue?.merchantId || connected.value.merchantId,
        userId: session?.userId || queue?.userId || connected.value.userId,
        temporary: session ? temporary : connected.value.temporary,
        session: session || connected.value.session,
        queue: queue || connected.value.queue,
      };
      if (session && !temporary && !historyLoading.value && !messages.value.length) {
        void loadHistory(true);
      }
      return;
    }
    connected.value = {
      message: queueMessage(queue) || sessionEventMessage(event) || "",
      merchantId: session?.merchantId || queue?.merchantId || 0,
      userId: session?.userId || queue?.userId || 0,
      sessionNo,
      temporary,
      session,
      queue,
    };
    sessionClosed.value = false;
    resetHistoryState();
    if (session && !temporary) {
      void loadHistory(true);
    }
  }

  function sessionIsGuest(session?: { extJson?: string | Record<string, unknown> }) {
    if (!session?.extJson) return false;
    if (typeof session.extJson === "object") {
      return Boolean(session.extJson.isGuest);
    }
    try {
      return Boolean((JSON.parse(session.extJson) as { isGuest?: boolean }).isGuest);
    } catch {
      return false;
    }
  }

  function isConnectedPayload(data: unknown): data is ConnectedPayload {
    return Boolean(
      data &&
        typeof data === "object" &&
        "sessionNo" in data &&
        typeof (data as ConnectedPayload).sessionNo === "string",
    );
  }

  function agentNameFromMessage(message?: ChatMessage) {
    if (!message || message.senderType !== 2) return "";
    return (message.sender?.nickname || "").trim();
  }

  function serviceStatusMessage(agentName = "", preferred = "") {
    const text = preferred.trim();
    if (text.includes("客服正在为你服务")) return text;
    const name = agentName.trim();
    return name ? `${name} 客服正在为你服务` : "客服正在为你服务";
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

  function prependMessages(list: ChatMessage[]) {
    const seen = new Set(messages.value.map((item) => item.messageNo));
    const next = list.filter((item) => item.messageNo && !seen.has(item.messageNo));
    if (!next.length) {
      return;
    }
    messages.value = [...next, ...messages.value];
  }

  function resetHistoryState() {
    historyLoading.value = false;
    historyHasMore.value = false;
    historyNextCursor.value = 0;
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
      extra: message.extra ?? "",
      readTime: message.readTime ?? message.read_time ?? 0,
      createTimes: message.createTimes ?? message.create_times ?? 0,
      updateTimes: message.updateTimes ?? message.update_times ?? 0,
    };
  }

  function normalizeSenderType(value: unknown) {
    if (typeof value === "number") return value;
    const senderTypes = optionValueMap("chatSenderType", {
      CHAT_SENDER_TYPE_USER: 1,
      CHAT_SENDER_TYPE_AGENT: 2,
      CHAT_SENDER_TYPE_SYSTEM: 3,
    });
    if (typeof value === "string" && senderTypes[value] !== undefined) {
      return senderTypes[value];
    }
    return 0;
  }

  return {
    connected,
    error,
    queueStatus,
    agentAccepted,
    sessionClosed,
    historyHasMore,
    historyLoading,
    isOpen,
    isTemporary,
    messages,
    reconnectLabel,
    reconnectingIn,
    status,
    close,
    connect,
    endSession,
    loadHistory,
    resetMessages,
    sendText,
  };
}
