<script setup lang="ts">
import AssetCoinIcon from '@/components/assets/AssetCoinIcon.vue'
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

const chainLabels: Record<number, string> = {
  1: 'BTC',
  2: 'ETH',
  3: 'TRX',
  4: 'BSC',
  5: 'SOL',
  6: 'POLYGON',
  20: 'TRC20',
  21: 'ERC20',
  22: 'BEP20',
}

function closeSheet() {
  emit('update:modelValue', false)
}

function chainLabel(config: AssetCoinConfig) {
  if (config.chainCode) return chainLabels[config.chainCode] || String(config.chainCode)
  return config.coin === 'USDT' ? 'TRC20' : ''
}

function isSelected(config: AssetCoinConfig) {
  if (!props.selectedConfig) return false
  if (props.selectedConfig.id && config.id) return props.selectedConfig.id === config.id
  return props.selectedConfig.coin === config.coin && props.selectedConfig.chainCode === config.chainCode
}

function selectConfig(config: AssetCoinConfig) {
  emit('select', config)
  closeSheet()
}
</script>

<template>
  <Teleport to="body">
    <Transition name="asset-coin-sheet">
      <div v-if="modelValue" class="asset-coin-sheet" @click.self="closeSheet">
        <section class="asset-coin-sheet__panel" role="dialog" aria-modal="true">
          <span class="asset-coin-sheet__handle" />
          <button type="button" class="asset-coin-sheet__close" aria-label="关闭" @click="closeSheet">×</button>
          <h2>{{ title || '币种选择' }}</h2>

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
            <p v-if="!props.configs.length" class="asset-coin-sheet__empty">暂无可选币种</p>
          </div>
        </section>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
button {
  border: 0;
  background: transparent;
  color: inherit;
  font: inherit;
}

.asset-coin-sheet {
  position: fixed;
  inset: 0;
  z-index: 1000;
  display: flex;
  align-items: flex-end;
  background: rgba(0, 0, 0, 0.58);
  backdrop-filter: blur(8px);
}

.asset-coin-sheet__panel {
  position: relative;
  width: 100%;
  max-height: 68dvh;
  overflow: hidden auto;
  border-radius: 24px 24px 0 0;
  background: #24252d;
  padding: 38px 0 16px;
  color: #fff;
}

.asset-coin-sheet__handle {
  position: absolute;
  top: 14px;
  left: 50%;
  width: 34px;
  height: 4px;
  border-radius: 999px;
  background: #a8a9af;
  transform: translateX(-50%);
}

.asset-coin-sheet__close {
  position: absolute;
  top: 22px;
  right: 28px;
  width: 32px;
  height: 32px;
  color: #fff;
  font-size: 34px;
  line-height: 28px;
}

.asset-coin-sheet h2 {
  margin: 0 0 28px;
  text-align: center;
  font-size: 18px;
  font-weight: 800;
}

.asset-coin-sheet__list {
  display: grid;
}

.asset-coin-sheet__item {
  display: grid;
  grid-template-columns: 28px auto auto 1fr;
  align-items: center;
  gap: 10px;
  min-height: 52px;
  padding: 0 28px;
  text-align: left;
}

.asset-coin-sheet__item.is-active {
  background: #343640;
}

.asset-coin-sheet__item strong {
  font-size: 16px;
  font-weight: 800;
}

.asset-coin-sheet__item small {
  padding: 3px 8px;
  border-radius: 999px;
  background: #4a4c58;
  color: #fff;
  font-size: 11px;
  font-weight: 800;
}

.asset-coin-sheet__check {
  justify-self: end;
  color: #02b904;
  font-size: 20px;
  font-weight: 900;
}

.asset-coin-sheet__empty {
  margin: 12px 28px 22px;
  color: #9b9da6;
  font-size: 14px;
  text-align: center;
}

.asset-coin-sheet-enter-active,
.asset-coin-sheet-leave-active {
  transition: opacity 0.18s ease;
}

.asset-coin-sheet-enter-active .asset-coin-sheet__panel,
.asset-coin-sheet-leave-active .asset-coin-sheet__panel {
  transition: transform 0.18s ease;
}

.asset-coin-sheet-enter-from,
.asset-coin-sheet-leave-to {
  opacity: 0;
}

.asset-coin-sheet-enter-from .asset-coin-sheet__panel,
.asset-coin-sheet-leave-to .asset-coin-sheet__panel {
  transform: translateY(18px);
}
</style>
