<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from "vue";
import { authChatMerchant } from "@/api/chat";
import { useChatSocket } from "@/composables/useChatSocket";
import type { ChatMerchant } from "@/types/chat";

type ChatMode = "mobile" | "desktop";

const defaultApiKey = "ck_b56f695fef22019e1326354e318ab0c6";
const defaultApiSecret =
  "cs_cfdea0dd4b4c6b687ba5cb0d4b9c1e1cd0d58e61a6fa56e1faf5575575e1df7a";

const apiKey = ref(defaultApiKey);
const apiSecret = ref(defaultApiSecret);
const userId = ref("");
const nickname = ref("访客");
const avatarUrl = ref("");
const draft = ref("");
const resourceInput = ref<HTMLInputElement>();
const selectedResourceName = ref("");
const merchant = ref<ChatMerchant | null>(null);
const authError = ref("");
const authing = ref(false);
const activeMode = ref<ChatMode>("mobile");

const chat = useChatSocket();

const showDesktopFrame = computed(() => activeMode.value === "desktop");
const connectionLabel = computed(() => {
  if (chat.status.value === "connecting") return "连接中";
  if (chat.status.value === "reconnecting") {
    return chat.reconnectLabel.value || "重连中";
  }
  if (chat.isOpen.value) {
    if (chat.sessionClosed.value) return "已结束";
    return chat.isTemporary.value ? "临时会话" : "登录会话";
  }
  if (chat.status.value === "closed") return "已断开";
  return "未连接";
});

function hydrateFromQuery() {
  const params = new URLSearchParams(window.location.search);
  const queryApiKey = params.get("apiKey");
  const queryApiSecret = params.get("apiSecret");
  const queryUserId = params.get("userId");
  const queryNickname = params.get("nickname");
  const queryAvatarUrl = params.get("avatarUrl");
  const queryMode = params.get("mode");

  if (queryApiKey) apiKey.value = queryApiKey;
  if (queryApiSecret) apiSecret.value = queryApiSecret;
  if (queryUserId) userId.value = queryUserId;
  if (queryNickname) nickname.value = queryNickname;
  if (queryAvatarUrl) avatarUrl.value = queryAvatarUrl;
  activeMode.value = queryMode === "desktop" ? "desktop" : "mobile";
}

async function authorize() {
  authError.value = "";
  authing.value = true;
  try {
    merchant.value = await authChatMerchant({
      apiKey: apiKey.value.trim(),
      apiSecret: apiSecret.value.trim(),
    });
    chat.resetMessages();
  } catch (err) {
    merchant.value = null;
    authError.value = err instanceof Error ? err.message : "认证失败";
  } finally {
    authing.value = false;
  }
}

async function ensureMerchant() {
  if (merchant.value?.merchantId) return merchant.value;
  await authorize();
  return merchant.value;
}

async function connectChat() {
  const nextMerchant = await ensureMerchant();
  if (!nextMerchant) return;
  chat.connect(nextMerchant.merchantId, {
    userId: userId.value,
    nickname: nickname.value,
    avatarUrl: avatarUrl.value,
  });
}

function closeChatPage() {
  chat.close();
  if (window.history.length > 1) {
    window.history.back();
    return;
  }
  window.location.assign("/");
}

function sendMessage() {
  chat.sendText(draft.value, nickname.value, avatarUrl.value);
  draft.value = "";
}

function openResourcePicker() {
  if (!chat.isOpen.value || chat.sessionClosed.value) return;
  resourceInput.value?.click();
}

function handleResourceSelected(event: Event) {
  const input = event.target as HTMLInputElement;
  selectedResourceName.value = input.files?.[0]?.name || "";
}

onMounted(() => {
  hydrateFromQuery();
  void connectChat();
});

onBeforeUnmount(() => {
  chat.close();
});
</script>

<template>
  <input
    ref="resourceInput"
    class="resource-input"
    type="file"
    @change="handleResourceSelected"
  >

  <main
    class="chat-page"
    :class="{ 'chat-page--desktop': showDesktopFrame }"
  >
    <section
      class="chat-shell"
      aria-label="chat conversation"
    >
      <header class="chat-header">
        <button
          class="back-button"
          type="button"
          @click="closeChatPage"
        >
          {{ activeMode === "desktop" ? "关闭" : "返回" }}
        </button>
        <div>
          <p>WkLive Support</p>
          <strong>{{ connectionLabel }}</strong>
        </div>
        <div class="session-meta">
          <span>{{ chat.connected.value?.sessionNo || "等待连接" }}</span>
          <button
            type="button"
            :disabled="!chat.isOpen.value"
            @click="chat.close"
          >
            断开
          </button>
        </div>
      </header>

      <div class="message-list">
        <div class="welcome-message">
          <strong>您好</strong>
          <span>请描述您遇到的问题，客服会在这里接收并回复。</span>
        </div>

        <div
          v-if="chat.queueStatus.value"
          class="queue-message"
        >
          {{ chat.queueStatus.value }}
        </div>

        <article
          v-for="message in chat.messages.value"
          :key="message.messageNo"
          class="message-row"
          :class="{
            mine: message.senderType === 1,
            system: message.senderType === 3,
          }"
        >
          <div class="bubble">
            <span>{{ message.sender?.nickname || (message.senderType === 1 ? nickname : "客服") }}</span>
            <p>{{ message.content }}</p>
          </div>
        </article>
      </div>

      <p
        v-if="chat.error.value || authError"
        class="error-line"
      >
        {{ chat.error.value || authError }}
      </p>

      <form
        class="composer"
        @submit.prevent="sendMessage"
      >
        <button
          class="resource-button"
          type="button"
          :disabled="!chat.isOpen.value || chat.sessionClosed.value"
          @click="openResourcePicker"
        >
          资源
        </button>
        <div class="composer-input">
          <textarea
            v-model="draft"
            :disabled="!chat.isOpen.value || chat.sessionClosed.value"
            :placeholder="chat.sessionClosed.value ? '本次会话已结束' : '输入消息'"
            rows="2"
          />
          <span
            v-if="selectedResourceName"
            class="resource-name"
          >
            {{ selectedResourceName }}
          </span>
        </div>
        <button
          class="send-button"
          :disabled="!chat.isOpen.value || chat.sessionClosed.value || !draft.trim()"
          type="submit"
        >
          发送
        </button>
      </form>
    </section>
  </main>
</template>
