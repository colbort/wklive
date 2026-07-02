<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { ElMessage } from "element-plus";
import {
  chatAdminWsUrl,
  options as loadOptions,
  pageMessages,
  pageSessions,
  updateAgentStatus,
} from "@/api/chat";
import { chatEventType, type ChatEventType } from "@/api/constant";
import WorkbenchChatPanel from "@/components/workbench/WorkbenchChatPanel.vue";
import WorkbenchCustomerPanel from "@/components/workbench/WorkbenchCustomerPanel.vue";
import WorkbenchSessionList from "@/components/workbench/WorkbenchSessionList.vue";
import { useAuthStore } from "@/stores/auth";
import type {
  AcceptChatSessionPayload,
  ChatAdminUiWsReq,
  ChatAgentPayload,
  ChatWsEvent,
  ChatAgent,
  ChatMessage,
  ChatQueuePayload,
  ChatSession,
  ChatSessionExtJson,
  ChatUserStatePayload,
  CloseAgentChatSessionPayload,
  SendAgentMessagePayload,
  WsConnectedPayload,
} from "@/types/chat";
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
  activeNeedsAccept.value
    ? []
    : activeMessages.value.filter((message) => !isQueueSystemMessage(message)),
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
  [chatEventType.USER_JOIN]: (event) => handleUserJoinWsEvent(event),
  [chatEventType.USER_LEAVE]: (event) =>
    event.userState && handleUserLeaveWsEvent(event.userState),
  [chatEventType.QUEUE_UPDATE]: (event) =>
    event.queue && handleQueueUpdateWsEvent(event.queue),
  [chatEventType.AGENT_ACCEPTED]: (event) =>
    event.agent && handleAgentAcceptedWsEvent(event.agent),
  [chatEventType.AGENT_LEAVE]: (event) =>
    event.agent && handleAgentStateWsEvent(event.agent),
  [chatEventType.SESSION_CLOSE]: (event) =>
    event.session && handleSessionCloseWsEvent(event.session),
};

