<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import { ElMessage, type FormInstance, type FormRules } from "element-plus";
import { getChatConfig, updateChatConfig } from "@/api/chat";
import type {
  ChatFeatureConfig,
  ChatMerchantConfig,
  ChatThemeConfig,
} from "@/types/chat";

type ColorField = keyof ChatThemeConfig;
type ActivePanel = "config" | "integration";

const colorPattern = /^#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6}|[0-9a-fA-F]{8})$/;

const colorFields: Array<{ key: ColorField; label: string }> = [
  { key: "backgroundColor", label: "聊天背景" },
  { key: "primaryColor", label: "主色" },
  { key: "noticeBarColor", label: "公告背景" },
  { key: "noticeTextColor", label: "公告文字" },
  { key: "agentBubbleColor", label: "客服气泡" },
  { key: "userBubbleColor", label: "用户气泡" },
];

const featureFields: Array<{ key: keyof ChatFeatureConfig; label: string }> = [
  { key: "enableCopy", label: "复制消息" },
  { key: "enableRevoke", label: "撤回消息" },
  { key: "enableDelete", label: "删除消息" },
  { key: "enableQuote", label: "引用消息" },
  { key: "enableForward", label: "转发消息" },
];

const defaultTheme: ChatThemeConfig = {
  backgroundColor: "#FFFFFF",
  primaryColor: "#128577",
  noticeBarColor: "#FF0000",
  noticeTextColor: "#00A3FF",
  agentBubbleColor: "#FFFFFF",
  userBubbleColor: "#128577",
};

const defaultFeatures: ChatFeatureConfig = {
  enableCopy: true,
  enableRevoke: true,
  enableDelete: true,
  enableQuote: false,
  enableForward: false,
};

const loading = ref(false);
const saving = ref(false);
const activePanel = ref<ActivePanel>("config");
const formRef = ref<FormInstance>();
const merchantConfig = ref<ChatMerchantConfig | null>(null);

const form = reactive({
  title: "在线客服",
  uiConfig: { ...defaultTheme },
  featureConfig: { ...defaultFeatures },
});

const rules: FormRules = {
  title: [
    { required: true, message: "请输入标题", trigger: "blur" },
    { max: 128, message: "标题不能超过 128 个字符", trigger: "blur" },
  ],
};

const previewStyle = computed(() => ({
  background: form.uiConfig.backgroundColor,
  "--chat-primary": form.uiConfig.primaryColor,
  "--chat-notice-bg": form.uiConfig.noticeBarColor,
  "--chat-notice-text": form.uiConfig.noticeTextColor,
  "--chat-agent-bubble": form.uiConfig.agentBubbleColor,
  "--chat-user-bubble": form.uiConfig.userBubbleColor,
}));
const apiKeyText = computed(
  () => merchantConfig.value?.apiKey || "<CHAT_API_KEY>",
);
const apiSecretText = computed(
  () => merchantConfig.value?.apiSecret || "<CHAT_API_SECRET>",
);
const tokenRequestExample = computed(
  () => `POST /chat/internal/tokens
Content-Type: application/json

{
  "apiKey": "${apiKeyText.value}",
  "apiSecret": "${apiSecretText.value}",
  "userId": 10001,
  "nickname": "张三",
  "avatarUrl": "https://example.com/avatar.png",
  "isGuest": false,  // 是否游客
  "ttlSeconds": 1800 // chatToken 过期时间，最长 1800 秒
}`,
);
const iframeExample = `<iframe
  src="https://chat.example.com/?page=chat&mode=mobile&chatToken=<CHAT_TOKEN>"
  style="width: 100%; height: 720px; border: 0;"
  allow="clipboard-write"
></iframe>`;
const configRequestExample = `GET /chat/config
Authorization: Bearer <CHAT_TOKEN>`;

onMounted(() => {
  void loadConfig();
});

async function loadConfig() {
  loading.value = true;
  try {
    const resp = await getChatConfig();
    merchantConfig.value = resp.data;
    Object.assign(form, {
      title: resp.data.title || "在线客服",
      uiConfig: {
        ...defaultTheme,
        ...(resp.data.uiConfig || {}),
      },
      featureConfig: {
        ...defaultFeatures,
        ...(resp.data.featureConfig || {}),
      },
    });
  } finally {
    loading.value = false;
  }
}

function validateColors() {
  const invalid = colorFields.find(
    (field) => !colorPattern.test(form.uiConfig[field.key] || ""),
  );
  if (invalid) {
    ElMessage.warning(`${invalid.label}需填写 Hex 颜色`);
    return false;
  }
  return true;
}

