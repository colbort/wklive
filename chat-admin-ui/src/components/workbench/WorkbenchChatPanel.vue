<script setup lang="ts">
import { ref, watch } from "vue";
import ChatMessageBubble from "@/components/workbench/ChatMessageBubble.vue";
import type { ChatMessage, ChatSession } from "@/types/chat";

const props = defineProps<{
  session?: ChatSession;
  messages: ChatMessage[];
  loading: boolean;
  activeNeedsAccept: boolean;
  activeClosed: boolean;
  canReply: boolean;
  canAccept: boolean;
  acceptDisabledReason: string;
  wsOnline: boolean;
  agentId: number;
  userId: number;
  showGuestRefreshNotice: boolean;
  showMobileBack?: boolean;
  resolveUrl?: (url: string) => Promise<string> | string;
}>();

const emit = defineEmits<{
  accept: [];
  back: [];
  close: [];
  send: [content: string];
  sendImage: [file: File];
}>();

const draft = ref("");
const imageInput = ref<HTMLInputElement>();

watch(
  () => props.session?.sessionNo,
  () => {
    draft.value = "";
  },
);

function isOwnAgentMessage(message: ChatMessage) {
  return (
    props.userId > 0 &&
    Number(message.senderType || message.sender?.type || 0) === 2 &&
    Number(message.sender?.id || 0) === props.userId
  );
}

function messageDirection(message: ChatMessage) {
  if (message.senderType === 3) return "system";
  if (isOwnAgentMessage(message)) {
    return "sent";
  }
  return "received";
}

function messageSenderName(message: ChatMessage) {
  if (message.senderType === 3) return "系统";
  if (isOwnAgentMessage(message)) return "我";
  return message.sender?.nickname || "用户";
}

function submit() {
  const content = draft.value.trim();
  if (!content || !props.canReply) return;
  emit("send", content);
  draft.value = "";
}

function openImagePicker() {
  if (!props.canReply) return;
  imageInput.value?.click();
}

function handleImageSelected(event: Event) {
  const input = event.target as HTMLInputElement;
  const file = input.files?.[0];
  input.value = "";
  if (!file) return;
  emit("sendImage", file);
}

function formatNickname(nickname: string) {
  if (nickname.length <=5) return nickname
  return nickname.slice(0, 5)
}
</script>

<template>
  <section class="chat-panel workbench-region">
    <input
      ref="imageInput"
      class="resource-input"
      type="file"
      accept="image/*"
      @change="handleImageSelected"
    />
    <header class="chat-header">
      <button
        v-if="showMobileBack"
        class="mobile-chat-back"
        type="button"
        @click="emit('back')"
      >
        &lt; 会话
      </button>
      <div>
        <h2>{{ formatNickname(session?.extJson?.nickname || "请选择会话") }}</h2>
        <span class="chat-session-no">{{ session?.sessionNo || "-" }}</span>
      </div>
      <div class="chat-actions">
        <el-tooltip
          v-if="activeNeedsAccept"
          :content="acceptDisabledReason"
          :disabled="canAccept || !acceptDisabledReason"
          placement="top"
        >
          <span>
            <el-button
              type="success"
              :disabled="!canAccept"
              @click="emit('accept')"
            >
              接待
            </el-button>
          </span>
        </el-tooltip>
        <el-button v-if="session && !activeClosed">
          转接
        </el-button>
        <el-button
          v-if="session && !activeClosed"
          type="primary"
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
      <ChatMessageBubble
        v-for="message in messages"
        :key="message.messageNo"
        :message="message"
        :direction="messageDirection(message)"
        :sender-name="messageSenderName(message)"
        :resolve-url="resolveUrl"
      />
    </div>

    <footer class="composer">
      <el-input
        v-model="draft"
        type="textarea"
        resize="none"
        :disabled="!canReply"
        :autosize="{ minRows: 3, maxRows: 4 }"
        :placeholder="activeClosed ? '会话已结束' : activeNeedsAccept ? '请先接待该会话' : '输入回复内容'"
        @keydown.ctrl.enter.prevent="submit"
      />
      <div class="composer-actions">
        <el-button
          :disabled="!canReply"
          @click="openImagePicker"
        >
          图片
        </el-button>
        <el-button
          type="primary"
          :disabled="!canReply"
          @click="submit"
        >
          发送
        </el-button>
      </div>
    </footer>
  </section>
</template>
