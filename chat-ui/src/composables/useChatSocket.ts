import { computed, ref } from "vue";
import {
  createChatSocket,
  getChatFileBlob,
  listChatMessagesWithMeta,
  options as loadOptions,
  sendChatSocketEvent,
  uploadChatFile,
} from "@/api/chat";
import { createChatWsRequest } from "@/types/chat";
import type {
  ChatAgentPayload,
  ChatErrorPayload,
  ChatEvaluationPayload,
  ChatMessageReceiptPayload,
  ChatMessageOperatePayload,
  ChatMessage,
  OptionGroup,
  ChatQueuePayload,
  ChatSession,
  ChatSystemNoticePayload,
  ChatTypingPayload,
  ChatWsEvent,
  WsConnectedPayload,
  SendUserMessagePayload,
} from "@/types/chat";
import { chatEventType, type ChatEventType } from "@/api/constant";

type ConnectOptions = Record<string, never>;

const reconnectDelays = [1000, 2000, 5000, 10000, 15000];
const messageStatus = {
  SENDING: 1,
  RECALLED: 6,
  DELETED: 7,
};
const messageOperateType = {
  RECALL: 1,
  DELETE: 2,
};
const messageDeleteScope = {
  BOTH: 2,
};
const recallWindowMs = 3 * 60 * 1000;

