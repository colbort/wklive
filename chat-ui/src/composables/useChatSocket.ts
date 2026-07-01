import { computed, ref } from "vue";
import {
  closeMyChatSession,
  createChatSocket,
  listChatMessagesWithMeta,
  options as loadOptions,
  sendChatSocketUserMessage,
} from "@/api/chat";
import type {
  ChatMessage,
  OptionGroup,
  ChatQueuePayload,
  ChatWsEvent,
  WsConnectedPayload,
  SendUserMessagePayload,
  ChatAgentPayload,
} from "@/types/chat";
import { chatEventType, type ChatEventType } from "@/api/constant";

type ConnectOptions = Record<string, never>;

const reconnectDelays = [1000, 2000, 5000, 10000, 15000];

export function useChatSocket() {
  const socket = ref<WebSocket | null>(null);
  const connected = ref<WsConnectedPayload | null>(null);
  const messages = ref<ChatMessage[]>([]);
  const error = ref("");
  const queueStatus = ref("");
  const optionGroups = ref<OptionGroup[]>([]);
  const agent = ref<ChatAgentPayload | null>(null);
  const sessionClosed = ref(false);
  const historyLoading = ref(false);
  const historyHasMore = ref(false);
  const historyNextCursor = ref(0);
  const reconnectingIn = ref(0);
  const status = ref<
    "idle" | "connecting" | "open" | "closed" | "reconnecting"
  >("idle");

  let lastOptions: ConnectOptions | null = null;
  let manualClose = false;
  let suppressNextCloseReconnect = false;
  let reconnectAttempts = 0;
  let reconnectTimer: ReturnType<typeof window.setTimeout> | null = null;
  let reconnectTicker: ReturnType<typeof window.setInterval> | null = null;

  const isOpen = computed(() => status.value === "open");
  const isGuest = computed(() => Boolean(connected.value?.isGuest));
  const reconnectLabel = computed(() => {
    if (status.value !== "reconnecting" || reconnectingIn.value <= 0) {
      return "";
    }
    return `${reconnectingIn.value}s 后重连`;
  });

  const incomingEventHandlers: Partial<
    Record<ChatEventType, (event: ChatWsEvent) => void>
  > = {
    [chatEventType.WS_CONNECTED]: handleWsConnectedEvent,
    [chatEventType.SYSTEM_NOTICE]: handleSystemNoticeEvent,
    [chatEventType.ERROR]: handleErrorEvent,
    [chatEventType.EVALUATION_INVITE]: handleEvaluationInviteEvent,
    [chatEventType.EVALUATION_SUBMIT]: handleEvaluationSubmitEvent,
    [chatEventType.TYPING]: handleTypingEvent,
    [chatEventType.QUEUE_UPDATE]: handleQueueUpdateEvent,
    [chatEventType.AGENT_ACCEPTED]: handleAgentAcceptedEvent,
    [chatEventType.SESSION_CLOSE]: handleSessionCloseEvent,
    [chatEventType.MESSAGE]: handleMessageEvent,
  };

  // Connection lifecycle
  function connect() {
    manualClose = false;
    void loadChatOptions();
    lastOptions = {};
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

  function close() {
    manualClose = true;
    lastOptions = null;
    clearReconnectTimer();
    closeSocketOnly(false);
    connected.value = null;
    resetHistoryState();
    status.value = "idle";
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

  // Outbound events
  function sendText(content: string) {
    const text = content.trim();
    if (!canSendMessage(text)) {
      return;
    }
    sendUserMessageEvent(buildTextMessagePayload(text));
  }

  function canSendMessage(content: string) {
    return Boolean(
      socket.value &&
        socket.value.readyState === WebSocket.OPEN &&
        content &&
        !sessionClosed.value,
    );
  }

  function buildTextMessagePayload(
    content: string,
  ): SendUserMessagePayload {
    return {
      sessionNo: connected.value?.sessionNo || "",
      merchantId: connected.value?.merchantId || 0,
      messageType: 1,
      content,
      sender: {
        type: 1,
        id: connected.value?.userId || 0,
        nickname: connected.value?.nickname || "",
        avatarUrl: connected.value?.avatarUrl?.trim() || "",
      },
      receiver: {
        type: 2,
        id: agent.value?.agentUserId || 0,
        nickname: agent.value?.agentName || "",
        avatarUrl: agent.value?.agentAvatar?.trim() || "",
      }
    };
  }

  function sendUserMessageEvent(payload: SendUserMessagePayload) {
    if (!socket.value) return;
    sendChatSocketUserMessage(socket.value, payload);
  }

  async function endSession(closeReason = "user_closed", keepalive = false) {
    const current = connected.value;
    const canCloseSession = Boolean(
      current?.sessionNo && !current.isGuest && !sessionClosed.value,
    );

    if (canCloseSession) {
      try {
        await closeMyChatSession(closeReason, keepalive);
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

  // Incoming events
  function handleSocketMessage(payload: string) {
    const event = parseSocketEvent(payload);
    if (!event) return;
    dispatchIncomingEvent(event);
  }

  function parseSocketEvent(payload: string): ChatWsEvent | null {
    try {
      return JSON.parse(payload) as ChatWsEvent;
    } catch {
      error.value = "无效的消息格式";
      return null;
    }
  }

  function dispatchIncomingEvent(event: ChatWsEvent) {
    incomingEventHandlers[event.eventType]?.(event);
  }

  function handleWsConnectedEvent(event: ChatWsEvent) {
    const payload = event.connected as ChatWsEvent["connected"] & {
      connected?: WsConnectedPayload;
    };
    if (!payload.connected) return;
    connected.value = payload.connected;
  }

  function handleSystemNoticeEvent() {}

  function handleErrorEvent(event: ChatWsEvent) {
    error.value = event.error?.errorMessage || event.msg || "消息发送失败";
  }

  function handleEvaluationInviteEvent(event: ChatWsEvent) {
    queueStatus.value =
      event.evaluation?.comment || event.msg || "请对本次服务进行评价";
  }

  function handleEvaluationSubmitEvent(event: ChatWsEvent) {
    if (event.code && event.code !== 200) {
      error.value = event.msg || "评价提交失败";
      return;
    }
    queueStatus.value = "评价已提交";
  }

  function handleTypingEvent(event: ChatWsEvent) {
    if (event.typing?.senderType === 3) return;
    queueStatus.value = event.typing?.text || "客服正在输入";
  }

  function handleQueueUpdateEvent(event: ChatWsEvent) {
    const message = eventMessage(event);
    if (agent.value?.accept) {
      queueStatus.value = serviceStatusMessage(agent.value.agentName);
      return;
    }
    queueStatus.value =
      queueMessage(event.queue) || message?.content || "正在排队，客服会尽快接入。";
  }

  function handleAgentAcceptedEvent(event: ChatWsEvent) {
    const payload = event.agent as ChatWsEvent["agent"] & {
      agent?: ChatAgentPayload;
    };
    agent.value = payload
    agent.value.accept = true
  }

  function handleSessionCloseEvent(event: ChatWsEvent) {
    const message = eventMessage(event);
    sessionClosed.value = true;
    queueStatus.value = event.msg || message?.content || "本次会话已结束。";
    closeSocketOnly(true);
    status.value = "closed";
  }

  function handleMessageEvent(event: ChatWsEvent) {
    const message = eventMessage(event);
    if (!message) return;
    if (message.senderType === 2) {
      agentAccepted.value = true;
      activeAgentName.value =
        agentNameFromMessage(message) || activeAgentName.value;
      queueStatus.value = serviceStatusMessage(activeAgentName.value);
    }
    pushMessage(message);
  }

  function eventMessage(event: ChatWsEvent): ChatMessage | undefined {
    if (!event.message) return undefined;
    return normalizeMessage(event.message);
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

  async function loadHistory(initial = false) {
    const current = connected.value;
    if (
      !current?.sessionNo ||
      current.isGuest ||
      historyLoading.value ||
      (!initial && !historyHasMore.value)
    ) {
      return false;
    }

    historyLoading.value = true;
    try {
      const resp = await listChatMessagesWithMeta({
        cursor: initial ? 0 : historyNextCursor.value,
        limit: 20,
      });
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

  function optionValueMap(groupKey: string, fallback: Record<string, number>) {
    const group = optionGroups.value.find((item) => item.key === groupKey);
    if (!group?.options.length) return fallback;
    return group.options.reduce<Record<string, number>>(
      (map, item) => {
        map[item.code] = item.value;
        return map;
      },
      { ...fallback },
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

  function queueMessage(queue?: ChatQueuePayload) {
    if (!queue) return "";
    const position = Number(queue.queuePosition || 0);
    if (position > 1) {
      return `正在排队，您前面还有 ${position - 1} 人。`;
    }
    if (position === 1) return "您是当前队列第 1 位，客服即将接入。";
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
    const next = list.filter(
      (item) => item.messageNo && !seen.has(item.messageNo),
    );
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

  function normalizeMessage(message: Partial<ChatMessage>): ChatMessage {
    return {
      id: message.id,
      messageNo: message.messageNo ?? "",
      sessionNo: message.sessionNo ?? "",
      merchantId: message.merchantId ?? 0,
      senderType: normalizeSenderType(message.sender?.type),
      sender: message.sender,
      receiver: message.receiver,
      messageType: message.messageType ?? 0,
      content: message.content ?? "",
      url: message.url ?? "",
      fileName: message.fileName ?? "",
      fileSize: message.fileSize ?? 0,
      mimeType: message.mimeType ?? "",
      width: message.width ?? 0,
      height: message.height ?? 0,
      duration: message.duration ?? 0,
      status: message.status ?? 0,
      extra: message.extra ?? "",
      readTime: message.readTime ?? 0,
      createTime: message.createTime ?? 0,
      updateTime: message.updateTime ?? 0,
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
    isGuest,
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
