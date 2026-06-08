<script setup lang="ts">
import AssetCoinIcon from '@/components/assets/AssetCoinIcon.vue'
import BottomDrawer from '@/components/common/BottomDrawer.vue'
import { apiGetAssetOptions } from '@/api/asset'
import { useOptions } from '@/composables/useOptions'
import { useI18n } from '@/i18n'
import type { AssetCoinConfig } from '@/types/asset'

const props = defineProps<{
  modelValue: boolean
  title?: string
  configs: AssetCoinConfig[]
  selectedConfig?: AssetCoinConfig | null
  operationType?: number
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  select: [config: AssetCoinConfig]
}>()

const assetOptions = useOptions(apiGetAssetOptions)
const { t } = useI18n()

function closeSheet() {
  emit('update:modelValue', false)
}

function chainLabel(config: AssetCoinConfig) {
  if (config.chainCode) {
    const option = assetOptions
      .getGroup('chainCode')
      .find((item) => item.value === config.chainCode)
    return option ? formatChainCode(option.code) : String(config.chainCode)
  }
  return config.coin === 'USDT' ? 'TRC20' : ''
}

function formatChainCode(code: string) {
  return code.replace(/^CHAIN_CODE_/, '')
}

function isSelected(config: AssetCoinConfig) {
  if (!props.selectedConfig) return false
  if (props.selectedConfig.id && config.id) return props.selectedConfig.id === config.id
  return (
    props.selectedConfig.coin === config.coin && props.selectedConfig.chainCode === config.chainCode
  )
}

function selectConfig(config: AssetCoinConfig) {
  emit('select', config)
  closeSheet()
}
</script>

<template>
  <BottomDrawer
    :model-value="modelValue"
    :title="title || t('assetFlow.chooseCoin')"
    :close-label="t('common.close')"
    max-height="68dvh"
    @update:model-value="emit('update:modelValue', $event)"
  >
    <div class="asset-coin-sheet__list">
      <button
        v-for="config in props.configs"
        :key="`${config.id}-${config.coin}-${config.chainCode}`"
        type="button"
        class="asset-coin-sheet__item"
        :class="{ 'is-active': isSelected(config) }"
        @click="selectConfig(config)"
      >
        <AssetCoinIcon :coin="config.coin" :config="config" />
        <strong>{{ config.coin }}</strong>
        <small v-if="chainLabel(config)">{{ chainLabel(config) }}</small>
        <span v-if="isSelected(config)" class="asset-coin-sheet__check">✓</span>
      </button>
      <p v-if="!props.configs.length" class="asset-coin-sheet__empty">{{ t('assets.noCoins') }}</p>
    </div>
  </BottomDrawer>
</template>

<style scoped>
button {
  border: 0;
  background: transparent;
  color: inherit;
  font: inherit;
}

.asset-coin-sheet__list {
  display: grid;
  margin: 0 -22px;
}

.asset-coin-sheet__item {
  display: grid;
  grid-template-columns: 28px minmax(0, auto) minmax(0, auto) minmax(20px, 1fr);
  align-items: center;
  gap: 10px;
  min-height: 52px;
  padding: 0 28px;
  text-align: left;
}

.asset-coin-sheet__item.is-active {
  background: var(--field-bg-strong);
}

.asset-coin-sheet__item strong {
  min-width: 0;
  overflow: hidden;
  font-size: 16px;
  font-weight: 800;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.asset-coin-sheet__item small {
  min-width: 0;
  overflow: hidden;
  padding: 3px 8px;
  border-radius: 999px;
  background: #4a4c58;
  color: var(--text);
  font-size: 11px;
  font-weight: 800;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.asset-coin-sheet__check {
  justify-self: end;
  color: var(--accent);
  font-size: 20px;
  font-weight: 900;
}

@media (max-width: 390px) {
  .asset-coin-sheet__item {
    min-height: 48px;
    padding: 0 20px;
  }
}

.asset-coin-sheet__empty {
  margin: 12px 28px 22px;
  color: var(--muted);
  font-size: 14px;
  text-align: center;
}

</style>
