<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from "vue";
import type { ChatMessage } from "@/types/chat";

const props = defineProps<{
  message: ChatMessage;
  direction: "sent" | "received" | "system";
  senderName: string;
  resolveUrl?: (url: string) => Promise<string> | string;
}>();

const emit = defineEmits<{
  recall: [message: ChatMessage];
  delete: [message: ChatMessage];
}>();

const rootRef = ref<HTMLElement | null>(null);
const imageSrc = ref("");
const menuOpen = ref(false);
const menuLeft = ref(0);
const menuTop = ref(0);
let objectUrl = "";
const sendingStatus = 1;
const recalledStatus = 6;
const deletedStatus = 7;
const recallWindowMs = 3 * 60 * 1000;

function canDelete() {
  return (
    props.direction === "sent" &&
    props.message.status !== sendingStatus &&
    props.message.status !== recalledStatus &&
    props.message.status !== deletedStatus
  );
}

function canRecall() {
  const createTime = Number(
    props.message.createTime || props.message.updateTime || 0,
  );
  return (
    canDelete() && createTime > 0 && Date.now() - createTime <= recallWindowMs
  );
}

function hasMenu() {
  return canDelete() || canRecall();
}

watch(
  () =>
    [props.message.messageType, props.message.url, props.resolveUrl] as const,
  () => {
    void loadImageSrc();
  },
  { immediate: true },
);

onMounted(() => {
  window.addEventListener("click", closeMenu);
  window.addEventListener("scroll", closeMenu, true);
  window.addEventListener("keydown", handleKeydown);
  window.addEventListener("contextmenu", handleWindowContextMenu);
});

onBeforeUnmount(clearObjectUrl);
onBeforeUnmount(() => {
  window.removeEventListener("click", closeMenu);
  window.removeEventListener("scroll", closeMenu, true);
  window.removeEventListener("keydown", handleKeydown);
  window.removeEventListener("contextmenu", handleWindowContextMenu);
});

function openMenu(event: MouseEvent) {
  if (!hasMenu()) return;
  event.preventDefault();
  menuLeft.value = event.clientX;
  menuTop.value = event.clientY;
  menuOpen.value = true;
}

function closeMenu() {
  menuOpen.value = false;
}

function handleKeydown(event: KeyboardEvent) {
  if (event.key === "Escape") closeMenu();
}

function handleWindowContextMenu(event: MouseEvent) {
  if (!rootRef.value?.contains(event.target as Node)) {
    closeMenu();
  }
}

function emitRecall() {
  if (!canRecall()) return;
  emit("recall", props.message);
  closeMenu();
}

function emitDelete() {
  if (!canDelete()) return;
  emit("delete", props.message);
  closeMenu();
}

function fileSizeText(size = 0) {
  if (!size) return "";
  if (size < 1024) return `${size} B`;
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`;
  return `${(size / 1024 / 1024).toFixed(1)} MB`;
}

async function loadImageSrc() {
  clearObjectUrl();
  imageSrc.value = "";
  if (props.message.messageType !== 2 || !props.message.url) return;
  try {
    const resolved = props.resolveUrl
      ? await props.resolveUrl(props.message.url)
      : props.message.url;
    imageSrc.value = resolved;
    if (resolved.startsWith("blob:")) {
      objectUrl = resolved;
    }
  } catch {
    imageSrc.value = props.message.url;
  }
}

function clearObjectUrl() {
  if (objectUrl) {
    URL.revokeObjectURL(objectUrl);
    objectUrl = "";
  }
}
</script>

<template>
  <article
    v-if="message.status !== recalledStatus"
    ref="rootRef"
    class="message-row"
    :class="direction"
    @contextmenu="openMenu"
  >
    <div class="bubble">
      <span>{{ senderName }}</span>
      <p v-if="message.status === deletedStatus">
        消息已删除
      </p>
      <img
        v-else-if="message.messageType === 2 && imageSrc"
        class="bubble-image"
        :src="imageSrc"
        :alt="message.fileName || message.content || 'image'"
      >
      <video
        v-else-if="message.messageType === 4 && message.url"
        class="bubble-video"
        :src="message.url"
        controls
        playsinline
      />
      <audio
        v-else-if="message.messageType === 5 && message.url"
        class="bubble-audio"
        :src="message.url"
        controls
      />
      <a
        v-else-if="message.messageType === 3 && message.url"
        class="bubble-file"
        :href="message.url"
        target="_blank"
        rel="noreferrer"
      >
        <strong>{{ message.fileName || message.content || "文件" }}</strong>
        <small>{{ message.mimeType }} {{ fileSizeText(message.fileSize) }}</small>
      </a>
      <p v-else>
        {{ message.content }}
      </p>
    </div>
    <div
      v-if="menuOpen && hasMenu()"
      class="message-context-menu"
      :style="{ left: `${menuLeft}px`, top: `${menuTop}px` }"
      @click.stop
    >
      <button
        v-if="canRecall()"
        type="button"
        @click="emitRecall"
      >
        撤回
      </button>
      <button
        v-if="canDelete()"
        type="button"
        @click="emitDelete"
      >
        删除
      </button>
    </div>
  </article>
</template>
