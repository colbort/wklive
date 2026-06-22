<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { ElMessage } from "element-plus";
import {
  chatAdminWsUrl,
  pageMessages,
  pageSessions,
  sendAgentMessage,
} from "@/api/chat";
import { useAuthStore } from "@/stores/auth";
import type { ChatMessage, ChatSession } from "@/types/chat";

interface WsBase {
  code: number;
  msg: string;
}

interface WsResult<T> {
  base?: WsBase;
  data?: T;
}

type WsEventData = ChatMessage | WsResult<ChatMessage>;

const statusFilter = ref("waiting");
const input = ref("");
const activeSessionNo = ref("");
const loadingSessions = ref(false);
const loadingMessages = ref(false);
const sessions = ref<ChatSession[]>([]);
const messages = ref<Record<string, ChatMessage[]>>({});
const wsState = ref<"idle" | "open" | "closed">("idle");
const auth = useAuthStore();
let socket: WebSocket | null = null;
let refreshTimer: number | null = null;

const sessionStatus = {
  waiting: 1,
  serving: 2,
  pendingUser: 3,
  pendingAgent: 4,
  closed: 5,
} as const;

const filteredSessions = computed(() =>
  sessions.value.filter((session) => matchStatusFilter(session)),
);

const activeSession = computed(
  () =>
    filteredSessions.value.find(
      (item) => item.sessionNo === activeSessionNo.value,
    ) || filteredSessions.value[0],
);
const activeMessages = computed(
  () =>
    activeSession.value ? messages.value[activeSession.value.sessionNo] || [] : [],
);
const merchantId = computed(
  () => auth.user?.merchantId || auth.agent?.merchantId || 0,
);
const agentId = computed(() => auth.agent?.id || 0);
const wsOnline = computed(() => wsState.value === "open");
const activeNeedsAccept = computed(
  () =>
    Boolean(activeSession.value) &&
    activeSession.value?.status === sessionStatus.pendingAgent &&
    !activeSession.value?.agentId,
);

onMounted(async () => {
  await loadSessions();
  connectWs();
});

onBeforeUnmount(() => {
  if (refreshTimer) {
    window.clearTimeout(refreshTimer);
  }
  socket?.close();
});

watch(statusFilter, async () => {
  syncActiveSession();
});

watch(activeSessionNo, async (sessionNo) => {
  if (sessionNo) {
    await loadMessages(sessionNo);
  }
});

async function loadSessions() {
  if (!merchantId.value) return;
  loadingSessions.value = true;
  try {
    const resp = await pageSessions({
      merchantId: merchantId.value,
      limit: 50,
    });
    sessions.value = mergeTransientSessions(resp.data);
    syncActiveSession();
  } catch (err) {
    ElMessage.error(err instanceof Error ? err.message : "加载会话失败");
  } finally {
    loadingSessions.value = false;
  }
}

async function loadMessages(sessionNo: string) {
  if (!merchantId.value) return;
  if (isGuestSession(sessionNo)) {
    return;
  }
  loadingMessages.value = true;
  try {
    const resp = await pageMessages(sessionNo, {
      merchantId: merchantId.value,
      limit: 50,
    });
    messages.value[sessionNo] = resp.data;
  } catch (err) {
    ElMessage.error(err instanceof Error ? err.message : "加载消息失败");
  } finally {
    loadingMessages.value = false;
  }
}

function connectWs() {
  if (!auth.token || !merchantId.value) return;
  socket?.close();
  socket = new WebSocket(
    chatAdminWsUrl({
      token: auth.token,
      merchantId: merchantId.value,
      agentId: agentId.value,
    }),
  );
  socket.onopen = () => {
    wsState.value = "open";
  };
  socket.onclose = () => {
    wsState.value = "closed";
  };
  socket.onerror = () => {
    wsState.value = "closed";
  };
  socket.onmessage = (event) => {
    handleWsMessage(event.data);
  };
}

