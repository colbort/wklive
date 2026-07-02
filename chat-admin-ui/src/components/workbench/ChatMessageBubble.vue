<script setup lang="ts">
import type { ChatMessage } from "@/types/chat";

defineProps<{
  message: ChatMessage;
  direction: "sent" | "received" | "system";
  senderName: string;
}>();

function fileSizeText(size = 0) {
  if (!size) return "";
  if (size < 1024) return `${size} B`;
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`;
  return `${(size / 1024 / 1024).toFixed(1)} MB`;
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
        v-if="message.messageType === 2 && message.url"
        class="bubble-image"
        :src="message.url"
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