onMounted(async () => {
  destroyed = false;
  restoreWorkbenchState();
  void loadAdminOptions();
  await loadSessions();
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

async function loadSessions() {
  if (!merchantId.value) return;
  loadingSessions.value = true;
  try {
    const resp = await pageSessions({
      merchantId: merchantId.value,
      limit: 50,
    });
    sessions.value = mergeLiveSessions(resp.data.map(normalizeSession));
    syncActiveSession();
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
      limit: 50,
    });
    messages.value[sessionNo] = sortMessages(
      resp.data
        .map(normalizeMessage)
        .filter((message) => !isQueueSystemMessage(message)),
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
}

function handleAgentAcceptedWsEvent(payload: ChatAgentPayload) {
  markSessionAccepted(payload);
}

function handleSessionCloseWsEvent(payload: ChatSession) {
  markSessionClosed(payload);
  scheduleRefreshSessions();
}

function handleQueueUpdateWsEvent(payload: ChatQueuePayload) {
  updateQueueFromPayload(payload);
  focusWaitingSession(payload.sessionNo);
  scheduleRefreshSessions();
}

function handleUserJoinWsEvent(event: ChatWsEvent) {
  const session = event.session ? normalizeSession(event.session) : undefined;
  const sessionNo = session?.sessionNo || event.userState?.sessionNo || "";
  if (!sessionNo) {
    scheduleRefreshSessions();
    return;
  }
  if (session) {
    upsertSession(session);
  } else if (!sessions.value.some((item) => item.sessionNo === sessionNo)) {
    upsertSession(sessionFromUserState(event.userState));
  }
  focusWaitingSession(sessionNo);
  scheduleRefreshSessions();
}

function handleUserLeaveWsEvent(payload: ChatUserStatePayload) {
  handleUserLeave(payload);
}

function handleMessageWsEvent(payload: ChatMessage) {
  applyWsSessionMessage(payload);
}

function handleWsConnectedWsEvent(payload: WsConnectedPayload) {
  connected.value = payload;
}

function applyWsSessionMessage(message: ChatMessage) {
  let messagePushed = false;
  if (message.sessionNo && message.messageNo) {
    normalizeMessageEnums(message);
    if (shouldAppendWsMessage(message)) {
      messagePushed = pushMessage(message);
    }
  }
  if (message.sessionNo && shouldAppendWsMessage(message)) {
    upsertSessionFromMessage(message, messagePushed);
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
  return message;
}

function focusWaitingSession(sessionNo: string) {
  if (!mobileChatOpen.value || sessionNo !== activeSessionNo.value) {
    statusFilter.value = "waiting";
    mobileChatOpen.value = false;
  }
}

function handleUserLeave(payload: ChatUserStatePayload) {
  const sessionNo = payload.sessionNo;
  if (!sessionNo) {
    scheduleRefreshSessions();
    return;
  }
  const index = sessions.value.findIndex(
    (item) => item.sessionNo === sessionNo,
  );
  if (index < 0) {
    scheduleRefreshSessions();
    return;
  }
  const session = sessions.value[index];
  const status = sessionStatusValue(session.status);
  if (isGuestSession(session)) {
    removeSession(sessionNo);
    if (activeSessionNo.value === sessionNo) {
      syncActiveSession();
      mobileChatOpen.value = false;
    }
    return;
  }
  if (status === sessionStatus.waiting || needsAccept(session)) {
    removeSession(sessionNo);
    if (activeSessionNo.value === sessionNo) {
      syncActiveSession();
      mobileChatOpen.value = false;
    }
    if (statusFilter.value === "waiting") {
      scheduleRefreshSessions();
    }
    return;
  }
  if (status === sessionStatus.closed) {
    if (statusFilter.value === "closed") {
      scheduleRefreshSessions();
    }
    return;
  }
  const message = payload.userName || "用户已离开";
  setSessionStatusMessage(sessionNo, message);
  session.lastMessage = message;
  session.lastMessageTime = Number(Date.now());
  session.updateTimes = session.lastMessageTime;
  ElMessage.info(message);
}

function removeSession(sessionNo: string) {
  sessions.value = sessions.value.filter((item) => item.sessionNo !== sessionNo);
  const { [sessionNo]: _removedMessage, ...nextStatusMessages } =
    sessionStatusMessages.value;
  sessionStatusMessages.value = nextStatusMessages;
  const { [sessionNo]: _removedMessages, ...nextMessages } = messages.value;
  messages.value = nextMessages;
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

function updateQueueFromPayload(queue: ChatQueuePayload) {
  if (!queue.sessionNo) return;
  setSessionStatusMessage(queue.sessionNo, queueStatusMessage(queue));
  const exists = sessions.value.find(
    (item) => item.sessionNo === queue.sessionNo,
  );
  if (!exists) return;
  const nextStatus = sessionStatusValue(queue.sessionStatus);
  if (nextStatus) {
    exists.status = nextStatus;
  }
  exists.updateTimes = Number(queue.actionTime || exists.updateTimes);
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
  const list = messages.value[message.sessionNo] || [];
  if (list.some((item) => item.messageNo === message.messageNo)) return false;
  messages.value[message.sessionNo] = sortMessages([...list, message]);
  return true;
}

function upsertSessionFromMessage(
  message: ChatMessage,
  shouldCountUnread = true,
) {
  const exists = sessions.value.find(
    (item) => item.sessionNo === message.sessionNo,
  );
  if (exists) {
    exists.lastMessage = message.content;
    exists.lastMessageNo = message.messageNo;
    exists.lastMessageTime = message.createTime;
    exists.lastSenderType = senderTypeValue(message.senderType);
    exists.updateTimes = message.updateTime;
    const senderType = senderTypeValue(message.senderType);
    if (
      senderType === 1 &&
      sessionStatusValue(exists.status) !== sessionStatus.closed
    ) {
      exists.status = sessionStatus.pendingAgent;
      if (shouldCountUnread) {
        exists.agentUnreadCount += 1;
      }
    }
    if (
      senderType === 2 &&
      sessionStatusValue(exists.status) !== sessionStatus.closed
    ) {
      exists.status = sessionStatus.pendingUser;
      if (shouldCountUnread) {
        exists.userUnreadCount += 1;
      }
      exists.agentId = Number(messageAgentId(message) || agentId.value);
    }
    return;
  }

  sessions.value = [transientSessionFromMessage(message), ...sessions.value];
  syncActiveSession();
}

function markSessionAccepted(payload: ChatAgentPayload) {
  const sessionNo = payload.sessionNo;
  if (!sessionNo) return;
  setSessionStatusMessage(sessionNo, payload.remark || "");
  const session = sessions.value.find((item) => item.sessionNo === sessionNo);
  const acceptedAgentId = Number(payload.agentId || agentId.value);
  if (!session) {
    scheduleRefreshSessions();
  } else {
    session.agentId = acceptedAgentId;
    session.status = sessionStatus.serving;
    session.updateTimes = Number(payload.actionTime || session.updateTimes);
  }
  if (acceptedAgentId === agentId.value) {
    statusFilter.value = "serving";
    activeSessionNo.value = sessionNo;
  } else {
    syncActiveSession();
  }
}

function markSessionClosed(payload: ChatSession) {
  const sessionNo = payload.sessionNo;
  if (!sessionNo) return;
  const closedSession = normalizeSession(payload);
  upsertSession(closedSession);
  const session = sessions.value.find((item) => item.sessionNo === sessionNo);
  if (isGuestSession(session)) {
    removeSession(sessionNo);
    if (activeSessionNo.value === sessionNo) {
      syncActiveSession();
      mobileChatOpen.value = false;
    }
    return;
  }
  setSessionStatusMessage(
    sessionNo,
    closedSession.closeReason || "本次会话已结束",
  );
  if (session) {
    session.status = sessionStatus.closed;
    session.closeTime = closedSession.closeTime || Date.now();
    session.closeReason = closedSession.closeReason || session.closeReason || "";
    session.lastMessage =
      closedSession.closeReason ||
      closedSession.lastMessage ||
      session.lastMessage ||
      "本次会话已结束";
    session.lastMessageTime =
      closedSession.lastMessageTime || session.closeTime;
    session.updateTimes = closedSession.updateTimes || session.closeTime;
  }
  if (activeSessionNo.value === sessionNo) {
    statusFilter.value = "closed";
  }
  syncActiveSession();
}

function upsertSession(session: ChatSession) {
  const normalized = normalizeSession(session);
  const exists = sessions.value.find(
    (item) => item.sessionNo === normalized.sessionNo,
  );
  if (!exists) {
    const guestKey = guestSessionKey(normalized);
    const rest = guestKey
      ? sessions.value.filter((item) => guestSessionKey(item) !== guestKey)
      : sessions.value;
    sessions.value = [normalized, ...rest];
    return;
  }
  Object.assign(exists, normalized);
  sessions.value = collapseGuestSessions(sessions.value);
}

function transientSessionFromMessage(message: ChatMessage): ChatSession {
  const senderType = senderTypeValue(message.senderType);
  return {
    id: 0,
    sessionNo: message.sessionNo,
    merchantId: Number(message.merchantId || merchantId.value),
    userId: Number(messageUserId(message)),
    source: 2,
    status:
      senderType === 2 ? sessionStatus.pendingUser : sessionStatus.pendingAgent,
    priority: 1,
    agentId: Number(messageAgentId(message)),
    groupId: 0,
    title: message.sender?.nickname || "访客",
    category: "",
    lastMessage: message.content,
    lastSenderType: message.senderType,
    lastMessageTime: message.createTime,
    userUnreadCount: senderType === 2 ? 1 : 0,
    agentUnreadCount: senderType === 1 ? 1 : 0,
    closeTime: 0,
    closeReason: "",
    extJson: {
      nickname: message.sender?.nickname || "访客",
      avatarUrl: message.sender?.avatarUrl || "",
    },
    lastMessageNo: message.messageNo,
    createTimes: message.createTime,
    updateTimes: message.updateTime || message.createTime,
  };
}

function sessionFromUserState(payload?: ChatUserStatePayload): ChatSession {
  const now = Date.now();
  const nickname = payload?.userName || "访客";
  return {
    id: 0,
    sessionNo: payload?.sessionNo || "",
    merchantId: merchantId.value,
    userId: Number(payload?.userId || 0),
    source: Number(payload?.source || 0),
    status: sessionStatus.waiting,
    priority: 1,
    agentId: 0,
    groupId: 0,
    title: nickname,
    category: "",
    lastMessage: payload?.online ? "用户已进入客服页面" : "用户离线",
    lastSenderType: 1,
    lastMessageTime: now,
    userUnreadCount: 0,
    agentUnreadCount: 0,
    closeTime: 0,
    closeReason: "",
    extJson: {
      nickname,
      avatarUrl: payload?.avatar || "",
    },
    avatarUrl: payload?.avatar || "",
    lastMessageNo: "",
    createTimes: now,
    updateTimes: now,
  };
}

function messageAgentId(message?: ChatMessage) {
  if (!message) return 0;
  if (senderTypeValue(message.sender?.type) === 2) return message.sender?.id || 0;
  if (senderTypeValue(message.receiver?.type) === 2)
    return message.receiver?.id || 0;
  return 0;
}

function messageUserId(message?: ChatMessage) {
  if (!message) return 0;
  if (senderTypeValue(message.sender?.type) === 1) return message.sender?.id || 0;
  if (senderTypeValue(message.receiver?.type) === 1)
    return message.receiver?.id || 0;
  return 0;
}

function mergeLiveSessions(nextSessions: ChatSession[]) {
  const validNextSessions = nextSessions.filter((item) =>
    Boolean(item.sessionNo),
  );
  const nextNos = new Set(validNextSessions.map((item) => item.sessionNo));
  const localLiveSessions = sessions.value.filter(
    (item) =>
      Boolean(item.sessionNo) &&
      !nextNos.has(item.sessionNo) &&
      sessionStatusValue(item.status) !== sessionStatus.closed,
  );
  return collapseGuestSessions([...localLiveSessions, ...validNextSessions]);
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
  message.senderType = senderTypeValue(message.sender?.type);
}

function normalizeMessage(message: ChatMessage) {
  normalizeMessageEnums(message);
  return message;
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

function sendWsEvent(request: ChatAdminUiWsReq) {
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
  sendWsEvent({
    eventType: chatEventType.MESSAGE,
    data: buildAgentMessagePayload(content),
  });
}

function canSendMessage(content: string) {
  return Boolean(
    content &&
      activeSession.value &&
      merchantId.value &&
      agentId.value &&
      !activeNeedsAccept.value &&
      !activeClosed.value,
  );
}

function buildAgentMessagePayload(content: string): SendAgentMessagePayload {
  return {
    merchantId: merchantId.value,
    sessionNo: activeSession.value?.sessionNo || "",
    messageType: 1,
    content,
    sender: {
      type: 2,
      id: auth.user?.id || 0,
      nickname: auth.user?.nickname || "",
      avatarUrl: auth.user?.avatarUrl || "",
    },
  };
}

function closeSession() {
  if (!activeSession.value) {
    return;
  }
  sendWsEvent({
    eventType: chatEventType.SESSION_CLOSE,
    data: buildCloseSessionPayload(activeSession.value),
  });
}

function buildCloseSessionPayload(
  session: ChatSession,
): CloseAgentChatSessionPayload {
  return {
    merchantId: merchantId.value,
    userId: session.userId,
    sessionNo: session.sessionNo,
    closeReason: "closed by agent",
    isGuest: session.isGuest || false,
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
  sendWsEvent({
    eventType: chatEventType.AGENT_ACCEPTED,
    data: buildAcceptSessionPayload(activeSession.value),
  });
}

function buildAcceptSessionPayload(
  session: ChatSession,
): AcceptChatSessionPayload {
  return {
    merchantId: merchantId.value,
    agentId: agentId.value,
    sessionNo: session.sessionNo,
    isGuest: session.isGuest,
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
      @accept="acceptSession"
      @back="backToSessions"
      @close="closeSession"
      @send="send"
    />

    <WorkbenchCustomerPanel :session="activeSession" />
  </section>
</template>
