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
  grid-template-columns: 28px auto auto 1fr;
  align-items: center;
  gap: 8px;
  width: 100%;
  min-height: 56px;
  padding: 0 16px;
  border-radius: 14px;
  background: #292b36;
  text-align: left;
}

.asset-picker strong {
  font-size: 16px;
}

.asset-picker small {
  padding: 3px 7px;
  border-radius: 999px;
  background: #4a4c58;
  color: #fff;
  font-size: 10px;
  font-weight: 800;
}

.asset-picker__arrow {
  justify-self: end;
  color: #989aa3;
  font-size: 24px;
}
</style>
