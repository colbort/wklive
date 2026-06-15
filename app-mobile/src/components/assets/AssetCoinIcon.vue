<script setup lang="ts">
import { computed } from 'vue'

import { useSystemStore } from '@/stores/system'
import type { AssetCoinConfig } from '@/types/asset'
import { resolveSystemAssetUrl } from '@/utils/assetUrl'

const props = defineProps<{
  coin: string
  config?: AssetCoinConfig
}>()

const systemStore = useSystemStore()
const iconUrl = computed(() =>
  resolveSystemAssetUrl(systemStore.systemCore.assetUrl, props.config?.iconUrl),
)

function coinIconText() {
  return (props.config?.iconText || props.config?.symbol || props.coin || '?')
    .slice(0, 3)
    .toUpperCase()
}
</script>

<template>
  <span class="asset-coin-icon" :style="{ backgroundColor: config?.iconBgColor || 'var(--coin-fallback-bg)' }">
    <img v-if="iconUrl" :src="iconUrl" :alt="coin">
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
