<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { ElMessage } from "element-plus";
import {
  chatAdminWsUrl,
  getChatFileBlob,
  options as loadOptions,
  pageMessages,
  pageSessions,
  updateAgentStatus,
  uploadChatFile,
} from "@/api/chat";
import { chatEventType, type ChatEventType } from "@/api/constant";
import WorkbenchChatPanel from "@/components/workbench/WorkbenchChatPanel.vue";
import WorkbenchCustomerPanel from "@/components/workbench/WorkbenchCustomerPanel.vue";
import WorkbenchSessionList from "@/components/workbench/WorkbenchSessionList.vue";
import { useAuthStore } from "@/stores/auth";
import type {
  AcceptChatSessionPayload,
  ChatAgentPayload,
  ChatWsEvent,
  ChatWsRequest,
  ChatAgent,
  ChatEvaluationPayload,
  ChatMessage,
  ChatMessageOperatePayload,
  ChatMessageReceiptPayload,
  ChatQueuePayload,
  ChatSession,
  ChatSessionExtJson,
  ChatSystemNoticePayload,
  ChatTypingPayload,
  ChatUserStatePayload,
  CloseAgentChatSessionPayload,
  SendAgentMessagePayload,
  WsConnectedPayload,
} from "@/types/chat";
import { createChatWsRequest } from "@/types/chat";
import {
  optionGroup,
  withOptionLabels,
  type DisplayOptionItem,
} from "@/utils/options";

