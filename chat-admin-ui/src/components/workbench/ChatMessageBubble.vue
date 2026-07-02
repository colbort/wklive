<script setup lang="ts">
import { onBeforeUnmount, ref, watch } from "vue";
import type { ChatMessage } from "@/types/chat";

const props = defineProps<{
  message: ChatMessage;
  direction: "sent" | "received" | "system";
  senderName: string;
  resolveUrl?: (url: string) => Promise<string> | string;
}>();

const imageSrc = ref("");
let objectUrl = "";

watch(
  () => [props.message.messageType, props.message.url, props.resolveUrl] as const,
  () => {
    void loadImageSrc();
  },
  { immediate: true },
);

onBeforeUnmount(clearObjectUrl);

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
  <div
    class="message-row"
    :class="direction"
  >
    <div class="bubble">
      <span>{{ senderName }}</span>
      <img
        v-if="message.messageType === 2 && imageSrc"
        class="bubble-image"
        :src="imageSrc"
        :alt="message.fileName || message.content || 'image'"
      />
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
      <p v-else>{{ message.content }}</p>
    </div>
  </div>
</template>
