<script setup lang="ts">
import { computed, ref } from "vue";
import type { ChatMessage, ChatSession } from "@/types/chat";

const statusFilter = ref("serving");
const input = ref("");
const activeSessionNo = ref("CS20260621001");

const sessions: ChatSession[] = [
  {
    id: 1,
    sessionNo: "CS20260621001",
    merchantId: 10001,
    userId: 9001,
    status: 2,
    priority: 2,
    agentId: 3001,
    groupId: 1,
    title: "订单发货咨询",
    category: "订单问题",
    lastMessage: "我的订单什么时候发货？",
    lastMessageTime: Date.now() - 180000,
    userUnreadCount: 2,
  },
  {
    id: 2,
    sessionNo: "CS20260621002",
    merchantId: 10001,
    userId: 9002,
    status: 1,
    priority: 3,
    agentId: 0,
    groupId: 2,
    title: "退款进度",
    category: "售后服务",
    lastMessage: "麻烦帮我看下退款。",
    lastMessageTime: Date.now() - 600000,
    userUnreadCount: 1,
  },
  {
    id: 3,
    sessionNo: "CS20260620009",
    merchantId: 10001,
    userId: 9003,
    status: 5,
    priority: 2,
    agentId: 3001,
    groupId: 1,
    title: "活动规则",
    category: "售前咨询",
    lastMessage: "好的，谢谢。",
    lastMessageTime: Date.now() - 86400000,
    userUnreadCount: 0,
  },
];

const messages = ref<Record<string, ChatMessage[]>>({
  CS20260621001: [
    {
      id: 1,
      messageNo: "M1",
      sessionNo: "CS20260621001",
      senderType: 1,
      content: "你好，我的订单什么时候发货？",
      createTimes: Date.now() - 420000,
    },
    {
      id: 2,
      messageNo: "M2",
      sessionNo: "CS20260621001",
      senderType: 2,
      content: "您好，我帮您核实一下订单状态。",
      createTimes: Date.now() - 360000,
    },
    {
      id: 3,
      messageNo: "M3",
      sessionNo: "CS20260621001",
      senderType: 1,
      content: "订单号是 202606210088。",
      createTimes: Date.now() - 180000,
    },
  ],
  CS20260621002: [
    {
      id: 4,
      messageNo: "M4",
      sessionNo: "CS20260621002",
      senderType: 1,
      content: "麻烦帮我看下退款。",
      createTimes: Date.now() - 600000,
    },
  ],
  CS20260620009: [
    {
      id: 5,
      messageNo: "M5",
      sessionNo: "CS20260620009",
      senderType: 2,
      content: "活动规则已经发您了。",
      createTimes: Date.now() - 86600000,
    },
    {
      id: 6,
      messageNo: "M6",
      sessionNo: "CS20260620009",
      senderType: 1,
      content: "好的，谢谢。",
      createTimes: Date.now() - 86400000,
    },
  ],
});

const filteredSessions = computed(() => {
  const statusMap: Record<string, number[]> = {
    waiting: [1],
    serving: [2, 3, 4],
    closed: [5],
  };
  return sessions.filter((item) =>
    statusMap[statusFilter.value].includes(item.status),
  );
});

const activeSession = computed(
  () =>
    sessions.find((item) => item.sessionNo === activeSessionNo.value) ||
    sessions[0],
);
const activeMessages = computed(
  () => messages.value[activeSession.value.sessionNo] || [],
);

function send() {
  const content = input.value.trim();
  if (!content) return;
  const sessionNo = activeSession.value.sessionNo;
  messages.value[sessionNo] = [
    ...(messages.value[sessionNo] || []),
    {
      id: Date.now(),
      messageNo: `M${Date.now()}`,
      sessionNo,
      senderType: 2,
      content,
      createTimes: Date.now(),
    },
  ];
  input.value = "";
}
</script>

<template>
  <section class="workbench">
    <aside class="session-panel">
      <div class="panel-header">
        <h2>会话</h2>
        <el-tag type="success">
          在线
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
        <button
          v-for="session in filteredSessions"
          :key="session.sessionNo"
          type="button"
          class="session-item"
          :class="{ active: session.sessionNo === activeSession.sessionNo }"
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
          <h2>{{ activeSession.title }}</h2>
          <span>{{ activeSession.sessionNo }}</span>
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
          :autosize="{ minRows: 3, maxRows: 4 }"
          placeholder="输入回复内容"
          @keydown.ctrl.enter.prevent="send"
        />
        <div class="composer-actions">
          <el-button>快捷回复</el-button>
          <el-button
            type="primary"
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
          <dd>{{ activeSession.userId }}</dd>
        </div>
        <div>
          <dt>问题分类</dt>
          <dd>{{ activeSession.category }}</dd>
        </div>
        <div>
          <dt>优先级</dt>
          <dd>{{ activeSession.priority === 3 ? "高" : "普通" }}</dd>
        </div>
        <div>
          <dt>分组 ID</dt>
          <dd>{{ activeSession.groupId }}</dd>
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
