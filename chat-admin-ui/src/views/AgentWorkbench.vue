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

const statusFilter = ref("serving");
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

const statusMap: Record<string, number> = {
  waiting: 1,
  serving: 2,
  closed: 5,
};

const filteredSessions = computed(() => sessions.value);

const activeSession = computed(
  () =>
    sessions.value.find((item) => item.sessionNo === activeSessionNo.value) ||
    sessions.value[0],
);
const activeMessages = computed(
  () =>
    activeSession.value ? messages.value[activeSession.value.sessionNo] || [] : [],
);
const merchantId = computed(() => auth.user?.merchantId || 0);
const agentId = computed(() => auth.agent?.id || 0);
const wsOnline = computed(() => wsState.value === "open");

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
  activeSessionNo.value = "";
  await loadSessions();
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
      agentId: statusFilter.value === "serving" ? agentId.value : undefined,
      status: statusMap[statusFilter.value],
      limit: 50,
    });
    sessions.value = resp.data;
    if (!activeSessionNo.value && resp.data.length) {
      activeSessionNo.value = resp.data[0].sessionNo;
    }
    if (
      activeSessionNo.value &&
      !resp.data.some((item) => item.sessionNo === activeSessionNo.value)
    ) {
      activeSessionNo.value = resp.data[0]?.sessionNo || "";
    }
  } catch (err) {
    ElMessage.error(err instanceof Error ? err.message : "加载会话失败");
  } finally {
    loadingSessions.value = false;
  }
}

async function loadMessages(sessionNo: string) {
  if (!merchantId.value) return;
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
    const event = JSON.parse(payload) as { type: string; data?: ChatMessage };
    if (event.type !== "chat.message") return;
    if (event.data?.sessionNo === activeSessionNo.value) {
      pushMessage(event.data);
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
  const list = messages.value[message.sessionNo] || [];
  if (list.some((item) => item.messageNo === message.messageNo)) return;
  messages.value[message.sessionNo] = [...list, message];
}

async function send() {
  const content = input.value.trim();
  if (!content || !activeSession.value || !merchantId.value || !agentId.value) {
    return;
  }
  const sessionNo = activeSession.value.sessionNo;
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
          :disabled="!activeSession"
          :autosize="{ minRows: 3, maxRows: 4 }"
          placeholder="输入回复内容"
          @keydown.ctrl.enter.prevent="send"
        />
        <div class="composer-actions">
          <el-button>快捷回复</el-button>
          <el-button
            type="primary"
            :disabled="!activeSession"
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