function handleWsMessage(payload: string) {
  try {
    const event = JSON.parse(payload) as {
      type: string;
      data?: WsEventData;
      base?: WsBase;
    };
    if (
      event.type !== "chat.message" &&
      event.type !== "chat.session.accepted" &&
      event.type !== "accept_chat_session.result" &&
      event.type !== "send_agent_message.result"
    ) {
      return;
    }
    const base = event.base || eventBase(event.data);
    if (base && base.code !== 200) {
      ElMessage.error(base.msg || "发送失败");
      return;
    }
    const message = eventMessage(event.data);
    if (!message?.sessionNo || !message.messageNo) {
      return;
    }
    normalizeMessageEnums(message);
    if (!pushMessage(message)) {
      return;
    }
    if (
      event.type === "chat.session.accepted" ||
      event.type === "accept_chat_session.result"
    ) {
      markSessionAccepted(message);
    } else {
      upsertSessionFromMessage(message);
    }
    const session = sessions.value.find(
      (item) => item.sessionNo === message.sessionNo,
    );
    if (
      message.senderType === 1 &&
      !session?.agentId &&
      statusFilter.value !== "waiting"
    ) {
      statusFilter.value = "waiting";
    }
    scheduleRefreshSessions();
  } catch {
    // ignore invalid push payload
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
  if (!message.messageNo) return false;
  const list = messages.value[message.sessionNo] || [];
  if (list.some((item) => item.messageNo === message.messageNo)) return false;
  messages.value[message.sessionNo] = [...list, message];
  return true;
}

function upsertSessionFromMessage(message: ChatMessage) {
  const exists = sessions.value.find(
    (item) => item.sessionNo === message.sessionNo,
  );
  if (exists) {
    exists.lastMessage = message.content;
    exists.lastMessageNo = message.messageNo;
    exists.lastMessageTime = message.createTimes;
    exists.lastSenderType = message.senderType;
    exists.updateTimes = message.updateTimes;
    if (message.senderType === 1 && exists.status !== sessionStatus.closed) {
      exists.status = sessionStatus.pendingAgent;
      exists.agentUnreadCount += 1;
    }
    if (message.senderType === 2 && exists.status !== sessionStatus.closed) {
      exists.status = sessionStatus.pendingUser;
      exists.userUnreadCount += 1;
      exists.agentId = message.agentId || agentId.value;
    }
    return;
  }

  sessions.value = [
    transientSessionFromMessage(message),
    ...sessions.value,
  ];
  syncActiveSession();
}

function markSessionAccepted(message: ChatMessage) {
  const session = sessions.value.find(
    (item) => item.sessionNo === message.sessionNo,
  );
  const acceptedAgentId = Number(message.sender?.id || message.agentId || agentId.value);
  if (!session) {
    const next = transientSessionFromMessage(message);
    next.agentId = acceptedAgentId;
    next.status = sessionStatus.serving;
    sessions.value = [next, ...sessions.value];
  } else {
    session.agentId = acceptedAgentId;
    session.status = sessionStatus.serving;
    session.lastMessage = message.content || session.lastMessage;
    session.lastMessageNo = message.messageNo || session.lastMessageNo;
    session.lastMessageTime = message.createTimes || session.lastMessageTime;
    session.lastSenderType = message.senderType || session.lastSenderType;
    session.updateTimes = message.updateTimes || session.updateTimes;
  }
  if (acceptedAgentId === agentId.value) {
    statusFilter.value = "serving";
    activeSessionNo.value = message.sessionNo;
  } else {
    syncActiveSession();
  }
}

function transientSessionFromMessage(message: ChatMessage): ChatSession {
  const senderType = senderTypeValue(message.senderType);
  return {
    id: 0,
    sessionNo: message.sessionNo,
    merchantId: Number(message.merchantId || merchantId.value),
    userId: Number(message.userId || 0),
    source: 2,
    status:
      senderType === 2
        ? sessionStatus.pendingUser
        : sessionStatus.pendingAgent,
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
  const nextNos = new Set(nextSessions.map((item) => item.sessionNo));
  const transient = sessions.value.filter(
    (item) => isGuestSession(item.sessionNo) && !nextNos.has(item.sessionNo),
  );
  return [...transient, ...nextSessions];
}

function isGuestSession(sessionNo: string) {
  return sessionNo.startsWith("GS");
}

function matchStatusFilter(session: ChatSession) {
  if (statusFilter.value === "waiting") {
    return (
      session.status === sessionStatus.waiting ||
      (session.status === sessionStatus.pendingAgent && !session.agentId)
    );
  }
  if (statusFilter.value === "serving") {
    return (
      session.agentId === agentId.value &&
      ([
        sessionStatus.serving,
        sessionStatus.pendingUser,
        sessionStatus.pendingAgent,
      ] as number[]).includes(session.status)
    );
  }
  return session.status === sessionStatus.closed;
}

function normalizeMessageEnums(message: ChatMessage) {
  message.senderType = senderTypeValue(message.senderType);
}

function senderTypeValue(value: unknown) {
  if (typeof value === "number") return value;
  if (value === "CHAT_SENDER_TYPE_USER") return 1;
  if (value === "CHAT_SENDER_TYPE_AGENT") return 2;
  if (value === "CHAT_SENDER_TYPE_SYSTEM") return 3;
  return 0;
}

function eventMessage(
  data?: WsEventData,
) {
  if (!data) return undefined;
  if (isWsResult(data)) return data.data;
  return data;
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

async function send() {
  const content = input.value.trim();
  if (
    !content ||
    !activeSession.value ||
    !merchantId.value ||
    !agentId.value ||
    activeNeedsAccept.value
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
    pushMessage(resp.data);
    input.value = "";
    await loadSessions();
  } catch (err) {
    ElMessage.error(err instanceof Error ? err.message : "发送失败");
  }
}

function acceptSession() {
  if (!activeSession.value || !socket || socket.readyState !== WebSocket.OPEN) {
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
  <section class="workbench">
    <aside class="session-panel">
      <div class="panel-header">
        <h2>会话</h2>
        <el-tag :type="wsOnline ? 'success' : 'info'">
          {{ wsOnline ? "在线" : "离线" }}
        </el-tag>
      </div>
      <el-segmented
        v-model="statusFilter"
        :options="[
          { label: '待接待', value: 'waiting' },
          { label: '进行中', value: 'serving' },
          { label: '已结束', value: 'closed' },
        ]"
        class="session-filter"
      />
      <div class="session-list">
        <div
          v-if="loadingSessions"
          class="empty-state"
        >
          加载中...
        </div>
        <div
          v-else-if="!filteredSessions.length"
          class="empty-state"
        >
          暂无会话
        </div>
        <button
          v-for="session in filteredSessions"
          :key="session.sessionNo"
          type="button"
          class="session-item"
          :class="{ active: session.sessionNo === activeSession?.sessionNo }"
          @click="activeSessionNo = session.sessionNo"
        >
          <span class="session-title">{{ session.title }}</span>
          <span class="session-meta">{{ session.category }}</span>
          <span class="session-last">{{ session.lastMessage }}</span>
          <span
            v-if="session.userUnreadCount"
            class="unread"
          >{{
            session.userUnreadCount
          }}</span>
        </button>
      </div>
    </aside>

    <section class="chat-panel">
      <header class="chat-header">
        <div>
          <h2>{{ activeSession?.title || "请选择会话" }}</h2>
          <span>{{ activeSession?.sessionNo || "-" }}</span>
        </div>
        <div class="chat-actions">
          <el-button
            v-if="activeNeedsAccept"
            type="success"
            :disabled="!wsOnline || !agentId"
            @click="acceptSession"
          >
            接待
          </el-button>
          <el-button>转接</el-button>
          <el-button type="primary">
            结束会话
          </el-button>
        </div>
      </header>

      <div class="message-list">
        <div
          v-if="loadingMessages"
          class="empty-state"
        >
          加载消息中...
        </div>
        <div
          v-else-if="!activeSession"
          class="empty-state"
        >
          选择左侧会话后查看消息
        </div>
        <div
          v-for="message in activeMessages"
          :key="message.messageNo"
          class="message-row"
          :class="{ mine: message.senderType === 2 }"
        >
          <div class="bubble">
            {{ message.content }}
          </div>
        </div>
      </div>

      <footer class="composer">
        <el-input
          v-model="input"
          type="textarea"
          resize="none"
          :disabled="!activeSession || activeNeedsAccept"
          :autosize="{ minRows: 3, maxRows: 4 }"
          :placeholder="activeNeedsAccept ? '请先接待该会话' : '输入回复内容'"
          @keydown.ctrl.enter.prevent="send"
        />
        <div class="composer-actions">
          <el-button>快捷回复</el-button>
          <el-button
            type="primary"
            :disabled="!activeSession || activeNeedsAccept"
            @click="send"
          >
            发送
          </el-button>
        </div>
      </footer>
    </section>

    <aside class="customer-panel">
      <div class="panel-header">
        <h2>客户信息</h2>
      </div>
      <dl class="info-list">
        <div>
          <dt>用户 ID</dt>
          <dd>{{ activeSession?.userId || "-" }}</dd>
        </div>
        <div>
          <dt>问题分类</dt>
          <dd>{{ activeSession?.category || "-" }}</dd>
        </div>
        <div>
          <dt>优先级</dt>
          <dd>{{ activeSession?.priority === 3 ? "高" : "普通" }}</dd>
        </div>
        <div>
          <dt>分组 ID</dt>
          <dd>{{ activeSession?.groupId || "-" }}</dd>
        </div>
      </dl>

      <div class="note-block">
        <h3>接待备注</h3>
        <el-input
          type="textarea"
          resize="none"
          :rows="5"
          placeholder="记录客户偏好、订单号或处理进展"
        />
      </div>
    </aside>
  </section>
</template>