async function submit() {
  await formRef.value?.validate();
  if (!validateColors()) return;

  saving.value = true;
  try {
    const resp = await updateChatConfig({
      title: form.title.trim(),
      uiConfig: { ...form.uiConfig },
      featureConfig: { ...form.featureConfig },
    });
    merchantConfig.value = resp.data;
    ElMessage.success("保存成功");
  } finally {
    saving.value = false;
  }
}
</script>

<template>
  <section
    v-loading="loading"
    class="chat-config-page"
    :class="{ 'chat-config-page--integration': activePanel === 'integration' }"
  >
    <div class="chat-config-form">
      <div class="panel-tabs">
        <button
          class="panel-tab"
          :class="{ active: activePanel === 'config' }"
          type="button"
          @click="activePanel = 'config'"
        >
          基础信息
        </button>
        <button
          class="panel-tab"
          :class="{ active: activePanel === 'integration' }"
          type="button"
          @click="activePanel = 'integration'"
        >
          对接说明
        </button>
      </div>
      <el-form
        ref="formRef"
        class="chat-config-scroll"
        :model="form"
        :rules="rules"
        label-position="top"
      >
        <section class="config-section">
          <el-form-item
            v-if="activePanel === 'config'"
            label="标题"
            prop="title"
          >
            <el-input
              v-model.trim="form.title"
              maxlength="128"
              show-word-limit
            />
          </el-form-item>
        </section>

        <template v-if="activePanel === 'config'">
          <section class="config-section">
            <h2>界面颜色</h2>
            <div class="color-grid">
              <el-form-item
                v-for="field in colorFields"
                :key="field.key"
                :label="field.label"
              >
                <div class="color-control">
                  <el-color-picker
                    v-model="form.uiConfig[field.key]"
                    color-format="hex"
                    :predefine="[
                      '#FFFFFF',
                      '#128577',
                      '#FF0000',
                      '#00A3FF',
                      '#F3F5F8',
                      '#111827',
                    ]"
                  />
                  <el-input
                    v-model.trim="form.uiConfig[field.key]"
                    maxlength="9"
                  />
                </div>
              </el-form-item>
            </div>
          </section>

          <section class="config-section">
            <h2>功能开关</h2>
            <div class="feature-grid">
              <label
                v-for="field in featureFields"
                :key="field.key"
                class="feature-item"
              >
                <span>{{ field.label }}</span>
                <el-switch v-model="form.featureConfig[field.key]" />
              </label>
            </div>
          </section>
        </template>

        <section
          v-else
          class="config-section"
        >
          <h2>客服对接</h2>
          <div class="integration-guide">
            <div class="integration-step">
              <strong>1. 服务端换取 ChatToken</strong>
              <p>业务系统后端使用 API Key 和 API Secret 换取用户聊天 JWT。</p>
              <pre><code>{{ tokenRequestExample }}</code></pre>
            </div>
            <div class="integration-step">
              <strong>2. iframe 嵌入客服页</strong>
              <p>页面只传 ChatToken，不在浏览器暴露 API Secret。</p>
              <pre><code>{{ iframeExample }}</code></pre>
            </div>
            <div class="integration-step">
              <strong>3. chat-ui 读取配置</strong>
              <p>chat-ui 会使用同一个 JWT 获取标题、颜色和功能开关。</p>
              <pre><code>{{ configRequestExample }}</code></pre>
            </div>
          </div>
        </section>
      </el-form>
      <div
        v-if="activePanel === 'config'"
        class="config-actions"
      >
        <el-button @click="loadConfig">
          重置
        </el-button>
        <el-button
          type="primary"
          :loading="saving"
          @click="submit"
        >
          保存配置
        </el-button>
      </div>
    </div>

    <aside
      v-if="activePanel === 'config'"
      class="chat-config-preview"
      :style="previewStyle"
    >
      <header>
        {{ form.title }}
      </header>
      <div class="preview-notice">
        狗剩 正在为你服务
      </div>
      <div class="preview-body">
        <div class="preview-message agent-message">
          <span>狗剩</span>
          <p>您好，请问有什么可以帮您？</p>
        </div>
        <div class="preview-message user-message">
          <span>我</span>
          <p>想咨询订单处理进度</p>
        </div>
      </div>
      <footer>
        <button type="button">
          图片
        </button>
        <div>输入消息</div>
        <button type="button">
          结束
        </button>
      </footer>
    </aside>
  </section>
</template>

<style scoped>
.chat-config-page {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 360px;
  gap: 18px;
  height: 100%;
  min-height: 0;
}

.chat-config-page--integration {
  grid-template-columns: minmax(0, 1fr);
}

.chat-config-form,
.chat-config-preview {
  border: 1px solid #e6e9ef;
  border-radius: 8px;
  background: #fff;
}

.chat-config-form {
  display: grid;
  grid-template-rows: auto minmax(0, 1fr) auto;
  min-width: 0;
  min-height: 0;
  overflow: hidden;
}

