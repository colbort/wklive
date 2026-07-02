<script setup lang="ts">
import {
  computed,
  nextTick,
  onBeforeUnmount,
  onMounted,
  ref,
  watch,
} from "vue";
import { setChatTokenCookie } from "@/api/chat";
import { useChatSocket } from "@/composables/useChatSocket";

type ChatMode = "mobile" | "desktop";

const draft = ref("");
const resourceInput = ref<HTMLInputElement>();
const messageInput = ref<HTMLTextAreaElement>();
const messageList = ref<HTMLDivElement>();
const selectedResourceName = ref("");
const authError = ref("");
const activeMode = ref<ChatMode>("mobile");
const rating = ref(5);
const evaluationComment = ref("");

const chat = useChatSocket();

const showDesktopFrame = computed(() => activeMode.value === "desktop");
const hasDraft = computed(() => draft.value.trim().length > 0);
const composerActionLabel = computed(() => (hasDraft.value ? "发送" : "结束"));
const canSubmitEvaluation = computed(
  () =>
    chat.isOpen.value &&
    !chat.evaluationSubmitting.value &&
    !chat.evaluationSubmitted.value &&
    rating.value >= 1 &&
    rating.value <= 5,
);
const showEvaluationPanel = computed(
  () =>
    Boolean(chat.evaluationInvite.value) ||
    chat.sessionClosed.value ||
    chat.evaluationSubmitted.value,
);

function messageDirection(message: {
  senderType: number;
  sender?: { id?: number };
}) {
  if (message.senderType === 3) return "system";
  const userId = chat.connected.value?.userId || 0;
  if (userId > 0 && message.sender?.id === userId) return "sent";
  return "received";
}

function messageSenderName(message: {
  senderType: number;
  sender?: { id?: number; nickname?: string };
}) {
  if (message.senderType === 3) return "系统";
  const userId = chat.connected.value?.userId || 0;
  if (userId > 0 && message.sender?.id === userId) return "我";
  return message.sender?.nickname || "客服";
}

function hydrateFromQuery() {
  const params = new URLSearchParams(window.location.search);
  const queryChatToken = params.get("chatToken");
  const queryMode = params.get("mode");
  activeMode.value = queryMode === "desktop" ? "desktop" : "mobile";
  return queryChatToken?.trim() || "";
}

async function connectChat() {
  authError.value = "";
  const token = hydrateFromQuery();
  if (token) {
    try {
      await setChatTokenCookie(token);
      removeChatTokenFromUrl();
    } catch (err) {
      authError.value = err instanceof Error ? err.message : "认证失败";
      return;
    }
  }
  chat.resetMessages();
  chat.connect(token);
}

function removeChatTokenFromUrl() {
  const url = new URL(window.location.href);
  url.searchParams.delete("chatToken");
  window.history.replaceState({}, "", `${url.pathname}${url.search}${url.hash}`);
}

function sendMessage() {
  chat.sendText(draft.value);
  draft.value = "";
  void nextTick(resizeMessageInput);
}

async function endChat() {
  await chat.endSession("user_closed");
}

async function submitEvaluation() {
  const submitted = await chat.submitEvaluation(
    rating.value,
    evaluationComment.value,
  );
  if (submitted) {
    evaluationComment.value = "";
  }
}

function handleComposerAction() {
  if (hasDraft.value) {
    sendMessage();
    return;
  }
  void endChat();
}

function openResourcePicker() {
  if (!chat.isOpen.value || chat.sessionClosed.value) return;
  resourceInput.value?.click();
}

function handleResourceSelected(event: Event) {
  const input = event.target as HTMLInputElement;
  selectedResourceName.value = input.files?.[0]?.name || "";
}

function resizeMessageInput() {
  const input = messageInput.value;
  if (!input) return;
  const maxHeight = 120;
  input.style.height = "auto";
  input.style.height = `${Math.min(input.scrollHeight, maxHeight)}px`;
}

