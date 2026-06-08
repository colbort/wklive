<script setup lang="ts">
import type { AssetCoinConfig } from '@/types/asset'

const props = defineProps<{
  coin: string
  config?: AssetCoinConfig
}>()

function coinIconText() {
  return (props.config?.iconText || props.config?.symbol || props.coin || '?')
    .slice(0, 3)
    .toUpperCase()
}
</script>

<template>
  <span class="asset-coin-icon" :style="{ backgroundColor: config?.iconBgColor || '#16ad77' }">
    <img v-if="config?.iconUrl" :src="config.iconUrl" :alt="coin" />
    <span v-else>{{ coinIconText() }}</span>
  </span>
</template>

<style scoped>
.asset-coin-icon {
  display: grid;
  width: 28px;
  height: 28px;
  place-items: center;
  overflow: hidden;
  border-radius: 50%;
  color: var(--text);
  font-size: 0.45rem;
  font-weight: 800;
}

.asset-coin-icon img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
</style>