.chat-config-page--integration .chat-config-form {
  grid-template-rows: auto minmax(0, 1fr);
}

.config-section + .config-section {
  margin-top: 24px;
}

.config-section h2 {
  margin-bottom: 14px;
  font-size: 15px;
  font-weight: 750;
}

.panel-tabs {
  display: inline-grid;
  grid-auto-flow: column;
  align-self: start;
  justify-self: start;
  gap: 0;
  margin: 20px 20px 0;
  overflow: hidden;
  border-radius: 4px;
  background: #f2f4f7;
}

.chat-config-scroll {
  min-height: 0;
  overflow-y: auto;
  padding: 20px 20px 28px;
}

.panel-tab {
  min-width: 72px;
  min-height: 44px;
  border: 0;
  padding: 0 18px;
  color: #667085;
  background: #f2f4f7;
  font-weight: 650;
}

.panel-tab.active {
  color: #fff;
  background: #409eff;
}

.color-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 4px 18px;
}

.color-control {
  display: grid;
  grid-template-columns: 40px minmax(0, 1fr);
  align-items: center;
  gap: 10px;
  width: 100%;
}

.feature-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.feature-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 42px;
  border: 1px solid #e6e9ef;
  border-radius: 6px;
  padding: 0 12px;
  color: #344054;
  background: #fbfcfe;
}

.integration-guide {
  display: grid;
  gap: 12px;
}

.integration-step {
  display: grid;
  gap: 8px;
  border: 1px solid #e6e9ef;
  border-radius: 8px;
  padding: 12px;
  background: #fbfcfe;
}

.integration-step strong {
  color: #1d2939;
  font-weight: 750;
}

.integration-step p {
  color: #667085;
  font-size: 13px;
}

.integration-step pre {
  max-width: 100%;
  margin: 0;
  overflow: auto;
  border-radius: 6px;
  padding: 12px;
  color: #d6e4ff;
  background: #101828;
}

.integration-step code {
  font-family:
    ui-monospace,
    SFMono-Regular,
    Menlo,
    Monaco,
    Consolas,
    "Liberation Mono",
    monospace;
  font-size: 12px;
  line-height: 1.55;
  white-space: pre;
}

.config-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  border-top: 1px solid #e6e9ef;
  padding: 14px 20px;
  background: #fff;
}

.chat-config-preview {
  display: grid;
  grid-template-rows: 56px 34px minmax(260px, 1fr) 64px;
  height: 100%;
  min-height: 0;
  overflow: hidden;
}

.chat-config-preview header {
  display: grid;
  place-items: center;
  color: #fff;
  background: #0d111b;
  font-size: 18px;
  font-weight: 800;
}

.preview-notice {
  display: grid;
  place-items: center;
  color: var(--chat-notice-text);
  background: var(--chat-notice-bg);
  font-weight: 700;
}

.preview-body {
  display: flex;
  flex-direction: column;
  gap: 18px;
  padding: 24px 18px;
}

.preview-message {
  max-width: 74%;
  border-radius: 8px;
  padding: 12px 14px;
  box-shadow: 0 12px 24px rgba(15, 23, 42, 0.08);
}

.preview-message span {
  display: block;
  margin-bottom: 6px;
  color: #7b8798;
  font-size: 12px;
}

.preview-message p {
  color: #182230;
  font-size: 15px;
}

.agent-message {
  align-self: flex-start;
  background: var(--chat-agent-bubble);
}

.user-message {
  align-self: flex-end;
  background: var(--chat-user-bubble);
}

.user-message span,
.user-message p {
  color: #fff;
}

.chat-config-preview footer {
  display: grid;
  grid-template-columns: 68px minmax(0, 1fr) 68px;
  gap: 10px;
  border-top: 1px solid #e6e9ef;
  padding: 10px;
  background: #fff;
}

.chat-config-preview footer button,
.chat-config-preview footer div {
  display: grid;
  place-items: center;
  min-width: 0;
  border: 1px solid #d6dde8;
  border-radius: 6px;
  color: var(--chat-primary);
  background: #fff;
  font-weight: 700;
}

.chat-config-preview footer div {
  justify-content: start;
  padding: 0 14px;
  color: #8a94a6;
  font-weight: 500;
}

.chat-config-preview footer button {
  border: 0;
  color: #fff;
  background: var(--chat-primary);
}

@media (max-width: 980px) {
  .chat-config-page {
    grid-template-columns: 1fr;
  }

  .chat-config-preview {
    min-height: 460px;
  }
}

@media (max-width: 640px) {
  .chat-config-form {
    padding: 14px;
  }

  .color-grid,
  .feature-grid {
    grid-template-columns: 1fr;
  }
}
</style>
