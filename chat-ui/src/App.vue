<script setup lang="ts">
import { computed, ref } from "vue";
import ChatPage from "@/pages/ChatPage.vue";
import ChatTestPage from "@/pages/ChatTestPage.vue";

const routePath = ref(window.location.pathname);
const routeParams = ref(new URLSearchParams(window.location.search));
const isChatRoute = computed(
  () =>
    routeParams.value.get("page") === "chat" ||
    routePath.value.replace(/\/$/, "").endsWith("/chat"),
);
</script>

<template>
  <ChatPage v-if="isChatRoute" />
  <ChatTestPage v-else />
</template>
