<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { ElMessage } from "element-plus";
import {
  chatAdminWsUrl,
  options as loadOptions,
  pageMessages,
  pageSessions,
  sendAgentMessage,
  updateAgentStatus,
} from "@/api/chat";
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
import { withOptionLabels, type DisplayOptionItem } from "@/utils/options";

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
const wsState = ref<"idle" | "open" | "closed">("idle");
const changingAgentStatus = ref(false);
const defaultAgentStatusOptions: DisplayOptionItem[] = [
  {
    key: "chat.agent.status.offline",
    label: "离线",
    value: 1,
    tagType: "info",
  },
  {
    key: "chat.agent.status.online",
    label: "在线",
    value: 2,
    tagType: "success",
  },
  {
    key: "chat.agent.status.busy",
    label: "忙碌",
    value: 3,
    tagType: "warning",
  },
  {
    key: "chat.agent.status.resting",
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
const activeIsGuest = computed(() =>
  isGuestSession(activeSession.value?.sessionNo),
);
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
    if (resp.data.agentStatuses?.length) {
      agentStatusOptions.value = withOptionLabels(resp.data.agentStatuses);
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
    sessions.value = mergeTransientSessions(resp.data.map(normalizeSession));
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
    messages.value[sessionNo] = resp.data
      .map(normalizeMessage)
      .filter((message) => !isQueueSystemMessage(message));
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
  socket.onmessage = (event) => {
    handleWsMessage(event.data);
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
  try {
    const event = JSON.parse(payload) as WsEvent;
    if (
      event.type !== "chat.message" &&
      event.type !== "chat.session.accepted" &&
      event.type !== "chat.session.closed" &&
      event.type !== "chat.queue.updated" &&
      event.type !== "chat.agent.status.updated" &&
      event.type !== "accept_chat_session.result" &&
      event.type !== "send_agent_message.result" &&
      event.type !== "close_chat_session.result"
    ) {
      return;
    }
    const base = event.base || eventBase(event.data);
    if (base && base.code !== 200) {
      ElMessage.error(base.msg || "发送失败");
      return;
    }
    if (event.type === "chat.agent.status.updated") {
      updateAgentFromEvent(event.agent);
      return;
    }
    upsertSessionFromEvent(event);
    const message = eventMessage(event.data);
    let messagePushed = false;
    if (message?.sessionNo && message.messageNo) {
      normalizeMessageEnums(message);
      if (
        !isQueueSystemMessage(message) &&
        event.type !== "accept_chat_session.result"
      ) {
        messagePushed = pushMessage(message);
      }
    }
    if (
      event.type === "chat.session.accepted" ||
      event.type === "accept_chat_session.result"
    ) {
      markSessionAccepted(event, message);
    } else if (
      event.type === "chat.session.closed" ||
      event.type === "close_chat_session.result"
    ) {
      markSessionClosed(event, message);
    } else {
      if (message?.sessionNo && !isQueueSystemMessage(message)) {
        upsertSessionFromMessage(message, messagePushed);
      }
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
    if (event.type === "chat.queue.updated") {
      statusFilter.value = "waiting";
      mobileChatOpen.value = false;
      scheduleRefreshSessions();
    }
  } catch {
    // ignore invalid push payload
  }
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
  messages.value[message.sessionNo] = [...list, message];
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
    exists.lastMessageTime = message.createTimes;
    exists.lastSenderType = senderTypeValue(message.senderType);
    exists.updateTimes = message.updateTimes;
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
      exists.agentId = Number(message.agentId || agentId.value);
    }
    return;
  }

  sessions.value = [transientSessionFromMessage(message), ...sessions.value];
  syncActiveSession();
}

function markSessionAccepted(event: WsEvent, message?: ChatMessage) {
  const sessionEventData = sessionEvent(event);
  const sessionNo = sessionEventData?.sessionNo || message?.sessionNo || "";
  const session = sessions.value.find((item) => item.sessionNo === sessionNo);
  const acceptedAgentId = Number(
    sessionEventData?.agentId ||
      message?.sender?.id ||
      message?.agentId ||
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
      message?.createTimes ||
      sessionEventData?.createdAt ||
      session.lastMessageTime;
    session.lastSenderType = message?.senderType || session.lastSenderType;
    session.updateTimes =
      message?.updateTimes ||
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
  const session = sessions.value.find((item) => item.sessionNo === sessionNo);
  if (session) {
    session.status = sessionStatus.closed;
    session.closeTime =
      sessionEventData?.createdAt || message?.createTimes || Date.now();
    session.closeReason = sessionEventData?.reason || session.closeReason;
    session.lastMessage =
      message?.content ||
      sessionEventData?.message ||
      session.lastMessage ||
      "本次会话已结束";
    session.lastMessageTime = message?.createTimes || session.closeTime;
    session.updateTimes = message?.updateTimes || session.closeTime;
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
    userId: Number(message.userId || 0),
    userNickname: message.sender?.nickname || "访客",
    userAvatarUrl: message.sender?.avatarUrl || "",
    source: 2,
    status:
      senderType === 2 ? sessionStatus.pendingUser : sessionStatus.pendingAgent,
    priority: 1,
    agentId: Number(message.agentId || 0),
    groupId: 0,
    title: message.sender?.nickname || "访客",
    category: "",
    lastMessage: message.content,
    lastSenderType: message.senderType,
    lastMessageTime: message.createTimes,
    userUnreadCount: senderType === 2 ? 1 : 0,
    agentUnreadCount: senderType === 1 ? 1 : 0,
    closeTime: 0,
    closeReason: "",
    extJson: "",
    lastMessageNo: message.messageNo,
    createTimes: message.createTimes,
    updateTimes: message.updateTimes || message.createTimes,
  };
}

function mergeTransientSessions(nextSessions: ChatSession[]) {
  const validNextSessions = nextSessions.filter((item) =>
    Boolean(item.sessionNo),
  );
  const nextNos = new Set(validNextSessions.map((item) => item.sessionNo));
  const transient = sessions.value.filter(
    (item) => isGuestSession(item.sessionNo) && !nextNos.has(item.sessionNo),
  );
  return [...transient, ...validNextSessions];
}

function isGuestSession(sessionNo?: string) {
  return typeof sessionNo === "string" && sessionNo.startsWith("GS");
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
  message.senderType = senderTypeValue(message.senderType);
}

function normalizeMessage(message: ChatMessage) {
  normalizeMessageEnums(message);
  return message;
}

function isQueueSystemMessage(message?: ChatMessage) {
  return (
    senderTypeValue(message?.senderType) === 3 &&
    typeof message?.content === "string" &&
    message.content.includes("正在排队")
  );
}

function normalizeSession(session: ChatSession) {
  session.status = sessionStatusValue(session.status);
  session.agentId = Number(session.agentId || 0);
  session.userId = Number(session.userId || 0);
  session.userNickname = session.userNickname || session.title || "";
  session.userAvatarUrl = session.userAvatarUrl || "";
  session.groupId = Number(session.groupId || 0);
  session.lastSenderType = senderTypeValue(session.lastSenderType);
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

async function send() {
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
  const sessionNo = activeSession.value.sessionNo;
  if (isGuestSession(sessionNo)) {
    socket?.send(
      JSON.stringify({
        type: "send_agent_message",
        data: {
          merchantId: merchantId.value,
          agentId: agentId.value,
          userId: activeSession.value.userId,
          sessionNo,
          messageType: 1,
          content,
        },
      }),
    );
    input.value = "";
    return;
  }
  try {
    const resp = await sendAgentMessage(sessionNo, {
      merchantId: merchantId.value,
      agentId: agentId.value,
      messageType: 1,
      content,
    });
    normalizeMessageEnums(resp.data);
    pushMessage(resp.data);
    upsertSessionFromMessage(resp.data, false);
    input.value = "";
  } catch (err) {
    ElMessage.error(err instanceof Error ? err.message : "发送失败");
  }
}

function closeSession() {
  if (!activeSession.value || !socket || socket.readyState !== WebSocket.OPEN) {
    return;
  }
  socket.send(
    JSON.stringify({
      type: "close_chat_session",
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
      type: "accept_chat_session",
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