async function handleMessageScroll() {
  const list = messageList.value;
  if (
    !list ||
    list.scrollTop > 24 ||
    chat.historyLoading.value ||
    !chat.historyHasMore.value
  ) {
    return;
  }
  const previousHeight = list.scrollHeight;
  const loaded = await chat.loadHistory(false);
  if (!loaded) return;
  await nextTick();
  list.scrollTop = list.scrollHeight - previousHeight;
}

async function scrollMessagesToBottom() {
  await nextTick();
  const list = messageList.value;
  if (!list) return;
  list.scrollTop = list.scrollHeight;
}

onMounted(() => {
  void connectChat();
  window.addEventListener("pagehide", handlePageHide);
});

watch(
  () => chat.messages.value[chat.messages.value.length - 1]?.messageNo,
  (messageNo, previousMessageNo) => {
    if (messageNo && messageNo !== previousMessageNo) {
      void scrollMessagesToBottom();
    }
  },
);

function handlePageHide() {
  void chat.endSession("page_leave", true);
}

onBeforeUnmount(() => {
  window.removeEventListener("pagehide", handlePageHide);
  void chat.endSession("page_leave", true);
});
</script>

<template>
  <input
    ref="resourceInput"
    class="resource-input"
    type="file"
    @change="handleResourceSelected"
  />

  <main class="chat-page" :class="{ 'chat-page--desktop': showDesktopFrame }">
    <section class="chat-shell" aria-label="chat conversation">
      <div v-if="chat.queueStatus.value" class="queue-message">
        {{ chat.queueStatus.value }}
      </div>

      <div ref="messageList" class="message-list" @scroll="handleMessageScroll">
        <div class="welcome-message">
          <strong>您好</strong>
          <span>请描述您遇到的问题，客服会在这里接收并回复。</span>
        </div>

        <article
          v-for="message in chat.messages.value"
          :key="message.messageNo"
          class="message-row"
          :class="{
            sent: messageDirection(message) === 'sent',
            received: messageDirection(message) === 'received',
            system: messageDirection(message) === 'system',
          }"
        >
          <div class="bubble">
            <span>{{ messageSenderName(message) }}</span>
            <p>{{ message.content }}</p>
          </div>
        </article>
      </div>

      <p v-if="chat.error.value || authError" class="error-line">
        {{ chat.error.value || authError }}
      </p>

      <section v-if="showEvaluationPanel" class="evaluation-panel">
        <div class="rating-row">
          <button
            v-for="value in 5"
            :key="value"
            class="rating-button"
            :class="{ active: value <= rating }"
            type="button"
            :disabled="
              !chat.isOpen.value ||
              chat.evaluationSubmitting.value ||
              chat.evaluationSubmitted.value
            "
            @click="rating = value"
          >
            ★
          </button>
        </div>
        <textarea
          v-model="evaluationComment"
          class="evaluation-input"
          :disabled="
            !chat.isOpen.value ||
            chat.evaluationSubmitting.value ||
            chat.evaluationSubmitted.value
          "
          placeholder="补充评价"
          rows="2"
        />
        <button
          class="evaluation-submit"
          type="button"
          :disabled="!canSubmitEvaluation"
          @click="submitEvaluation"
        >
          {{ chat.evaluationSubmitted.value ? "已评价" : "提交评价" }}
        </button>
      </section>

      <form class="composer" @submit.prevent="handleComposerAction">
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
            ref="messageInput"
            v-model="draft"
            :disabled="!chat.isOpen.value || chat.sessionClosed.value"
            :placeholder="
              chat.sessionClosed.value ? '本次会话已结束' : '输入消息'
            "
            rows="1"
            @input="resizeMessageInput"
          />
          <span v-if="selectedResourceName" class="resource-name">
            {{ selectedResourceName }}
          </span>
        </div>
        <button
          class="send-button"
          :disabled="!chat.isOpen.value || chat.sessionClosed.value"
          type="submit"
        >
          {{ composerActionLabel }}
        </button>
      </form>
    </section>
  </main>
</template>
