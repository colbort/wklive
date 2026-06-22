<script setup lang="ts">
import { computed, ref } from "vue";
import { authChatMerchant } from "@/api/chat";
import { useChatSocket } from "@/composables/useChatSocket";
import type { ChatMerchant } from "@/types/chat";

const apiKey = ref("ck_b56f695fef22019e1326354e318ab0c6");
const apiSecret = ref("cs_cfdea0dd4b4c6b687ba5cb0d4b9c1e1cd0d58e61a6fa56e1faf5575575e1df7a");
const userToken = ref("");
const nickname = ref("访客");
const draft = ref("");
const merchant = ref<ChatMerchant | null>(null);
const authError = ref("");
const authing = ref(false);

const chat = useChatSocket();

const canConnect = computed(() => Boolean(merchant.value?.merchantId));
const connectionLabel = computed(() => {
  if (chat.status.value === "connecting") {
    return "连接中";
  }
  if (chat.status.value === "reconnecting") {
    return chat.reconnectLabel.value || "重连中";
  }
  if (chat.isOpen.value) {
    return chat.isTemporary.value ? "临时会话" : "登录会话";
  }
  if (chat.status.value === "closed") {
    return "已断开";
  }
  return "未连接";
});

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

function connect() {
  if (!merchant.value) {
    return;
  }
  chat.connect(merchant.value.merchantId, userToken.value);
}

function sendMessage() {
  chat.sendText(draft.value, nickname.value);
  draft.value = "";
}
</script>

<template>
  <main class="chat-page">
    <section class="setup-panel" aria-label="chat setup">
      <div class="brand">
        <div class="brand-mark">WK</div>
        <div>
          <h1>客户客服</h1>
          <p>商户接入验证后即可发起对话</p>
        </div>
      </div>

      <form class="setup-form" @submit.prevent="authorize">
        <label>
          <span>API Key</span>
          <input v-model="apiKey" autocomplete="off" placeholder="输入商户 API Key" />
        </label>
        <label>
          <span>API Secret</span>
          <input
            v-model="apiSecret"
            autocomplete="off"
            placeholder="输入商户 API Secret"
            type="password"
          />
        </label>
        <button class="primary-button" type="submit" :disabled="authing">
          {{ authing ? "验证中" : "验证商户" }}
        </button>
      </form>

      <div v-if="merchant" class="merchant-box">
        <span>Merchant</span>
        <strong>{{ merchant.merchantId }}</strong>
      </div>
      <p v-if="authError" class="error-text">{{ authError }}</p>

      <div class="connect-box">
        <label>
          <span>昵称</span>
          <input v-model="nickname" autocomplete="off" />
        </label>
        <label>
          <span>用户 Token</span>
          <input v-model="userToken" autocomplete="off" placeholder="留空则以访客进入" />
        </label>
        <button class="secondary-button" :disabled="!canConnect" type="button" @click="connect">
          连接客服
        </button>
      </div>
    </section>

    <section class="chat-shell" aria-label="chat conversation">
      <header class="chat-header">
        <div>
          <p>WkLive Support</p>
          <strong>{{ connectionLabel }}</strong>
        </div>
        <div class="session-meta">
          <span>{{ chat.connected.value?.sessionNo || "等待连接" }}</span>
          <button type="button" :disabled="!chat.isOpen.value" @click="chat.close">断开</button>
        </div>
      </header>

      <div class="message-list">
        <div class="welcome-message">
          <strong>您好</strong>
          <span>请描述您遇到的问题，客服会在这里接收并回复。</span>
        </div>

        <div v-if="chat.queueStatus.value" class="queue-message">
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

      <p v-if="chat.error.value" class="error-line">{{ chat.error.value }}</p>

      <form class="composer" @submit.prevent="sendMessage">
        <textarea
          v-model="draft"
          :disabled="!chat.isOpen.value"
          placeholder="输入消息"
          rows="3"
        />
        <button class="send-button" :disabled="!chat.isOpen.value || !draft.trim()" type="submit">
          发送
        </button>
      </form>
    </section>
  </main>
</template>
