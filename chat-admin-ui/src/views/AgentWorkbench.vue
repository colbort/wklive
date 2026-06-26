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
import { chatAdminWsEventTypes, chatEventType } from "@/api/constant";
import WorkbenchChatPanel from "@/components/workbench/WorkbenchChatPanel.vue";
import WorkbenchCustomerPanel from "@/components/workbench/WorkbenchCustomerPanel.vue";
import WorkbenchSessionList from "@/components/workbench/WorkbenchSessionList.vue";
import { useAuthStore } from "@/stores/auth";
import type {
  ChatAgent,
  ChatMessage,
  ChatQueueInfo,
  ChatSession,
  ChatSessionEvent,
} from "@/types/chat";
import {
  optionGroup,
  withOptionLabels,
  type DisplayOptionItem,
} from "@/utils/options";

interface WsBase {
  code: number;
  msg: string;
}

interface WsResult<T> {
  base?: WsBase;
  data?: T;
}

type WsEventData = ChatMessage | WsResult<ChatMessage>;

interface WsEvent {
  type: string;
  data?: WsEventData;
  base?: WsBase;
  agent?: ChatAgent;
  session?: ChatSession;
  sessionEvent?: ChatSessionEvent;
  session_event?: ChatSessionEvent;
  queue?: ChatQueueInfo;
}

const statusFilter = ref("waiting");
const input = ref("");
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
  console.info("============.  1. ", sessionNo)
  if (sessionNo) {
    await loadMessages(sessionNo);
  }
});

