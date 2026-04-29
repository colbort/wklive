<script setup lang="ts">
import { useRouter } from 'vue-router'

defineProps<{
  title: string
  rightText?: string
  narrow?: boolean
}>()

const router = useRouter()
</script>

<template>
  <section class="asset-flow-page" :class="{ 'asset-flow-page--narrow': narrow }">
    <header class="asset-flow-header" :class="{ 'asset-flow-header--plain-right': !rightText }">
      <button type="button" class="asset-flow-back" aria-label="返回" @click="router.back()">‹</button>
      <h1>{{ title }}</h1>
      <button v-if="rightText" type="button" class="asset-flow-link">{{ rightText }}</button>
      <span v-else />
    </header>

    <main class="asset-flow-body">
      <slot />
    </main>
  </section>
</template>

<style scoped>
.asset-flow-page {
  min-height: 100dvh;
  padding: 12px 20px 36px;
  background: #0b0c15;
  color: #f8f8fb;
}

button {
  border: 0;
  background: transparent;
  color: inherit;
  font: inherit;
}

.asset-flow-header {
  display: grid;
  grid-template-columns: 40px 1fr 84px;
  align-items: center;
  min-height: 40px;
}

.asset-flow-header--plain-right {
  grid-template-columns: 40px 1fr 40px;
}

.asset-flow-back {
  display: grid;
  width: 40px;
  height: 40px;
  place-items: center;
  border-radius: 50%;
  background: #242633;
  font-size: 36px;
  line-height: 0.6;
}

.asset-flow-header h1 {
  margin: 0;
  font-size: 20px;
  text-align: center;
}

.asset-flow-link {
  color: #fff;
  font-size: 14px;
  font-weight: 700;
  text-align: right;
}

.asset-flow-body {
  padding-top: 24px;
}

@media (min-width: 768px) {
  .asset-flow-page--narrow {
    max-width: 720px;
    min-height: calc(100dvh - 76px);
    margin: 0 auto;
  }
}
</style>
