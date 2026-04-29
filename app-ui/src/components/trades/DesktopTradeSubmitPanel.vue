<script setup lang="ts">
import { computed } from 'vue'

import { apiGetTradeOptions } from '@/api/trade'
import { optionText, useOptions } from '@/composables/useOptions'
import type { ItickTenantProduct } from '@/types/itick'

defineProps<{
  selectedProduct: ItickTenantProduct | null
  orderMode: 'market' | 'limit'
}>()

const emit = defineEmits<{
  (e: 'update:orderMode', value: 'market' | 'limit'): void
}>()

const tradeOptions = useOptions(apiGetTradeOptions)
const orderTypeOptions = computed(() => {
  const options = tradeOptions.getGroup('orderType').filter((option) => {
    return ['ORDER_TYPE_MARKET', 'ORDER_TYPE_LIMIT'].includes(option.code)
  })

  return options.length
    ? options
    : [
        { value: 2, code: 'ORDER_TYPE_MARKET' },
        { value: 1, code: 'ORDER_TYPE_LIMIT' },
      ]
})
const marginModeOptions = computed(() => {
  const options = tradeOptions.getGroup('marginMode').filter((option) => option.value > 0)
  return options.length ? options : [{ value: 1, code: 'MARGIN_MODE_CROSS' }]
})

function orderModeFromCode(code: string): 'market' | 'limit' {
  return code === 'ORDER_TYPE_LIMIT' ? 'limit' : 'market'
}
</script>

<template>
  <aside class="desktop-order-panel">
    <div class="mode-switch">
      <button
        v-for="option in orderTypeOptions"
        :key="option.value"
        type="button"
        :class="{ active: orderMode === orderModeFromCode(option.code) }"
        @click="emit('update:orderMode', orderModeFromCode(option.code))"
      >
        {{ optionText(option) }}
      </button>
    </div>

    <div class="desktop-order-panel__grid">
      <div>
        <label>保证金模式</label>
        <button type="button">{{ optionText(marginModeOptions[0]) }}</button>
      </div>
      <div>
        <label>杠杆</label>
        <button type="button">50X</button>
      </div>
    </div>

    <label class="desktop-order-panel__label">
      数量 ({{ selectedProduct?.baseCoin || selectedProduct?.symbol || 'BTC' }})
    </label>
    <div class="trade-input">数量</div>
    <div class="percent-bar"><i /></div>
    <div class="percent-labels"><span>0%</span><span>25%</span><span>50%</span><span>75%</span><span>100%</span></div>

    <div class="account-lines">
      <span>可用</span><strong>0 USDT</strong>
      <span>换算</span><strong>1 手 = 1 BTC</strong>
    </div>

    <label class="checkbox-line"><i />止盈/止损</label>
    <div class="account-lines">
      <span>可开多</span><strong>0 手</strong>
      <span>保证金</span><strong>0 USDT</strong>
    </div>
    <button class="wide-action wide-action--buy" type="button">买涨</button>

    <label class="checkbox-line"><i />止盈/止损</label>
    <div class="account-lines">
      <span>可开空</span><strong>0 手</strong>
      <span>保证金</span><strong>0 USDT</strong>
    </div>
    <button class="wide-action wide-action--sell" type="button">买跌</button>
  </aside>
</template>

<style scoped>
.desktop-order-panel {
  width: 100%;
  min-height: 0;
  overflow-y: auto;
  overscroll-behavior: contain;
  padding: 14px 14px 18px;
}

.desktop-order-panel__grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
  margin-bottom: 14px;
}

.desktop-order-panel__grid label,
.desktop-order-panel__label {
  display: block;
  margin-bottom: 6px;
  font-size: 13px;
  font-weight: 600;
}

.desktop-order-panel__grid button {
  width: 100%;
  min-height: 52px;
  border: 0;
  border-radius: 12px;
  background: #242631;
  color: #fff;
  font: inherit;
  font-size: 15px;
  text-align: left;
  padding: 0 14px;
}

.mode-switch {
  display: grid;
  grid-template-columns: 1fr 1fr;
  min-height: 58px;
  margin-bottom: 18px;
  overflow: hidden;
  border-radius: 999px;
  background: #242631;
}

.mode-switch button {
  border: 0;
  background: transparent;
  color: #8f929d;
  font: inherit;
  font-size: 17px;
}

.mode-switch button.active {
  border-radius: 999px;
  background: #02b904;
  color: #fff;
}

.trade-input {
  display: flex;
  align-items: center;
  min-height: 50px;
  margin-bottom: 14px;
  padding: 0 14px;
  border-radius: 12px;
  background: #242631;
  color: #8f929d;
  font-size: 14px;
}

.percent-bar {
  height: 18px;
  margin-bottom: 10px;
  border-radius: 999px;
  background: linear-gradient(90deg, #1c1f2a 0 24%, transparent 24% 25%, #1c1f2a 25% 49%, transparent 49% 50%, #1c1f2a 50% 74%, transparent 74% 75%, #1c1f2a 75%);
}

.percent-bar i {
  display: block;
  width: 18px;
  height: 18px;
  border-radius: 999px;
  background: #02b904;
}

.percent-labels {
  display: flex;
  justify-content: space-between;
  margin-bottom: 18px;
  color: #8f929d;
  font-size: 12px;
}

.account-lines {
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 8px 14px;
  margin-bottom: 12px;
  color: #8f929d;
  font-size: 12px;
}

.account-lines strong {
  color: #fff;
  font-weight: 500;
}

.checkbox-line {
  display: flex;
  align-items: center;
  gap: 10px;
  margin: 8px 0 12px;
  font-size: 14px;
}

.checkbox-line i {
  width: 20px;
  height: 20px;
  border: 1px solid #f6f7fb;
  border-radius: 4px;
}

.wide-action {
  width: 100%;
  min-height: 48px;
  margin-bottom: 14px;
  border: 0;
  border-radius: 12px;
  color: #fff;
  font: inherit;
  font-size: 15px;
}

.wide-action--buy {
  background: #02b904;
}

.wide-action--sell {
  background: #ff5a49;
}
</style>