watch(activeNeedsAccept, async (needsAccept) => {
  if (!needsAccept && activeSessionNo.value) {
    console.info("============.  2. ", activeSessionNo.value)
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
    try {
      const event = JSON.parse(message.data) as WsEvent;
      handleWsMessage(event);
    } catch {
      // ignore invalid push payload
    }
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

function handleWsMessage(event: WsEvent) {
  const eventType = event.type || "";
  if (!chatAdminWsEventTypes.has(eventType)) {
    return;
  }
  const base = event.base || eventBase(event.data);
  if (base && base.code !== 200) {
    ElMessage.error(base.msg || "发送失败");
    return;
  }

  switch (eventType) {
    case chatEventType.AGENT_JOIN:
    case chatEventType.AGENT_LEAVE:
      updateAgentFromEvent(event.agent);
      return;

    case chatEventType.AGENT_ASSIGNED: {
      const message = applyWsSessionMessage(event, eventType);
      markSessionAccepted(event, message);
      return;
    }

    case chatEventType.SESSION_CLOSE: {
      const message = applyWsSessionMessage(event, eventType);
      markSessionClosed(event, message);
      return;
    }

    case chatEventType.QUEUE_JOIN:
    case chatEventType.QUEUE_UPDATE:
    case chatEventType.QUEUE_LEAVE:
      applyWsSessionMessage(event, eventType);
      focusWaitingSession(event);
      scheduleRefreshSessions();
      return;

    case chatEventType.USER_JOIN:
      focusWaitingSession(event);
      void loadSessions();
      return;

    case chatEventType.USER_LEAVE:
      handleUserLeave(event);
      return;

    default:
      applyWsSessionMessage(event, eventType);
  }
}

function applyWsSessionMessage(event: WsEvent, eventType: string) {
  upsertSessionFromEvent(event);
  const message = eventMessage(event.data);
  let messagePushed = false;
  if (message?.sessionNo && message.messageNo) {
    normalizeMessageEnums(message);
    if (shouldAppendWsMessage(eventType, message)) {
      messagePushed = pushMessage(message);
    }
  }
  if (message?.sessionNo && shouldAppendWsMessage(eventType, message)) {
    upsertSessionFromMessage(message, messagePushed);
  }
  const session = message?.sessionNo
    ? sessions.value.find((item) => item.sessionNo === message.sessionNo)
    : undefined;
  if (
    message?.senderType === 1 &&
    !session?.agentId &&
    statusFilter.value !== "waiting"
  ) {
    statusFilter.value = "waiting";
  }
  return message;
}

function focusWaitingSession(event: WsEvent) {
  const queueSessionNo =
    event.queue?.sessionNo || event.session?.sessionNo || "";
  if (!mobileChatOpen.value || queueSessionNo !== activeSessionNo.value) {
    statusFilter.value = "waiting";
    mobileChatOpen.value = false;
  }
}

function handleUserLeave(event: WsEvent) {
  const sessionNo = eventSessionNo(event);
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
  const message =
    sessionEvent(event)?.message || eventMessage(event.data)?.content || "用户已离开";
  setSessionStatusMessage(sessionNo, message);
  session.lastMessage = message;
  session.lastMessageTime = Number(sessionEvent(event)?.createdAt || Date.now());
  session.updateTimes = session.lastMessageTime;
  ElMessage.info(message);
}

function eventSessionNo(event: WsEvent) {
  return (
    event.session?.sessionNo ||
    event.queue?.sessionNo ||
    sessionEvent(event)?.sessionNo ||
    sessionEvent(event)?.session?.sessionNo ||
    eventMessage(event.data)?.sessionNo ||
    ""
  );
}

function removeSession(sessionNo: string) {
  sessions.value = sessions.value.filter((item) => item.sessionNo !== sessionNo);
  const { [sessionNo]: _removedMessage, ...nextStatusMessages } =
    sessionStatusMessages.value;
  sessionStatusMessages.value = nextStatusMessages;
  const { [sessionNo]: _removedMessages, ...nextMessages } = messages.value;
  messages.value = nextMessages;
}

function updateAgentFromEvent(agent?: ChatAgent) {
  if (!agent || agent.id !== agentId.value) return;
  auth.agent = {
    ...(auth.agent || agent),
    ...agent,
  };
}

function upsertSessionFromEvent(event: WsEvent) {
  const session = event.session || sessionEvent(event)?.session;
  if (session?.sessionNo) {
    upsertSession(normalizeSession(session));
  }
  const queue = event.queue || sessionEvent(event)?.queue;
  if (queue?.sessionNo) {
    setSessionStatusMessage(queue.sessionNo, queue.message);
    const exists = sessions.value.find(
      (item) => item.sessionNo === queue.sessionNo,
    );
    if (exists) {
      exists.groupId = queue.groupId || exists.groupId;
      exists.updateTimes = queue.updateTimes || exists.updateTimes;
    }
  }
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

function markSessionAccepted(event: WsEvent, message?: ChatMessage) {
  const sessionEventData = sessionEvent(event);
  const sessionNo = sessionEventData?.sessionNo || message?.sessionNo || "";
  setSessionStatusMessage(
    sessionNo,
    sessionEventData?.message || message?.content || "",
  );
  const session = sessions.value.find((item) => item.sessionNo === sessionNo);
  const acceptedAgentId = Number(
    sessionEventData?.agentId ||
      (message ? messageAgentId(message) : 0) ||
      agentId.value,
  );
  if (!session) {
    if (!message) return;
    const next = transientSessionFromMessage(message);
    next.agentId = acceptedAgentId;
    next.status = sessionStatus.serving;
    sessions.value = [next, ...sessions.value];
  } else {
    session.agentId = acceptedAgentId;
    session.status = sessionStatus.serving;
    session.lastMessage =
      message?.content || sessionEventData?.message || session.lastMessage;
    session.lastMessageNo = message?.messageNo || session.lastMessageNo;
    session.lastMessageTime =
      message?.createTime ||
      sessionEventData?.createdAt ||
      session.lastMessageTime;
    session.lastSenderType = message?.senderType || session.lastSenderType;
    session.updateTimes =
      message?.updateTime ||
      sessionEventData?.createdAt ||
      session.updateTimes;
  }
  if (acceptedAgentId === agentId.value) {
    statusFilter.value = "serving";
    activeSessionNo.value = sessionNo;
  } else {
    syncActiveSession();
  }
}

function markSessionClosed(event: WsEvent, message?: ChatMessage) {
  const sessionEventData = sessionEvent(event);
  const sessionNo = sessionEventData?.sessionNo || message?.sessionNo || "";
  if (!sessionNo) return;
  setSessionStatusMessage(
    sessionNo,
    sessionEventData?.message || message?.content || "本次会话已结束",
  );
  const session = sessions.value.find((item) => item.sessionNo === sessionNo);
  if (session) {
    session.status = sessionStatus.closed;
    session.closeTime =
      sessionEventData?.createdAt || message?.createTime || Date.now();
    session.closeReason = sessionEventData?.reason || session.closeReason;
    session.lastMessage =
      message?.content ||
      sessionEventData?.message ||
      session.lastMessage ||
      "本次会话已结束";
    session.lastMessageTime = message?.createTime || session.closeTime;
    session.updateTimes = message?.updateTime || session.closeTime;
  }
  if (activeSessionNo.value === sessionNo) {
    statusFilter.value = "closed";
  }
  syncActiveSession();
}

function upsertSession(session: ChatSession) {
  const exists = sessions.value.find(
    (item) => item.sessionNo === session.sessionNo,
  );
  if (!exists) {
    sessions.value = [normalizeSession(session), ...sessions.value];
    return;
  }
  Object.assign(exists, normalizeSession(session));
}

function transientSessionFromMessage(message: ChatMessage): ChatSession {
  const senderType = senderTypeValue(message.senderType);
  return {
    id: 0,
    sessionNo: message.sessionNo,
    merchantId: Number(message.merchantId || merchantId.value),
    userId: Number(messageUserId(message)),
    userNickname: message.sender?.nickname || "访客",
    userAvatarUrl: message.sender?.avatarUrl || "",
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
    extJson: "",
    lastMessageNo: message.messageNo,
    createTimes: message.createTime,
    updateTimes: message.updateTime || message.createTime,
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
  return [...localLiveSessions, ...validNextSessions];
}

function isGuestSession(session?: ChatSession) {
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

function setSessionStatusMessage(sessionNo: string, message?: string) {
  const text = message?.trim();
  if (!sessionNo || !text) return;
  sessionStatusMessages.value = {
    ...sessionStatusMessages.value,
    [sessionNo]: text,
  };
}

function matchStatusFilter(session: ChatSession) {
  const status = sessionStatusValue(session.status);
  const assignedAgentId = Number(session.agentId || 0);
  if (statusFilter.value === "waiting") {
    return (
      status === sessionStatus.waiting ||
      (status === sessionStatus.pendingAgent && !assignedAgentId)
    );
  }
  if (statusFilter.value === "serving") {
    return (
      assignedAgentId === agentId.value &&
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

function shouldAppendWsMessage(eventType: string, message?: ChatMessage) {
  if (!message || isQueueSystemMessage(message)) return false;
  return eventType === chatEventType.MESSAGE;
}

function normalizeSession(session: ChatSession) {
  session.id = Number(session.id || 0);
  session.merchantId = Number(session.merchantId || 0);
  session.status = sessionStatusValue(session.status);
  session.source = Number(session.source || 0);
  session.priority = Number(session.priority || 0);
  session.agentId = Number(session.agentId || 0);
  session.userId = Number(session.userId || 0);
  session.userNickname = session.userNickname || session.title || "";
  session.userAvatarUrl = session.userAvatarUrl || "";
  session.groupId = Number(session.groupId || 0);
  session.lastSenderType = senderTypeValue(session.lastSenderType);
  session.lastMessageTime = Number(session.lastMessageTime || 0);
  session.userUnreadCount = Number(session.userUnreadCount || 0);
  session.agentUnreadCount = Number(session.agentUnreadCount || 0);
  session.closeTime = Number(session.closeTime || 0);
  session.createTimes = Number(session.createTimes || 0);
  session.updateTimes = Number(session.updateTimes || 0);
  return session;
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

function eventMessage(data?: WsEventData) {
  if (!data) return undefined;
  if (isWsResult(data)) return data.data;
  return data;
}

function sessionEvent(event: WsEvent) {
  return event.sessionEvent || event.session_event;
}

function eventBase(data?: WsEventData) {
  return data && isWsResult(data) ? data.base : undefined;
}

function isWsResult(data: WsEventData): data is WsResult<ChatMessage> {
  return "data" in data || "base" in data;
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

function send() {
  const content = input.value.trim();
  if (
    !content ||
    !activeSession.value ||
    !merchantId.value ||
    !agentId.value ||
    activeNeedsAccept.value ||
    activeClosed.value
  ) {
    return;
  }
  socket?.send(
    JSON.stringify({
      type: chatEventType.MESSAGE,
      data: {
        merchantId: merchantId.value,
        agentId: agentId.value,
        userId: activeSession.value.userId,
        sessionNo: activeSession.value.sessionNo,
        messageType: 1,
        content,
      },
    }),
  );
  input.value = "";
}

function closeSession() {
  if (!activeSession.value || !socket || socket.readyState !== WebSocket.OPEN) {
    return;
  }
  socket.send(
    JSON.stringify({
      type: chatEventType.SESSION_CLOSE,
      data: {
        merchantId: merchantId.value,
        userId: activeSession.value.userId,
        sessionNo: activeSession.value.sessionNo,
        closeReason: "closed by agent",
      },
    }),
  );
}

function acceptSession() {
  if (!activeSession.value || !socket || socket.readyState !== WebSocket.OPEN) {
    return;
  }
  if (auth.agent?.status !== agentStatus.online) {
    ElMessage.warning("坐席在线后才能接待会话");
    return;
  }
  socket.send(
    JSON.stringify({
      type: chatEventType.AGENT_ASSIGNED,
      data: {
        merchantId: merchantId.value,
        agentId: agentId.value,
        userId: activeSession.value.userId,
        sessionNo: activeSession.value.sessionNo,
      },
    }),
  );
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
      v-model:input-value="input"
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
