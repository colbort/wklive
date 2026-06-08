<script setup lang="ts">
import AssetCoinIcon from '@/components/assets/AssetCoinIcon.vue'
import type { AssetCoinConfig } from '@/types/asset'

defineProps<{
  coin: string
  config?: AssetCoinConfig
  chain?: string
}>()

const emit = defineEmits<{
  click: []
}>()
</script>

<template>
  <button type="button" class="asset-picker" @click="emit('click')">
    <AssetCoinIcon :coin="coin" :config="config" />
    <strong>{{ coin }}</strong>
    <small v-if="chain">{{ chain }}</small>
    <span class="asset-picker__arrow">›</span>
  </button>
</template>

<style scoped>
button {
  border: 0;
  background: transparent;
  color: inherit;
  font: inherit;
}

.asset-picker {
  display: grid;
  grid-template-columns: 28px minmax(0, auto) minmax(0, auto) minmax(18px, 1fr);
  align-items: center;
  gap: 8px;
  width: 100%;
  min-height: 56px;
  padding: 0 16px;
  border-radius: 14px;
  background: var(--field-bg);
  text-align: left;
}

.asset-picker strong {
  min-width: 0;
  overflow: hidden;
  font-size: 0.8rem;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.asset-picker small {
  min-width: 0;
  overflow: hidden;
  padding: 3px 7px;
  border-radius: 999px;
  background: var(--selection-bg);
  color: var(--text);
  font-size: 0.5rem;
  font-weight: 800;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.asset-picker__arrow {
  justify-self: end;
  color: var(--muted);
  font-size: 1.2rem;
}

@media (max-width: 390px) {
  .asset-picker {
    grid-template-columns: 26px minmax(0, auto) minmax(0, auto) minmax(16px, 1fr);
    gap: 7px;
    min-height: 52px;
    padding: 0 12px;
    border-radius: 13px;
  }

  .asset-picker strong {
    font-size: 0.75rem;
  }
}
</style>