const statusFilter = ref("waiting");
const connected = ref<WsConnectedPayload | null>(null);
const activeSessionNo = ref("");
const loadingSessions = ref(false);
const loadingMessages = ref(false);
const mobileChatOpen = ref(false);
const sessions = ref<ChatSession[]>([]);
const messages = ref<Record<string, ChatMessage[]>>({});
const sessionStatusMessages = ref<Record<string, string>>({});
const wsState = ref<"idle" | "open" | "closed">("idle");
const changingAgentStatus = ref(false);
const pendingAgentMessageNos: Record<string, string[]> = {};
const messageStatus = {
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
const defaultAgentStatusOptions: DisplayOptionItem[] = [
  {
    code: "CHAT_AGENT_STATUS_OFFLINE",
    label: "离线",
    value: 1,
    tagType: "info",
  },
  {
    code: "CHAT_AGENT_STATUS_ONLINE",
    label: "在线",
    value: 2,
    tagType: "success",
  },
  {
    code: "CHAT_AGENT_STATUS_BUSY",
    label: "忙碌",
    value: 3,
    tagType: "warning",
  },
  {
    code: "CHAT_AGENT_STATUS_RESTING",
    label: "休息",
    value: 4,
    tagType: "info",
  },
  {
    code: "CHAT_AGENT_STATUS_AWAY",
    label: "暂离",
    value: 5,
    tagType: "info",
  },
];
const agentStatusOptions = ref<DisplayOptionItem[]>(defaultAgentStatusOptions);
const auth = useAuthStore();
let socket: WebSocket | null = null;
let refreshTimer: number | null = null;
let reconnectTimer: number | null = null;
let reconnectTimes = 0;
let destroyed = false;

const sessionStatus = {
  waiting: 1,
  serving: 2,
  pendingUser: 3,
  pendingAgent: 4,
  closed: 5,
} as const;

const agentStatus = {
  offline: 1,
  online: 2,
  busy: 3,
  resting: 4,
} as const;

const filteredSessions = computed(() =>
  sessions.value.filter((session) => matchStatusFilter(session)),
);

const activeSession = computed(
  () =>
    sessions.value.find((item) => item.sessionNo === activeSessionNo.value) ||
    filteredSessions.value[0],
);
const activeMessages = computed(() =>
  activeSession.value
    ? messages.value[activeSession.value.sessionNo] || []
    : [],
);
const visibleMessages = computed(() =>
  activeMessages.value.filter((message) => !isQueueSystemMessage(message)),
);
const merchantId = computed(
  () => auth.user?.merchantId || auth.agent?.merchantId || 0,
);
const agentId = computed(() => auth.agent?.id || 0);
const userId = computed(() => auth.user?.id || 0)
const agentStatusOption = computed(
  () =>
    agentStatusOptions.value.find(
      (item) => item.value === auth.agent?.status,
    ) || defaultAgentStatusOptions[0],
);
const wsOnline = computed(() => wsState.value === "open");
const activeNeedsAccept = computed(
  () =>
    Boolean(activeSession.value) &&
    needsAccept(activeSession.value) &&
    !activeSession.value?.agentId,
);
const activeClosed = computed(
  () => activeSession.value?.status === sessionStatus.closed,
);
const activeIsGuest = computed(() => isGuestSession(activeSession.value));
const showGuestRefreshNotice = computed(
  () =>
    Boolean(activeSession.value) &&
    activeIsGuest.value &&
    !visibleMessages.value.length &&
    Boolean(activeSession.value?.lastMessageNo),
);
const canReply = computed(
  () =>
    Boolean(activeSession.value) &&
    !activeNeedsAccept.value &&
    !activeClosed.value,
);
const agentCanAccept = computed(
  () =>
    wsOnline.value &&
    Boolean(agentId.value) &&
    auth.agent?.status === agentStatus.online,
);
const acceptDisabledReason = computed(() => {
  if (!wsOnline.value) return "连接断开，暂时不能接待";
  if (!agentId.value) return "当前账号不是坐席，不能接待";
  if (auth.agent?.status !== agentStatus.online)
    return "坐席在线后才能接待会话";
  return "";
});

const incomingWsHandlers: Partial<
  Record<ChatEventType, (event: ChatWsEvent) => void>
> = {
  [chatEventType.WS_CONNECTED]: (event) =>
    event.connected && handleWsConnectedWsEvent(event.connected),
  [chatEventType.MESSAGE]: (event) =>
    event.message && handleMessageWsEvent(event.message),
  [chatEventType.MESSAGE_DELIVERED]: (event) =>
    event.receipt && handleMessageReceiptWsEvent(event.receipt),
  [chatEventType.MESSAGE_RECALL]: (event) =>
    event.messageOperate && handleMessageOperateWsEvent(event.messageOperate),
  [chatEventType.MESSAGE_DELETE]: (event) =>
    event.messageOperate && handleMessageOperateWsEvent(event.messageOperate),
  [chatEventType.SYSTEM_NOTICE]: (event) =>
    event.systemNotice && handleSystemNoticeWsEvent(event.systemNotice),
  [chatEventType.USER_JOIN]: (event) => handleUserJoinWsEvent(event),
  [chatEventType.USER_LEAVE]: (event) =>
    event.userState && handleUserLeaveWsEvent(event.userState),
  [chatEventType.QUEUE_UPDATE]: (event) =>
    event.queue && handleQueueUpdateWsEvent(event.queue),
  [chatEventType.AGENT_ACCEPTED]: (event) =>
    event.agent && handleAgentAcceptedWsEvent(event.agent),
  [chatEventType.AGENT_JOIN]: (event) =>
    event.agent && handleAgentStateWsEvent(event.agent),
  [chatEventType.AGENT_LEAVE]: (event) =>
    event.agent && handleAgentStateWsEvent(event.agent),
  [chatEventType.SESSION_CLOSE]: (event) =>
    event.session && handleSessionCloseWsEvent(event.session),
  [chatEventType.EVALUATION_SUBMIT]: (event) =>
    event.evaluation && handleEvaluationSubmitWsEvent(event.evaluation),
  [chatEventType.TYPING]: (event) =>
    event.typing && handleTypingWsEvent(event.typing),
  [chatEventType.ERROR]: (event) =>
    event.error && ElMessage.error(event.error.errorMessage || "操作失败"),
};

onMounted(async () => {
  destroyed = false;
  restoreWorkbenchState();
  void loadAdminOptions();
  await loadSessions(true);
  connectWs();
});

onBeforeUnmount(() => {
  destroyed = true;
  if (refreshTimer) {
    window.clearTimeout(refreshTimer);
  }
  clearReconnectTimer();
  if (socket) {
    socket.onclose = null;
    socket.close();
  }
});

watch(statusFilter, async () => {
  persistWorkbenchState();
  syncActiveSession();
});

watch(activeSessionNo, async (sessionNo) => {
  persistWorkbenchState();
  if (sessionNo) {
    await loadMessages(sessionNo);
  }
});

watch(activeNeedsAccept, async (needsAccept) => {
  if (!needsAccept && activeSessionNo.value) {
    await loadMessages(activeSessionNo.value);
  }
});

function workbenchStorageKey() {
  return `chat-admin-ui:workbench:${merchantId.value}:${agentId.value}`;
}

function restoreWorkbenchState() {
  if (!merchantId.value || !agentId.value) return;
  try {
    const raw = window.sessionStorage.getItem(workbenchStorageKey());
    if (!raw) return;
    const state = JSON.parse(raw) as {
      statusFilter?: string;
      activeSessionNo?: string;
    };
    if (
      state.statusFilter === "waiting" ||
      state.statusFilter === "serving" ||
      state.statusFilter === "closed"
    ) {
      statusFilter.value = state.statusFilter;
    }
    if (typeof state.activeSessionNo === "string") {
      activeSessionNo.value = state.activeSessionNo;
    }
  } catch {
    window.sessionStorage.removeItem(workbenchStorageKey());
  }
}

function persistWorkbenchState() {
  if (!merchantId.value || !agentId.value) return;
  window.sessionStorage.setItem(
    workbenchStorageKey(),
    JSON.stringify({
      statusFilter: statusFilter.value,
      activeSessionNo: activeSessionNo.value,
    }),
  );
}

async function loadAdminOptions() {
  try {
    const resp = await loadOptions();
    const agentStatuses = optionGroup(resp.data.options, "chatAgentStatus");
    if (agentStatuses.length) {
      agentStatusOptions.value = withOptionLabels(agentStatuses);
    }
  } catch {
    agentStatusOptions.value = defaultAgentStatusOptions;
  }
}

async function changeAgentStatus(status: number) {
  if (!auth.agent || auth.agent.status === status) return;
  changingAgentStatus.value = true;
  try {
    const resp = await updateAgentStatus(auth.agent.id, { status });
    auth.agent = resp.data;
    ElMessage.success("状态已更新");
  } finally {
    changingAgentStatus.value = false;
  }
}

async function loadSessions(loadActiveMessages = false) {
  if (!merchantId.value) return;
  loadingSessions.value = true;
  try {
    const resp = await pageSessions({
      merchantId: merchantId.value,
      limit: 50,
    });
    const previousActiveSessionNo = activeSessionNo.value;
    sessions.value = collapseGuestSessions(resp.data.map(normalizeSession));
    syncActiveSession();
    if (
      loadActiveMessages &&
      activeSessionNo.value &&
      activeSessionNo.value === previousActiveSessionNo
    ) {
      await loadMessages(activeSessionNo.value);
    }
  } catch (err) {
    ElMessage.error(err instanceof Error ? err.message : "加载会话失败");
  } finally {
    loadingSessions.value = false;
  }
}

async function loadMessages(sessionNo: string) {
  if (!merchantId.value) return;
  const session = sessions.value.find((item) => item.sessionNo === sessionNo);
  if (sessionNeedsAccept(session)) {
    return;
  }
  loadingMessages.value = true;
  try {
    const resp = await pageMessages(sessionNo, {
      merchantId: merchantId.value,
      isGuest: session?.isGuest||false,
      limit: 50,
    });
    messages.value[sessionNo] = sortMessages(
      resp.data
        .map(normalizeMessage)
        .filter(
          (message) =>
            message.status !== messageStatus.RECALLED &&
            !isQueueSystemMessage(message),
        ),
    );
  } catch (err) {
    ElMessage.error(err instanceof Error ? err.message : "加载消息失败");
  } finally {
    loadingMessages.value = false;
  }
}

function connectWs() {
  if (!auth.token || !merchantId.value) return;
  clearReconnectTimer();
  if (socket) {
    socket.onclose = null;
    socket.close();
  }
  socket = new WebSocket(
    chatAdminWsUrl({
      token: auth.token,
      merchantId: merchantId.value,
      agentId: agentId.value,
    }),
  );
  socket.onopen = () => {
    wsState.value = "open";
    reconnectTimes = 0;
  };
  socket.onclose = () => {
    wsState.value = "closed";
    scheduleReconnect();
  };
  socket.onerror = () => {
    wsState.value = "closed";
  };
  socket.onmessage = (message: MessageEvent) => {
    handleWsMessage(message.data);
  };
}

function scheduleReconnect() {
  if (destroyed || !auth.token || !merchantId.value) return;
  clearReconnectTimer();
  const delays = [1000, 2000, 5000, 10000, 15000];
  const delay = delays[Math.min(reconnectTimes, delays.length - 1)];
  reconnectTimes += 1;
  reconnectTimer = window.setTimeout(() => {
    connectWs();
  }, delay);
}

function clearReconnectTimer() {
  if (!reconnectTimer) return;
  window.clearTimeout(reconnectTimer);
  reconnectTimer = null;
}

function handleWsMessage(payload: string) {
  const event = parseWsEvent(payload);
  if (!event) return;
  dispatchWsEvent(event);
}

function parseWsEvent(payload: string): ChatWsEvent | null {
  try {
    return JSON.parse(payload) as ChatWsEvent;
  } catch {
    return null;
  }
}

function dispatchWsEvent(event: ChatWsEvent) {
  if (event.code && event.code !== 200) {
    ElMessage.error(event.msg || event.error?.errorMessage || "发送失败");
    return;
  }
  incomingWsHandlers[event.eventType]?.(event);
}

function handleAgentStateWsEvent(payload: ChatAgentPayload) {
  updateAgentFromEvent(payload);
  scheduleRefreshSessions();
}

function handleAgentAcceptedWsEvent(payload: ChatAgentPayload) {
  if (payload.sessionNo) {
    setSessionStatusMessage(payload.sessionNo, payload.remark || "");
    if (Number(payload.agentId || agentId.value) === agentId.value) {
      statusFilter.value = "serving";
      activeSessionNo.value = payload.sessionNo;
    }
  }
  scheduleRefreshSessions();
}

function handleSessionCloseWsEvent(payload: ChatSession) {
  if (payload.sessionNo) {
    setSessionStatusMessage(
      payload.sessionNo,
      payload.closeReason || "本次会话已结束",
    );
    if (activeSessionNo.value === payload.sessionNo) {
      statusFilter.value = "closed";
    }
  }
  scheduleRefreshSessions();
}

function handleQueueUpdateWsEvent(payload: ChatQueuePayload) {
  if (payload.sessionNo) {
    setSessionStatusMessage(payload.sessionNo, queueStatusMessage(payload));
    focusWaitingSession(payload.sessionNo);
  }
  scheduleRefreshSessions();
}

function handleUserJoinWsEvent(event: ChatWsEvent) {
  const sessionNo = event.session?.sessionNo || event.userState?.sessionNo || "";
  if (!sessionNo) {
    scheduleRefreshSessions();
    return;
  }
  focusWaitingSession(sessionNo);
  scheduleRefreshSessions();
}

function handleUserLeaveWsEvent(payload: ChatUserStatePayload) {
  if (payload.sessionNo) {
    setSessionStatusMessage(payload.sessionNo, payload.userName || "用户已离开");
  }
  scheduleRefreshSessions();
}

function handleMessageWsEvent(payload: ChatMessage) {
  console.log("收到消息，消息编号：" + payload.messageNo)
  applyWsSessionMessage(payload);
}

function handleSystemNoticeWsEvent(payload: ChatSystemNoticePayload) {
  const content = payload.content || payload.title || "";
  if (!content) return;
  if (payload.level === "error") {
    ElMessage.error(content);
    return;
  }
  if (payload.level === "warning") {
    ElMessage.warning(content);
    return;
  }
  ElMessage.info(content);
}

function handleMessageReceiptWsEvent(payload: ChatMessageReceiptPayload) {
  applyMessageReceipt(payload);
}

function handleEvaluationSubmitWsEvent(payload: ChatEvaluationPayload) {
  if (!payload.sessionNo) return;
  setSessionStatusMessage(payload.sessionNo, payload.submitted === false ? "评价提交失败" : "用户已提交评价");
}

function handleTypingWsEvent(payload: ChatTypingPayload) {
  if (!payload.sessionNo) return;
  setSessionStatusMessage(payload.sessionNo, payload.text || "用户正在输入");
}

function handleWsConnectedWsEvent(payload: WsConnectedPayload) {
  connected.value = payload;
}

function applyWsSessionMessage(message: ChatMessage) {
  if (message.sessionNo && message.messageNo) {
    normalizeMessageEnums(message);
    upsertSessionFromMessage(message);
    if (shouldAppendWsMessage(message)) {
      pushMessage(message);
    }
  }
  const session = message.sessionNo
    ? sessions.value.find((item) => item.sessionNo === message.sessionNo)
    : undefined;
  if (
    message.senderType === 1 &&
    !session?.agentId &&
    statusFilter.value !== "waiting"
  ) {
    statusFilter.value = "waiting";
  }
  if (message.senderType === 1 && message.sessionNo) {
    activeSessionNo.value = message.sessionNo;
  }
  scheduleRefreshSessions();
  return message;
}

function focusWaitingSession(sessionNo: string) {
  if (!mobileChatOpen.value || sessionNo !== activeSessionNo.value) {
    statusFilter.value = "waiting";
    mobileChatOpen.value = false;
  }
}

function updateAgentFromEvent(agent?: ChatAgentPayload) {
  if (!agent || agent.agentId !== agentId.value) return;
  const nextStatus = Number(agent.agentStatus || 0);
  auth.agent = {
    ...(auth.agent || ({} as ChatAgent)),
    id: agent.agentId,
    status: nextStatus || auth.agent?.status || 0,
    welcomeMessage: agent.remark || auth.agent?.welcomeMessage || "",
    updateTimes: agent.actionTime || auth.agent?.updateTimes || 0,
  };
}

function scheduleRefreshSessions() {
  if (refreshTimer) {
    window.clearTimeout(refreshTimer);
  }
  refreshTimer = window.setTimeout(() => {
    void loadSessions();
  }, 250);
}

function pushMessage(message: ChatMessage) {
  if (!message.messageNo || isQueueSystemMessage(message)) return false;
  if (message.status === messageStatus.RECALLED) return false;
  if (message.status === messageStatus.DELETED) {
    markMessageDeleted(message, message.updateTime);
  }
  const list = messages.value[message.sessionNo] || [];
  if (list.some((item) => item.messageNo === message.messageNo)) return false;
  messages.value[message.sessionNo] = sortMessages([...list, message]);
  return true;
}

function applyMessageReceipt(payload: ChatMessageReceiptPayload) {
  if (!payload.sessionNo || !payload.messageNo) return;
  const list = messages.value[payload.sessionNo] || [];
  let message = list.find((item) => item.messageNo === payload.messageNo);
  if (!message) {
    const pendingNo = pendingAgentMessageNos[payload.sessionNo]?.shift();
    message = pendingNo
      ? list.find((item) => item.messageNo === pendingNo)
      : undefined;
    if (message) {
      message.messageNo = payload.messageNo;
    }
  }
  if (!message) return;
  message.status = Number(payload.messageStatus || message.status);
  message.updateTime = Number(payload.receiptTime || message.updateTime);
  messages.value[payload.sessionNo] = [...list];
}

function handleMessageOperateWsEvent(payload: ChatMessageOperatePayload) {
  if (!payload.sessionNo || !payload.messageNo) return;
  const list = messages.value[payload.sessionNo] || [];
  if (payload.operateType === messageOperateType.DELETE) {
    const message = list.find((item) => item.messageNo === payload.messageNo);
    if (!message) return;
    markMessageDeleted(message, payload.operatedAt);
    messages.value[payload.sessionNo] = [...list];
    return;
  }
  messages.value[payload.sessionNo] = list.filter(
    (item) => item.messageNo !== payload.messageNo,
  );
}

function trackPendingAgentMessage(sessionNo: string, messageNo: string) {
  if (!sessionNo || !messageNo) return;
  pendingAgentMessageNos[sessionNo] = [
    ...(pendingAgentMessageNos[sessionNo] || []),
    messageNo,
  ];
}

function isGuestSession(session?: ChatSession) {
  if (session?.isGuest) return true;
  return Boolean(session?.extJson?.isGuest);
}

function collapseGuestSessions(list: ChatSession[]) {
  const selected = new Map<string, ChatSession>();
  for (const session of list) {
    const key = guestSessionKey(session);
    if (!key) continue;
    const exists = selected.get(key);
    if (!exists || sessionSortTime(session) >= sessionSortTime(exists)) {
      selected.set(key, session);
    }
  }
  if (!selected.size) return list;
  return list.filter((session) => {
    const key = guestSessionKey(session);
    return !key || selected.get(key)?.sessionNo === session.sessionNo;
  });
}

function guestSessionKey(session?: ChatSession) {
  if (!session || !isGuestSession(session) || !session.userId) return "";
  return `${session.merchantId || merchantId.value}:${session.userId}`;
}

function sessionSortTime(session: ChatSession) {
  return Math.max(
    Number(session.lastMessageTime || 0),
    Number(session.updateTimes || 0),
    Number(session.createTimes || 0),
  );
}

function setSessionStatusMessage(sessionNo: string, message?: string) {
  const text = message?.trim();
  if (!sessionNo || !text) return;
  sessionStatusMessages.value = {
    ...sessionStatusMessages.value,
    [sessionNo]: text,
  };
}

function queueStatusMessage(queue?: { queuePosition?: number; waitingCount?: number }) {
  if (!queue) return "";
  const position = Number(queue.queuePosition || 0);
  if (position > 1) return `正在排队，前面还有 ${position - 1} 人`;
  if (position === 1) return "当前队列第 1 位";
  const waitingCount = Number(queue.waitingCount || 0);
  return waitingCount > 0 ? `当前等待 ${waitingCount} 人` : "";
}

function matchStatusFilter(session: ChatSession) {
  const status = sessionStatusValue(session.status);
  if (statusFilter.value === "waiting") {
    return status === sessionStatus.waiting;
  }
  if (statusFilter.value === "serving") {
    return (
      (
        [
          sessionStatus.serving,
          sessionStatus.pendingUser,
          sessionStatus.pendingAgent,
        ] as number[]
      ).includes(status)
    );
  }
  return status === sessionStatus.closed;
}

function sessionNeedsAccept(session?: ChatSession) {
  return needsAccept(session);
}

function needsAccept(session?: ChatSession) {
  const status = sessionStatusValue(session?.status);
  return (
    Boolean(session) &&
    (status === sessionStatus.waiting ||
      status === sessionStatus.pendingAgent) &&
    !Number(session?.agentId || 0)
  );
}

function normalizeMessageEnums(message: ChatMessage) {
  const raw = message as ChatMessage & {
    createTimes?: number;
    updateTimes?: number;
  };
  message.senderType = senderTypeValue(message.senderType || message.sender?.type);
  message.createTime = Number(message.createTime || raw.createTimes || 0);
  message.updateTime = Number(message.updateTime || raw.updateTimes || 0);
}

function normalizeMessage(message: ChatMessage) {
  normalizeMessageEnums(message);
  if (message.status === messageStatus.DELETED) {
    markMessageDeleted(message, message.updateTime);
  }
  return message;
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

function sortMessages(list: ChatMessage[]) {
  return [...list].sort(compareMessages);
}

function compareMessages(left: ChatMessage, right: ChatMessage) {
  const leftTime = Number(left.createTime || 0);
  const rightTime = Number(right.createTime || 0);
  if (leftTime !== rightTime) return leftTime - rightTime;

  const leftId = Number(left.id || 0);
  const rightId = Number(right.id || 0);
  if (leftId !== rightId) return leftId - rightId;

  return String(left.messageNo || "").localeCompare(String(right.messageNo || ""));
}

function isQueueSystemMessage(message?: ChatMessage) {
  return (
    senderTypeValue(message?.senderType) === 3 &&
    typeof message?.content === "string" &&
    message.content.includes("正在排队")
  );
}

function shouldAppendWsMessage(message?: ChatMessage) {
  if (!message || isQueueSystemMessage(message)) return false;
  return true;
}

function upsertSessionFromMessage(message: ChatMessage) {
  if (!message.sessionNo) return;
  const index = sessions.value.findIndex(
    (item) => item.sessionNo === message.sessionNo,
  );
  const current = index >= 0 ? sessions.value[index] : undefined;
  const now = Number(message.createTime || message.updateTime || Date.now());
  const user = message.senderType === 1 ? message.sender : message.receiver;
  const next: ChatSession = normalizeSession({
    id: current?.id || 0,
    sessionNo: message.sessionNo,
    merchantId: current?.merchantId || message.merchantId || merchantId.value,
    userId: current?.userId || Number(user?.id || 0),
    source: current?.source || 0,
    status: current?.status || sessionStatus.waiting,
    priority: current?.priority || 0,
    agentId: current?.agentId || 0,
    groupId: current?.groupId || 0,
    title: current?.title || "",
    category: current?.category || "",
    lastMessage: message.content || message.fileName || "[图片]",
    lastSenderType: message.senderType,
    lastMessageTime: now,
    userUnreadCount:
      current?.userUnreadCount ||
      (activeSessionNo.value === message.sessionNo ? 0 : 1),
    agentUnreadCount: current?.agentUnreadCount || 0,
    closeTime: current?.closeTime || 0,
    closeReason: current?.closeReason || "",
    extJson: {
      ...(current?.extJson || {}),
      nickname: current?.extJson?.nickname || user?.nickname || "",
      avatarUrl: current?.extJson?.avatarUrl || user?.avatarUrl || "",
    },
    isGuest: current?.isGuest || false,
    avatarUrl: current?.avatarUrl || user?.avatarUrl || "",
    lastMessageNo: message.messageNo,
    createTimes: current?.createTimes || now,
    updateTimes: now,
  });

  if (index >= 0) {
    sessions.value[index] = next;
  } else {
    sessions.value = [next, ...sessions.value];
  }
}

function normalizeSession(session: ChatSession) {
  session.id = Number(session.id || 0);
  session.merchantId = Number(session.merchantId || 0);
  session.status = sessionStatusValue(session.status);
  session.source = Number(session.source || 0);
  session.priority = Number(session.priority || 0);
  session.agentId = Number(session.agentId || 0);
  session.userId = Number(session.userId || 0);
  session.groupId = Number(session.groupId || 0);
  session.lastSenderType = senderTypeValue(session.lastSenderType);
  session.lastMessageTime = Number(session.lastMessageTime || 0);
  session.userUnreadCount = Number(session.userUnreadCount || 0);
  session.agentUnreadCount = Number(session.agentUnreadCount || 0);
  session.closeTime = Number(session.closeTime || 0);
  session.createTimes = Number(session.createTimes || 0);
  session.updateTimes = Number(session.updateTimes || 0);
  session.extJson = normalizeSessionExtJson(session);
  session.isGuest = Boolean(session.isGuest || session.extJson.isGuest);
  return session;
}

function normalizeSessionExtJson(session: ChatSession): ChatSessionExtJson {
  const ext = parseSessionExtJson(session.extJson);
  const nickname = firstString(
    ext.nickname,
  );
  const avatarUrl = firstString(
    ext.avatarUrl,
  );
  return {
    ...ext,
    nickname,
    avatarUrl,
  };
}

function parseSessionExtJson(value: unknown): ChatSessionExtJson {
  if (!value) return {};
  if (typeof value === "object" && !Array.isArray(value)) {
    return value as ChatSessionExtJson;
  }
  if (typeof value !== "string") return {};
  try {
    const parsed = JSON.parse(value);
    if (parsed && typeof parsed === "object" && !Array.isArray(parsed)) {
      return parsed as ChatSessionExtJson;
    }
  } catch {
    // ignore invalid ext_json
  }
  return {};
}

function firstString(...values: unknown[]) {
  for (const value of values) {
    if (typeof value !== "string") continue;
    const trimmed = value.trim();
    if (trimmed) return trimmed;
  }
  return "";
}

function sessionStatusValue(value: unknown) {
  if (typeof value === "number") return value;
  if (value === "CHAT_SESSION_STATUS_WAITING") return sessionStatus.waiting;
  if (value === "CHAT_SESSION_STATUS_SERVING") return sessionStatus.serving;
  if (value === "CHAT_SESSION_STATUS_PENDING_USER")
    return sessionStatus.pendingUser;
  if (value === "CHAT_SESSION_STATUS_PENDING_AGENT")
    return sessionStatus.pendingAgent;
  if (value === "CHAT_SESSION_STATUS_CLOSED") return sessionStatus.closed;
  const numeric = Number(value);
  return Number.isFinite(numeric) ? numeric : 0;
}

function senderTypeValue(value: unknown) {
  if (typeof value === "number") return value;
  if (value === "CHAT_SENDER_TYPE_USER") return 1;
  if (value === "CHAT_SENDER_TYPE_AGENT") return 2;
  if (value === "CHAT_SENDER_TYPE_SYSTEM") return 3;
  return 0;
}

function syncActiveSession() {
  const list = filteredSessions.value;
  if (!list.length) {
    activeSessionNo.value = "";
    return;
  }
  if (!list.some((item) => item.sessionNo === activeSessionNo.value)) {
    activeSessionNo.value = list[0].sessionNo;
  }
}

async function refreshSessions() {
  await loadSessions();
}

function selectSession(sessionNo: string) {
  activeSessionNo.value = sessionNo;
  mobileChatOpen.value = true;
}

function backToSessions() {
  mobileChatOpen.value = false;
}

function sendWsEvent(request: ChatWsRequest) {
  if (!socket || socket.readyState !== WebSocket.OPEN) {
    return false;
  }
  socket.send(JSON.stringify(request));
  return true;
}

function send(value: string) {
  const content = value.trim();
  if (!canSendMessage(content)) {
    return;
  }
  const clientMessageId = `agent-msg-${Date.now()}`;
  return sendAgentMessagePayload({
    clientMessageId,
    merchantId: merchantId.value,
    sessionNo: activeSession.value?.sessionNo || "",
    messageType: 1,
    content,
    isGuest: activeIsGuest.value,
    sender: {
      type: 2,
      id: auth.user?.id || 0,
      nickname: auth.user?.nickname || "",
      avatarUrl: auth.user?.avatarUrl || "",
    },
  });
}

async function sendImage(file: File) {
  if (!canSendResourceMessage()) {
    return false;
  }
  if (!file.type.startsWith("image/")) {
    ElMessage.warning("请选择图片文件");
    return false;
  }
  try {
    const resp = await uploadChatFile(file);
    const data = resp.data || resp.Data;
    if (!data?.url) {
      ElMessage.error("图片上传失败");
      return false;
    }
    const clientMessageId = `agent-msg-${Date.now()}`;
    return sendAgentMessagePayload({
      clientMessageId,
      merchantId: merchantId.value,
      sessionNo: activeSession.value?.sessionNo || "",
      messageType: 2,
      content: data.fileName || file.name,
      url: data.url,
      fileName: data.fileName || file.name,
      fileSize: data.fileSize || file.size,
      mimeType: data.mimeType || file.type,
      isGuest: activeIsGuest.value,
      sender: {
        type: 2,
        id: auth.user?.id || 0,
        nickname: auth.user?.nickname || "",
        avatarUrl: auth.user?.avatarUrl || "",
      },
    });
  } catch (err) {
    ElMessage.error(err instanceof Error ? err.message : "图片发送失败");
    return false;
  }
}

async function resolveFileUrl(url: string) {
  if (!url.trim()) return "";
  if (!url.startsWith("/chat_uploads/")) return url;
  const blob = await getChatFileBlob(url);
  return URL.createObjectURL(blob);
}

function sendAgentMessagePayload(payload: SendAgentMessagePayload) {
  const clientMessageId = payload.clientMessageId || `agent-msg-${Date.now()}`;
  payload.clientMessageId = clientMessageId;
  const sent = sendWsEvent(
    createChatWsRequest(chatEventType.MESSAGE, "message", payload, {
      requestId: clientMessageId,
    }),
  );
  if (sent) {
    pushMessage(buildOptimisticAgentMessage(payload, clientMessageId));
    trackPendingAgentMessage(payload.sessionNo || "", clientMessageId);
    scheduleRefreshSessions();
  }
  return sent;
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
  if (!message.sessionNo || !message.messageNo) return false;
  const eventType =
    operateType === messageOperateType.RECALL
      ? chatEventType.MESSAGE_RECALL
      : chatEventType.MESSAGE_DELETE;
  return sendWsEvent(
    createChatWsRequest(eventType, "messageOperate", {
      sessionNo: message.sessionNo,
      messageNo: message.messageNo,
      operateType,
      operatorId: auth.user?.id || 0,
      operatorType: 2,
      deleteScope,
      operatedAt: Date.now(),
      isGuest: isGuestSession(activeSession.value),
    }),
  );
}

function buildOptimisticAgentMessage(
  payload: SendAgentMessagePayload,
  clientMessageId: string,
): ChatMessage {
  const now = Date.now();
  return {
    id: 0,
    messageNo: clientMessageId,
    sessionNo: payload.sessionNo || "",
    merchantId: payload.merchantId || merchantId.value,
    senderType: 2,
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
    status: 0,
    extra: payload.extra || "",
    readTime: 0,
    createTime: now,
    updateTime: now,
  };
}

function canSendMessage(content: string) {
  return Boolean(
    content &&
      canSendResourceMessage(),
  );
}

function canSendResourceMessage() {
  return Boolean(
      socket &&
      socket.readyState === WebSocket.OPEN &&
      activeSession.value &&
      merchantId.value &&
      agentId.value &&
      !activeNeedsAccept.value &&
      !activeClosed.value,
  );
}

function closeSession() {
  if (!activeSession.value) {
    return;
  }
  sendWsEvent(
    createChatWsRequest(
      chatEventType.SESSION_CLOSE,
      "session",
      buildCloseSessionPayload(activeSession.value),
    ),
  );
}

function buildCloseSessionPayload(
  session: ChatSession,
): CloseAgentChatSessionPayload {
  return {
    merchantId: merchantId.value,
    userId: session.userId,
    sessionNo: session.sessionNo,
    closeReason: "closed by agent",
    isGuest: isGuestSession(session),
  };
}

function acceptSession() {
  if (!activeSession.value) {
    return;
  }
  if (auth.agent?.status !== agentStatus.online) {
    ElMessage.warning("坐席在线后才能接待会话");
    return;
  }
  sendWsEvent(
    createChatWsRequest(
      chatEventType.AGENT_ACCEPTED,
      "agent",
      buildAcceptSessionPayload(activeSession.value),
    ),
  );
}

function buildAcceptSessionPayload(
  session: ChatSession,
): AcceptChatSessionPayload {
  return {
    merchantId: merchantId.value,
    agentId: agentId.value,
    sessionNo: session.sessionNo,
    isGuest: isGuestSession(session),
    reason: "accepted by agent",
  };
}
</script>

<template>
  <section
    class="workbench"
    :class="{ 'mobile-chat-open': mobileChatOpen }"
  >
    <WorkbenchSessionList
      v-model:status-filter="statusFilter"
      :sessions="filteredSessions"
      :selected-session-no="activeSession?.sessionNo || ''"
      :loading="loadingSessions"
      :agent-status-option="agentStatusOption"
      :agent-status-options="agentStatusOptions"
      :status-changing="changingAgentStatus"
      @select="selectSession"
      @refresh="refreshSessions"
      @status-change="changeAgentStatus"
    />

    <WorkbenchChatPanel
      :session="activeSession"
      :messages="visibleMessages"
      :loading="loadingMessages"
      :active-needs-accept="activeNeedsAccept"
      :active-closed="activeClosed"
      :can-reply="canReply"
      :can-accept="agentCanAccept"
      :accept-disabled-reason="acceptDisabledReason"
      :ws-online="wsOnline"
      :agent-id="agentId"
      :user-id="userId"
      :show-guest-refresh-notice="showGuestRefreshNotice"
      :show-mobile-back="mobileChatOpen"
      :resolve-url="resolveFileUrl"
      @accept="acceptSession"
      @back="backToSessions"
      @close="closeSession"
      @send="send"
      @send-image="sendImage"
      @recall-message="recallMessage"
      @delete-message="deleteMessage"
    />

    <WorkbenchCustomerPanel :session="activeSession" />
  </section>
</template>
