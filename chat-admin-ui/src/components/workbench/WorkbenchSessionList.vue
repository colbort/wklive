<script setup lang="ts">
import type { ChatSession } from "@/types/chat";

defineProps<{
  sessions: ChatSession[];
  selectedSessionNo: string;
  loading: boolean;
  wsOnline: boolean;
  statusFilter: string;
}>();

const emit = defineEmits<{
  refresh: [];
  select: [sessionNo: string];
  "update:statusFilter": [value: string];
}>();

const statusOptions = [
  { label: "待接待", value: "waiting" },
  { label: "进行中", value: "serving" },
  { label: "已结束", value: "closed" },
];
</script>

<template>
  <aside class="session-panel workbench-region">
    <div class="panel-header">
      <h2>会话</h2>
      <div class="panel-tools">
        <el-button
          size="small"
          :loading="loading"
          @click="emit('refresh')"
        >
          刷新
        </el-button>
        <el-tag :type="wsOnline ? 'success' : 'info'">
          {{ wsOnline ? "在线" : "离线" }}
        </el-tag>
      </div>
    </div>
    <el-segmented
      :model-value="statusFilter"
      :options="statusOptions"
      class="session-filter"
      @update:model-value="emit('update:statusFilter', String($event))"
    />
    <div class="session-list">
      <div
        v-if="loading"
        class="empty-state"
      >
        加载中...
      </div>
      <div
        v-else-if="!sessions.length"
        class="empty-state"
      >
        暂无会话
      </div>
      <button
        v-for="session in sessions"
        :key="session.sessionNo"
        type="button"
        class="session-item"
        :class="{ active: session.sessionNo === selectedSessionNo }"
        @click="emit('select', session.sessionNo)"
      >
        <span class="session-title">{{ session.title }}</span>
        <span class="session-meta">{{ session.category }}</span>
        <span class="session-last">{{ session.lastMessage }}</span>
        <span
          v-if="session.userUnreadCount"
          class="unread"
        >
          {{ session.userUnreadCount }}
        </span>
      </button>
    </div>
  </aside>
</template>
