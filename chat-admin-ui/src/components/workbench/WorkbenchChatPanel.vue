<script setup lang="ts">
import type { ChatMessage, ChatSession } from "@/types/chat";

defineProps<{
  session?: ChatSession;
  messages: ChatMessage[];
  loading: boolean;
  inputValue: string;
  activeNeedsAccept: boolean;
  activeClosed: boolean;
  canReply: boolean;
  wsOnline: boolean;
  agentId: number;
  showGuestRefreshNotice: boolean;
}>();

const emit = defineEmits<{
  accept: [];
  close: [];
  send: [];
  "update:inputValue": [value: string];
}>();
</script>

<template>
  <section class="chat-panel workbench-region">
    <header class="chat-header">
      <div>
        <h2>{{ session?.title || "请选择会话" }}</h2>
        <span>{{ session?.sessionNo || "-" }}</span>
      </div>
      <div class="chat-actions">
        <el-button
          v-if="activeNeedsAccept"
          type="success"
          :disabled="!wsOnline || !agentId"
          @click="emit('accept')"
        >
          接待
        </el-button>
        <el-button>转接</el-button>
        <el-button
          type="primary"
          :disabled="!session || activeClosed"
          @click="emit('close')"
        >
          结束会话
        </el-button>
      </div>
    </header>

    <div class="message-list">
      <div
        v-if="loading"
        class="empty-state"
      >
        加载消息中...
      </div>
      <div
        v-else-if="!session"
        class="empty-state"
      >
        选择左侧会话后查看消息
      </div>
      <div
        v-else-if="activeNeedsAccept"
        class="empty-state"
      >
        接待后查看历史消息
      </div>
      <div
        v-else-if="showGuestRefreshNotice"
        class="empty-state"
      >
        临时会话刷新后只恢复会话摘要，后续新消息会继续显示
      </div>
      <div
        v-else-if="!messages.length"
        class="empty-state"
      >
        暂无消息
      </div>
      <div
        v-for="message in messages"
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
        :model-value="inputValue"
        type="textarea"
        resize="none"
        :disabled="!canReply"
        :autosize="{ minRows: 3, maxRows: 4 }"
        :placeholder="activeClosed ? '会话已结束' : activeNeedsAccept ? '请先接待该会话' : '输入回复内容'"
        @update:model-value="emit('update:inputValue', String($event))"
        @keydown.ctrl.enter.prevent="emit('send')"
      />
      <div class="composer-actions">
        <el-button>快捷回复</el-button>
        <el-button
          type="primary"
          :disabled="!canReply"
          @click="emit('send')"
        >
          发送
        </el-button>
      </div>
    </footer>
  </section>
</template>