export function useChatSocket() {
  const socket = ref<WebSocket | null>(null);
  const connected = ref<WsConnectedPayload | null>(null);
  const messages = ref<ChatMessage[]>([]);
  const error = ref("");
  const queueStatus = ref("");
  const optionGroups = ref<OptionGroup[]>([]);
  const agent = ref<ChatAgentPayload | null>(null);
  const evaluationInvite = ref<ChatEvaluationPayload | null>(null);
  const evaluationSubmitting = ref(false);
  const evaluationSubmitted = ref(false);
  const agentAccepted = computed(() => Boolean(agent.value));
  const sessionClosed = ref(false);
  const historyLoading = ref(false);
  const historyHasMore = ref(false);
  const historyNextCursor = ref(0);
  const chatToken = ref("");
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
  const pendingUserMessageNos: Record<string, string[]> = {};

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
    [chatEventType.WS_CONNECTED]: (event) =>
      event.connected && handleWsConnectedEvent(event.connected),
    [chatEventType.SYSTEM_NOTICE]: (event) =>
      event.systemNotice && handleSystemNoticeEvent(event.systemNotice),
    [chatEventType.ERROR]: (event) =>
      event.error && handleErrorEvent(event.error),
    [chatEventType.EVALUATION_INVITE]: (event) =>
      event.evaluation && handleEvaluationInviteEvent(event.evaluation),
    [chatEventType.EVALUATION_SUBMIT]: (event) =>
      event.evaluation && handleEvaluationSubmitEvent(event.evaluation),
    [chatEventType.TYPING]: (event) =>
      event.typing && handleTypingEvent(event.typing),
    [chatEventType.QUEUE_UPDATE]: (event) =>
      event.queue && handleQueueUpdateEvent(event.queue),
    [chatEventType.AGENT_ACCEPTED]: (event) =>
      event.agent && handleAgentAcceptedEvent(event.agent),
    [chatEventType.SESSION_CLOSE]: (event) =>
      event.session && handleSessionCloseEvent(event.session),
    [chatEventType.MESSAGE]: (event) =>
      event.message && handleMessageEvent(event.message),
    [chatEventType.MESSAGE_DELIVERED]: (event) =>
      event.receipt && handleMessageReceiptEvent(event.receipt),
    [chatEventType.MESSAGE_RECALL]: (event) =>
      event.messageOperate && handleMessageOperateEvent(event.messageOperate),
    [chatEventType.MESSAGE_DELETE]: (event) =>
      event.messageOperate && handleMessageOperateEvent(event.messageOperate),
  };

  // Connection lifecycle
  function connect(token = "") {
    manualClose = false;
    chatToken.value = token.trim();
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
    chatToken.value = "";
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
    const payload = buildTextMessagePayload(text);
    sendUserMessagePayload(payload);
  }

  async function sendImage(file: File) {
    if (!canSendResourceMessage()) {
      return false;
    }
    if (!file.type.startsWith("image/")) {
      error.value = "请选择图片文件";
      return false;
    }
    try {
      const uploaded = await uploadChatFile(file, chatToken.value);
      const payload = buildImageMessagePayload({
        url: uploaded.url,
        fileName: uploaded.fileName || file.name,
        fileSize: uploaded.fileSize || file.size,
        mimeType: uploaded.mimeType || file.type,
      });
      sendUserMessagePayload(payload);
      return true;
    } catch (err) {
      error.value = err instanceof Error ? err.message : "图片发送失败";
      return false;
    }
  }

  async function resolveFileUrl(url: string) {
    if (!url.trim()) return "";
    if (!url.startsWith("/chat_uploads/")) return url;
    const blob = await getChatFileBlob(url, chatToken.value);
    return URL.createObjectURL(blob);
  }

  function sendUserMessagePayload(payload: SendUserMessagePayload) {
    const clientMessageId = `user-msg-${Date.now()}`;
    payload.clientMessageId = clientMessageId;
    sendUserMessageEvent(payload);
    pushMessage(buildOptimisticUserMessage(payload, clientMessageId));
    trackPendingUserMessage(payload.sessionNo || "", clientMessageId);
  }

  function canSendMessage(content: string) {
    return Boolean(
      socket.value &&
        socket.value.readyState === WebSocket.OPEN &&
        content &&
        agentAccepted.value &&
        !sessionClosed.value,
    );
  }

  function canSendResourceMessage() {
    return Boolean(
      socket.value &&
        socket.value.readyState === WebSocket.OPEN &&
        agentAccepted.value &&
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
      },
    };
  }

  function buildImageMessagePayload(file: {
    url: string;
    fileName: string;
    fileSize: number;
    mimeType: string;
  }): SendUserMessagePayload {
    return {
      ...buildTextMessagePayload(file.fileName),
      messageType: 2,
      content: file.fileName,
      url: file.url,
      fileName: file.fileName,
      fileSize: file.fileSize,
      mimeType: file.mimeType,
    };
  }

  function sendUserMessageEvent(payload: SendUserMessagePayload) {
    if (!socket.value) return;
    sendChatSocketEvent(
      socket.value,
      createChatWsRequest(chatEventType.MESSAGE, "message", payload, {
        requestId: payload.clientMessageId,
      }),
    );
  }

  function recallMessage(message: ChatMessage) {
    if (!canRecallMessage(message)) return false;
    return sendMessageOperate(message, messageOperateType.RECALL);
  }

  function deleteMessage(message: ChatMessage) {
    return sendMessageOperate(
      message,
      messageOperateType.DELETE,
      messageDeleteScope.BOTH,
    );
  }

  function sendMessageOperate(
    message: ChatMessage,
    operateType: number,
    deleteScope = 0,
  ) {
    if (!socket.value || socket.value.readyState !== WebSocket.OPEN) return false;
    if (!message.sessionNo || !message.messageNo) return false;
    const eventType =
      operateType === messageOperateType.RECALL
        ? chatEventType.MESSAGE_RECALL
        : chatEventType.MESSAGE_DELETE;
    sendChatSocketEvent(
      socket.value,
      createChatWsRequest(eventType, "messageOperate", {
        sessionNo: message.sessionNo,
        messageNo: message.messageNo,
        operateType,
        operatorId: connected.value?.userId || 0,
        operatorType: 1,
        deleteScope,
        operatedAt: Date.now(),
      }),
    );
    return true;
  }

  async function endSession(closeReason = "user_closed", keepalive = false) {
    if (!sessionClosed.value && socket.value?.readyState === WebSocket.OPEN) {
      sendChatSocketEvent(
        socket.value,
        createChatWsRequest(
          chatEventType.SESSION_CLOSE,
          "session",
          {
            merchantId: connected.value?.merchantId || 0,
            closeReason,
          },
        ),
      );
    }

    sessionClosed.value = true;
    queueStatus.value = "本次会话已结束。";
    if (keepalive) {
      closeSocketOnly(true);
      status.value = "closed";
    }
    return true;
  }

  async function submitEvaluation(
    rating: number,
    comment = "",
    tags: string[] = [],
  ) {
    if (
      evaluationSubmitting.value ||
      evaluationSubmitted.value ||
      !socket.value ||
      socket.value.readyState !== WebSocket.OPEN
    ) {
      return false;
    }
    evaluationSubmitting.value = true;
    error.value = "";
    try {
      sendChatSocketEvent(
        socket.value,
        createChatWsRequest(
          chatEventType.EVALUATION_SUBMIT,
          "evaluation",
          {
            sessionNo: connected.value?.sessionNo || "",
            userId: connected.value?.userId || 0,
            agentId: agent.value?.agentId || evaluationInvite.value?.agentId || 0,
            evaluationId: evaluationInvite.value?.evaluationId || 0,
            rating,
            tags,
            comment: comment.trim(),
            submitted: true,
            evaluatedAt: Date.now(),
          },
        ),
      );
      return true;
    } catch (err) {
      error.value = err instanceof Error ? err.message : "评价提交失败";
      return false;
    } finally {
      evaluationSubmitting.value = false;
    }
  }

  function resetMessages() {
    messages.value = [];
    resetHistoryState();
  }

  // Incoming events
  function handleSocketMessage(payload: string) {
    const event = parseSocketEvent(payload);
    if (!event) return;
    incomingEventHandlers[event.eventType]?.(event);
  }

  function parseSocketEvent(payload: string): ChatWsEvent | null {
    try {
      return JSON.parse(payload) as ChatWsEvent;
    } catch {
      error.value = "无效的消息格式";
      return null;
    }
  }

  function handleWsConnectedEvent(payload: WsConnectedPayload) {
    connected.value = payload;
    void loadHistory(true);
  }

  function handleSystemNoticeEvent(payload: ChatSystemNoticePayload) {
    if (payload.showInChat) {
      queueStatus.value = payload.content;
    }
  }

  function handleErrorEvent(payload: ChatErrorPayload) {
    error.value = payload.errorMessage || "消息发送失败";
  }

  function handleEvaluationInviteEvent(payload: ChatEvaluationPayload) {
    evaluationInvite.value = payload;
    evaluationSubmitted.value = false;
    queueStatus.value = payload.comment || "请对本次服务进行评价";
  }

  function handleEvaluationSubmitEvent(payload: ChatEvaluationPayload) {
    evaluationSubmitted.value = payload.submitted !== false;
    queueStatus.value = payload.submitted === false ? "评价提交失败" : "评价已提交";
  }

  function handleTypingEvent(payload: ChatTypingPayload) {
    if (payload.senderType === 3) return;
    queueStatus.value = payload.text || "客服正在输入";
  }

  function handleQueueUpdateEvent(payload: ChatQueuePayload) {
    let message = "";
    const position = Number(payload.queuePosition || 0);
    if (position > 1) {
      message = `正在排队，您前面还有 ${position - 1} 人。`;
    } else {
      message = "您是当前队列第 1 位，客服即将接入。";
    }
    queueStatus.value = message;
  }

  function handleAgentAcceptedEvent(payload: ChatAgentPayload) {
    agent.value = payload;
    let name = payload.agentName
    queueStatus.value = name ? `${name} 正在为你服务` : "客服正在为你服务";
  }

  function handleSessionCloseEvent(_payload: ChatSession) {
    sessionClosed.value = true;
    queueStatus.value = "本次会话已结束。";
  }

  function handleMessageEvent(payload: ChatMessage) {
    const message = normalizeMessage(payload);
    pushMessage(message);
  }

  function handleMessageReceiptEvent(payload: ChatMessageReceiptPayload) {
    if (!payload.messageNo) return;
    const message = messages.value.find(
      (item) => item.messageNo === payload.messageNo,
    );
    if (!message) {
      const pendingNo = shiftPendingUserMessage(payload.sessionNo);
      const pendingMessage = messages.value.find(
        (item) => item.messageNo === pendingNo,
      );
      if (!pendingMessage) return;
      pendingMessage.messageNo = payload.messageNo;
      pendingMessage.status = Number(
        payload.messageStatus || pendingMessage.status,
      );
      pendingMessage.updateTime = Number(
        payload.receiptTime || pendingMessage.updateTime,
      );
      return;
    }
    message.status = Number(payload.messageStatus || message.status);
    message.updateTime = Number(payload.receiptTime || message.updateTime);
  }

  function handleMessageOperateEvent(payload: ChatMessageOperatePayload) {
    if (!payload.sessionNo || !payload.messageNo) return;
    if (payload.operateType === messageOperateType.DELETE) {
      const message = messages.value.find(
        (item) => item.messageNo === payload.messageNo,
      );
      if (!message) return;
      markMessageDeleted(message, payload.operatedAt);
      return;
    }
    messages.value = messages.value.filter(
      (item) => item.messageNo !== payload.messageNo,
    );
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
      }, chatToken.value);
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

  function pushMessage(message: ChatMessage) {
    if (!message?.messageNo) {
      return;
    }
    if (message.status === messageStatus.RECALLED) return;
    if (messages.value.some((item) => item.messageNo === message.messageNo)) {
      return;
    }
    if (message.status === messageStatus.DELETED) {
      markMessageDeleted(message, message.updateTime);
    }
    messages.value.push(message);
  }

  function buildOptimisticUserMessage(
    payload: SendUserMessagePayload,
    messageNo: string,
  ): ChatMessage {
    const now = Date.now();
    return {
      messageNo,
      sessionNo: payload.sessionNo || "",
      merchantId: payload.merchantId || 0,
      senderType: 1,
      sender: payload.sender,
      receiver: payload.receiver,
      messageType: payload.messageType,
      content: payload.content || "",
      url: payload.url || "",
      fileName: payload.fileName || "",
      fileSize: payload.fileSize || 0,
      mimeType: payload.mimeType || "",
      width: payload.width || 0,
      height: payload.height || 0,
      duration: payload.duration || 0,
      status: messageStatus.SENDING,
      extra: payload.extra || "",
      readTime: 0,
      createTime: now,
      updateTime: now,
    };
  }

  function trackPendingUserMessage(sessionNo: string, messageNo: string) {
    if (!sessionNo || !messageNo) return;
    if (!pendingUserMessageNos[sessionNo]) {
      pendingUserMessageNos[sessionNo] = [];
    }
    pendingUserMessageNos[sessionNo].push(messageNo);
  }

  function shiftPendingUserMessage(sessionNo: string) {
    if (!sessionNo) return "";
    const pending = pendingUserMessageNos[sessionNo];
    if (!pending?.length) return "";
    return pending.shift() || "";
  }

  function prependMessages(list: ChatMessage[]) {
    const seen = new Set(messages.value.map((item) => item.messageNo));
    const next = list.filter(
      (item) =>
        item.messageNo &&
        !seen.has(item.messageNo) &&
        item.status !== messageStatus.RECALLED,
    );
    if (!next.length) {
      return;
    }
    next.forEach((item) => {
      if (item.status === messageStatus.DELETED) {
        markMessageDeleted(item, item.updateTime);
      }
    });
    messages.value = [...next, ...messages.value];
  }

  function canRecallMessage(message: ChatMessage) {
    const createTime = Number(message.createTime || message.updateTime || 0);
    return createTime > 0 && Date.now() - createTime <= recallWindowMs;
  }

  function markMessageDeleted(message: ChatMessage, operatedAt?: number) {
    message.status = messageStatus.DELETED;
    message.content = "";
    message.url = "";
    message.fileName = "";
    message.fileSize = 0;
    message.mimeType = "";
    message.width = 0;
    message.height = 0;
    message.duration = 0;
    message.updateTime = Number(operatedAt || Date.now());
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
      createTime: message.createTime ?? (message as { createTimes?: number }).createTimes ?? 0,
      updateTime: message.updateTime ?? (message as { updateTimes?: number }).updateTimes ?? 0,
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
    agent,
    agentAccepted,
    evaluationInvite,
    evaluationSubmitted,
    evaluationSubmitting,
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
    resolveFileUrl,
    recallMessage,
    deleteMessage,
    submitEvaluation,
    sendImage,
    sendText,
  };
}
