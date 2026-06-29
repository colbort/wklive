import { computed, ref } from "vue";
import {
  closeMyChatSession,
  createChatSocket,
  listChatMessagesWithMeta,
  options as loadOptions,
  sendChatSocketUserMessage,
  // sendChatSocketUserMessage,
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
import { chatEventType } from "@/api/constant";

type ConnectOptions = Record<string, never>;

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
  message_type?: number;
  url?: string;
  fileName?: string;
  file_name?: string;
  mimeType?: string;
  mime_type?: string;
  fileSize?: number;
  file_size?: number;
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
  const isTemporary = computed(() => Boolean(connected.value?.temporary));
  const reconnectLabel = computed(() => {
    if (status.value !== "reconnecting" || reconnectingIn.value <= 0) {
      return "";
    }
    return `${reconnectingIn.value}s 后重连`;
  });

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
    const canCloseSession = Boolean(
      current?.sessionNo && !current.temporary && !sessionClosed.value,
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
    switch (event.eventType) {
      case chatEventType.SYSTEM_NOTICE: {
        ensureConnectedFromEvent(event);
        break;
      }
      case chatEventType.ERROR: {
        const data = event.data as { message?: string };
        error.value = data?.message || "消息发送失败";
        break;
      }
      case chatEventType.MESSAGE_DELIVERED: {
        const message = eventMessage(event);
        if (message) pushMessage(message);
        break;
      }

      case chatEventType.EVALUATION_INVITE: {
        queueStatus.value =
          sessionEventMessage(event) ||
          eventMessage(event)?.content ||
          "请对本次服务进行评价";
        break;
      }

      case chatEventType.EVALUATION_SUBMIT: {
        const resp = event.data as WsResp<RawChatMessage> | undefined;
        if (resp && wsRespCode(resp) && wsRespCode(resp) !== 200) {
          error.value = wsRespMsg(resp) || "评价提交失败";
          return;
        }
        queueStatus.value = "评价已提交";
        break;
      }

      case chatEventType.TYPING: {
        if (eventMessage(event)?.senderType === 3) return;
        queueStatus.value =
          sessionEventMessage(event) ||
          eventMessage(event)?.content ||
          "客服正在输入";
        break;
      }

      case chatEventType.QUEUE_UPDATE: {
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
        break;
      }

      case chatEventType.AGENT_ACCEPTED: {
        ensureConnectedFromEvent(event);
        const message = eventMessage(event);
        agentAccepted.value = true;
        activeAgentName.value =
          agentNameFromMessage(message) || activeAgentName.value;
        queueStatus.value = serviceStatusMessage(
          activeAgentName.value,
          sessionEventMessage(event) || message?.content || "",
        );
        break;
      }

      case chatEventType.SESSION_CLOSE: {
        ensureConnectedFromEvent(event);
        const message = eventMessage(event);
        sessionClosed.value = true;
        queueStatus.value =
          sessionEventMessage(event) || message?.content || "本次会话已结束。";
        closeSocketOnly(true);
        status.value = "closed";
        break;
      }

      case chatEventType.MESSAGE: {
        ensureConnectedFromEvent(event);
        const message = eventMessage(event);
        if (!message) return;
        if (message.senderType === 2) {
          agentAccepted.value = true;
          activeAgentName.value =
            agentNameFromMessage(message) || activeAgentName.value;
          queueStatus.value = serviceStatusMessage(activeAgentName.value);
        }
        pushMessage(message);
        break;
      }

      default:
        if (!event.eventType && event.data && isConnectedPayload(event.data)) {
          console.log("==================== " + event.data)
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
        }
    }
  }

  function eventMessage(event: ChatWsEvent): ChatMessage | undefined {
    if (!event.data) return undefined;
    return normalizeMessage(event.data as RawChatMessage);
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
    return group.options.reduce<Record<string, number>>(
      (map, item) => {
        if (fallback[item.code] !== undefined) {
          map[item.code] = fallback[item.code];
        } else {
          map[item.code] = item.value;
        }
        return map;
      },
      { ...fallback },
    );
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

  function sessionEvent(event: ChatWsEvent): ChatSessionEvent | undefined {
    return event.sessionEvent ?? event.session_event;
  }

  function sessionEventMessage(event: ChatWsEvent) {
    return sessionEvent(event)?.message || "";
  }

  function ensureConnectedFromEvent(event: ChatWsEvent) {
    const session = event.session || sessionEvent(event)?.session;
    const queue = event.queue || sessionEvent(event)?.queue;
    const sessionNo =
      session?.sessionNo ||
      queue?.sessionNo ||
      eventMessage(event)?.sessionNo ||
      "";
    if (!sessionNo) return;
    const temporary = session ? sessionIsGuest(session) : true;
    const queueMerchantId =
      typeof queue?.merchantId === "number" ? queue.merchantId : 0;
    const queueUserId = Number(queue?.userId || 0);
    if (connected.value?.sessionNo) {
      connected.value = {
        ...connected.value,
        message:
          queueMessage(queue) ||
          sessionEventMessage(event) ||
          connected.value.message,
        merchantId:
          session?.merchantId ||
          queueMerchantId ||
          connected.value.merchantId,
        userId: session?.userId || queueUserId || connected.value.userId,
        temporary: session ? temporary : connected.value.temporary,
        session: session || connected.value.session,
        queue: queue || connected.value.queue,
      };
      if (
        session &&
        !temporary &&
        !historyLoading.value &&
        !messages.value.length
      ) {
        void loadHistory(true);
      }
      return;
    }
    connected.value = {
      message: queueMessage(queue) || sessionEventMessage(event) || "",
      merchantId: session?.merchantId || queueMerchantId || 0,
      userId: session?.userId || queueUserId || 0,
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

  function sessionIsGuest(session?: {
    extJson?: string | Record<string, unknown>;
  }) {
    if (!session?.extJson) return false;
    if (typeof session.extJson === "object") {
      return Boolean(session.extJson.isGuest);
    }
    try {
      return Boolean(
        (JSON.parse(session.extJson) as { isGuest?: boolean }).isGuest,
      );
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
    if ("message" in queue && queue.message) return queue.message;
    const position =
      "queuePosition" in queue
        ? Number(queue.queuePosition || 0)
        : Number(queue.position || 0);
    if (position > 1)
      return `正在排队，您前面还有 ${position - 1} 人。`;
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
      senderType: normalizeSenderType(message.sender?.type),
      sender: message.sender,
      receiver: message.receiver,
      messageType: message.messageType ?? message.message_type ?? 0,
      content: message.content ?? "",
      url: message.url ?? "",
      fileName: message.fileName ?? message.file_name ?? "",
      fileSize: message.fileSize ?? message.file_size ?? 0,
      mimeType: message.mimeType ?? message.mime_type ?? "",
      width: message.width ?? 0,
      height: message.height ?? 0,
      duration: message.duration ?? 0,
      status: message.status ?? 0,
      extra: message.extra ?? "",
      readTime: message.readTime ?? message.read_time ?? 0,
      createTime: message.createTime ?? message.create_times ?? 0,
      updateTime: message.updateTime ?? message.update_times ?? 0,
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
