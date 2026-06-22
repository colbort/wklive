<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from "vue";
import { authChatMerchant } from "@/api/chat";
import { useChatSocket } from "@/composables/useChatSocket";
import type { ChatMerchant } from "@/types/chat";

type ChatMode = "mobile" | "desktop";

const apiKey = ref("ck_b56f695fef22019e1326354e318ab0c6");
const apiSecret = ref("cs_cfdea0dd4b4c6b687ba5cb0d4b9c1e1cd0d58e61a6fa56e1faf5575575e1df7a");
const userToken = ref("");
const nickname = ref("访客");
const draft = ref("");
const resourceInput = ref<HTMLInputElement>();
const selectedResourceName = ref("");
const merchant = ref<ChatMerchant | null>(null);
const authError = ref("");
const authing = ref(false);
const activeMode = ref<ChatMode | null>(null);

const chat = useChatSocket();

const canOpen = computed(() => Boolean(apiKey.value.trim() && apiSecret.value.trim()));
const showMobilePage = computed(() => activeMode.value === "mobile");
const showDesktopDialog = computed(() => activeMode.value === "desktop");
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

async function openMobileChat() {
  const nextMerchant = await ensureMerchant();
  if (!nextMerchant) return;
  activeMode.value = "mobile";
  connect(nextMerchant);
  window.history.pushState({ wkliveChatMode: "mobile" }, "");
  window.addEventListener("popstate", closeFromHistory);
}

async function openDesktopChat() {
  const nextMerchant = await ensureMerchant();
  if (!nextMerchant) return;
  activeMode.value = "desktop";
  connect(nextMerchant);
}

function connect(nextMerchant = merchant.value) {
  if (!nextMerchant) return;
  chat.connect(nextMerchant.merchantId, userToken.value);
}

function closeChat() {
  chat.close();
  activeMode.value = null;
  window.removeEventListener("popstate", closeFromHistory);
}

function closeMobileChat() {
  if (activeMode.value !== "mobile") return;
  window.history.back();
}

function closeDesktopChat() {
  closeChat();
}

function closeFromHistory() {
  if (activeMode.value === "mobile") {
    closeChat();
  }
}

function sendMessage() {
  chat.sendText(draft.value, nickname.value);
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

function hydrateFromQuery() {
  const params = new URLSearchParams(window.location.search);
  const queryApiKey = params.get("apiKey");
  const queryApiSecret = params.get("apiSecret");
  const queryToken = params.get("token");
  const queryNickname = params.get("nickname");
  const queryMode = params.get("mode");

  if (queryApiKey) apiKey.value = queryApiKey;
  if (queryApiSecret) apiSecret.value = queryApiSecret;
  if (queryToken) userToken.value = queryToken;
  if (queryNickname) nickname.value = queryNickname;

  if (params.get("auto") !== "1") return;

  if (queryMode === "mobile") {
    void openMobileChat();
    return;
  }

  if (queryMode === "desktop") {
    void openDesktopChat();
  }
}

onMounted(() => {
  hydrateFromQuery();
});

onBeforeUnmount(() => {
  window.removeEventListener("popstate", closeFromHistory);
  chat.close();
});
</script>

<template>
  <input
    ref="resourceInput"
    class="resource-input"
    type="file"
    @change="handleResourceSelected"
  />

  <main
    v-if="!showMobilePage"
    class="test-page"
  >
    <section class="test-hero">
      <div class="brand">
        <div class="brand-mark">
          WK
        </div>
        <div>
          <h1>客户客服测试页</h1>
          <p>选择手机或电脑形态打开客服页面</p>
        </div>
      </div>

      <div class="launch-actions">
        <button
          class="primary-button"
          type="button"
          :disabled="!canOpen || authing"
          @click="openMobileChat"
        >
          手机
        </button>
        <button
          class="secondary-button"
          type="button"
          :disabled="!canOpen || authing"
          @click="openDesktopChat"
        >
          电脑
        </button>
      </div>
    </section>

    <section
      class="setup-panel"
      aria-label="chat setup"
    >
      <form
        class="setup-form"
        @submit.prevent="authorize"
      >
        <label>
          <span>API Key</span>
          <input
            v-model="apiKey"
            autocomplete="off"
            placeholder="输入商户 API Key"
          />
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
        <button
          class="primary-button"
          type="submit"
          :disabled="authing"
        >
          {{ authing ? "验证中" : "验证商户" }}
        </button>
      </form>

      <div
        v-if="merchant"
        class="merchant-box"
      >
        <span>Merchant</span>
        <strong>{{ merchant.merchantId }}</strong>
      </div>
      <p
        v-if="authError"
        class="error-text"
      >
        {{ authError }}
      </p>

      <div class="connect-box">
        <label>
          <span>昵称</span>
          <input
            v-model="nickname"
            autocomplete="off"
          />
        </label>
        <label>
          <span>用户 Token</span>
          <input
            v-model="userToken"
            autocomplete="off"
            placeholder="留空则以访客进入"
          />
        </label>
      </div>
    </section>
  </main>

  <main
    v-if="showMobilePage"
    class="mobile-chat-page"
  >
    <section
      class="chat-shell"
      aria-label="chat conversation"
    >
      <header class="chat-header">
        <button
          class="back-button"
          type="button"
          @click="closeMobileChat"
        >
          返回
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
        v-if="chat.error.value"
        class="error-line"
      >
        {{ chat.error.value }}
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

  <div
    v-if="showDesktopDialog"
    class="dialog-mask"
    @click.self="closeDesktopChat"
  >
    <section
      class="desktop-dialog"
      role="dialog"
      aria-modal="true"
      aria-label="客服会话"
    >
      <section
        class="chat-shell"
        aria-label="chat conversation"
      >
        <header class="chat-header">
          <button
            class="back-button"
            type="button"
            @click="closeDesktopChat"
          >
            关闭
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
          v-if="chat.error.value"
          class="error-line"
        >
          {{ chat.error.value }}
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
    </section>
  </div>
</template>
