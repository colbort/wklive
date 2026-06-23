<script setup lang="ts">
import { computed, ref } from "vue";
import { authChatMerchant } from "@/api/chat";
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
const merchant = ref<ChatMerchant | null>(null);
const authError = ref("");
const authing = ref(false);

const canOpen = computed(() => Boolean(apiKey.value.trim() && apiSecret.value.trim()));

async function authorize() {
  authError.value = "";
  authing.value = true;
  try {
    merchant.value = await authChatMerchant({
      apiKey: apiKey.value.trim(),
      apiSecret: apiSecret.value.trim(),
    });
  } catch (err) {
    merchant.value = null;
    authError.value = err instanceof Error ? err.message : "认证失败";
  } finally {
    authing.value = false;
  }
}

function openChat(mode: ChatMode) {
  const params = new URLSearchParams();
  params.set("page", "chat");
  params.set("mode", mode);
  params.set("apiKey", apiKey.value.trim());
  params.set("apiSecret", apiSecret.value.trim());
  if (userId.value.trim()) params.set("userId", userId.value.trim());
  if (nickname.value.trim()) params.set("nickname", nickname.value.trim());
  if (avatarUrl.value.trim()) params.set("avatarUrl", avatarUrl.value.trim());

  const wsUrl = new URLSearchParams(window.location.search).get("wsUrl");
  if (wsUrl) params.set("wsUrl", wsUrl);

  window.location.assign(`/?${params.toString()}`);
}
</script>

<template>
  <main class="test-page">
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
          :disabled="!canOpen"
          @click="openChat('mobile')"
        >
          手机
        </button>
        <button
          class="secondary-button"
          type="button"
          :disabled="!canOpen"
          @click="openChat('desktop')"
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
          >
        </label>
        <label>
          <span>API Secret</span>
          <input
            v-model="apiSecret"
            autocomplete="off"
            placeholder="输入商户 API Secret"
            type="password"
          >
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
          >
        </label>
        <label>
          <span>用户 ID</span>
          <input
            v-model="userId"
            autocomplete="off"
            placeholder="留空则以访客进入"
          >
        </label>
        <label>
          <span>头像 URL</span>
          <input
            v-model="avatarUrl"
            autocomplete="off"
            placeholder="可选"
          >
        </label>
      </div>
    </section>
  </main>
</template>
