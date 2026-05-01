<script setup lang='ts'>
import { computed } from 'vue'

import { apiGetTradeOptions } from '@/api/trade'
import { optionText, useOptions } from '@/composables/useOptions'
import type { ItickTenantProduct } from '@/types/itick'

defineProps<{
  selectedProduct: ItickTenantProduct | null
  tradeKind: 'stock' | 'option' | 'forex' | 'commodity' | 'crypto'
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
  <section v-if="tradeKind === 'stock'" class="stock-panel">
    <div class="stock-alert">
      <span>!</span>
      <strong>该品种休市中，期间暂停交易</strong>
    </div>

    <div class="inner-tabs">
      <button class="active" type="button">普通交易</button>
      <button type="button">融资融券</button>
      <button type="button">盘前</button>
    </div>

    <div class="mode-switch compact">
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

    <div class="trade-input">数量</div>
    <div class="percent-bar"><i /></div>
    <div class="percent-labels"><span>0%</span><span>25%</span><span>50%</span><span>75%</span><span>100%</span></div>
    <button class="wide-action" type="button">登录/注册</button>
  </section>

  <section v-else-if="tradeKind === 'option'" class="option-panel">
    <div class="mini-chart">
      <svg viewBox="0 0 320 88" aria-label="走势">
        <path d="M0 50 C28 48 26 40 54 42 C88 47 82 28 119 25 C152 20 143 32 176 28 C218 26 199 64 231 63 C266 62 252 49 320 53" />
      </svg>
    </div>

    <h3>时间</h3>
    <div class="duration-grid">
      <button class="active" type="button"><strong>30S</strong><span>30%</span></button>
      <button type="button"><strong>60S</strong><span>40%</span></button>
      <button type="button"><strong>90S</strong><span>50%</span></button>
      <button type="button"><strong>120S</strong><span>60%</span></button>
      <button type="button"><strong>180S</strong><span>70%</span></button>
      <button type="button"><strong>360S</strong><span>80%</span></button>
    </div>

    <h3>投资额</h3>
    <div class="trade-input split"><span>>=100</span><strong>USD</strong></div>
    <div class="percent-bar"><i /></div>
    <div class="percent-labels"><span>0%</span><span>25%</span><span>50%</span><span>75%</span><span>100%</span></div>
    <div class="buyable"><span>可买</span><strong>0 USD</strong></div>
    <div class="dual-actions"><button type="button">登录</button><button type="button">注册</button></div>
  </section>

  <section v-else class="contract-panel">
    <div class="trade-form">
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

      <div class="form-row">
        <button type="button">{{ optionText(marginModeOptions[0]) }}⌄</button>
        <button type="button">1X⌄</button>
      </div>

      <div class="trade-input">
        数量({{ tradeKind === 'forex' ? selectedProduct?.symbol : selectedProduct?.baseCoin || selectedProduct?.symbol || 'BTC' }})
      </div>
      <div class="percent-bar"><i /></div>
      <div class="percent-labels"><span>0%</span><span>25%</span><span>50%</span><span>75%</span><span>100%</span></div>

      <div class="account-lines">
        <span>可用</span><strong>0 USD</strong>
        <span>换算</span><strong>{{ tradeKind === 'crypto' ? '1 手 = 1 BTC' : `1 手 = 1 ${selectedProduct?.symbol || ''}` }}</strong>
      </div>

      <label class="checkbox-line"><i />止盈/止损</label>
      <div class="account-lines">
        <span>可开多</span><strong>0 手</strong>
        <span>保证金</span><strong>0 USD</strong>
      </div>
      <button class="wide-action" type="button">登录</button>

      <label class="checkbox-line"><i />止盈/止损</label>
      <div class="account-lines">
        <span>可开空</span><strong>0 手</strong>
        <span>保证金</span><strong>0 USD</strong>
      </div>
      <button class="wide-action" type="button">注册</button>
    </div>
  </section>
</template>

<style scoped>
.mode-switch {
  display: grid;
  grid-template-columns: 1fr 1fr;
  min-height: 58px;
  margin-bottom: 18px;
  overflow: hidden;
  border-radius: 999px;
  background: #242631;
}

.mode-switch.compact {
  max-width: 260px;
}

.mode-switch button,
.form-row button,
.trade-input,
.wide-action,
.dual-actions button,
.inner-tabs button,
.duration-grid button {
  border: 0;
  background: transparent;
  color: inherit;
  font: inherit;
}

.mode-switch button {
  color: #8f929d;
  font-size: 17px;
}

.mode-switch button.active {
  border-radius: 999px;
  background: #02b904;
  color: #fff;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 14px;
  margin-bottom: 18px;
}

.form-row button,
.trade-input {
  min-height: 58px;
  padding: 0 18px;
  border-radius: 12px;
  background: #242631;
  color: #f6f7fb;
  text-align: left;
}

.trade-input {
  display: flex;
  align-items: center;
  margin-bottom: 18px;
  color: #8f929d;
  font-size: 17px;
}

.trade-input.split {
  justify-content: space-between;
}

.trade-input.split strong {
  color: #fff;
  font-weight: 500;
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
  margin-bottom: 24px;
  color: #8f929d;
  font-size: 14px;
}

.account-lines {
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 8px 14px;
  margin-bottom: 18px;
  color: #8f929d;
  font-size: 14px;
}

.account-lines strong {
  color: #fff;
  font-weight: 500;
}

.checkbox-line {
  display: flex;
  align-items: center;
  gap: 10px;
  margin: 8px 0 16px;
  font-size: 17px;
}

.checkbox-line i {
  width: 20px;
  height: 20px;
  border: 1px solid #f6f7fb;
  border-radius: 4px;
}

.wide-action,
.dual-actions button {
  min-height: 54px;
  border-radius: 12px;
  background: #181b25;
  color: #fff;
  font-size: 17px;
}

.wide-action {
  width: 100%;
  margin-bottom: 24px;
}

.dual-actions {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 18px;
  margin-bottom: 32px;
}

.stock-alert {
  display: flex;
  align-items: center;
  gap: 14px;
  margin: 0 -22px 22px;
  padding: 16px 28px;
  background: #282a34;
}

.stock-alert span {
  display: grid;
  place-items: center;
  width: 28px;
  height: 28px;
  border-radius: 999px;
  background: #ffa51f;
  color: #0b0c15;
}

.stock-alert strong {
  font-size: 15px;
  font-weight: 500;
}

.inner-tabs {
  display: flex;
  gap: 22px;
  margin-bottom: 22px;
}

.inner-tabs button {
  position: relative;
  color: #8f929d;
  font-size: 17px;
}

.inner-tabs button.active {
  color: #fff;
}

.inner-tabs button.active::after {
  position: absolute;
  right: 0;
  bottom: -12px;
  left: 0;
  height: 3px;
  border-radius: 999px;
  background: #02b904;
  content: '';
}

.option-panel h3 {
  margin: 24px 0 14px;
  font-size: 17px;
  font-weight: 500;
}

.mini-chart {
  margin-bottom: 22px;
  overflow: hidden;
  border-radius: 16px;
  background: radial-gradient(circle at 68% 20%, rgba(7, 201, 128, 0.26), transparent 36%), #10131d;
}

.mini-chart svg {
  width: 100%;
  height: 88px;
}

.mini-chart path {
  fill: none;
  stroke: #fff;
  stroke-width: 4;
  stroke-linecap: round;
}

.duration-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 14px;
}

.duration-grid button {
  display: grid;
  gap: 8px;
  min-height: 76px;
  padding: 14px 10px;
  border-radius: 14px;
  background: #242631;
  color: #8f929d;
  text-align: left;
}

.duration-grid button.active {
  background: #02b904;
  color: #fff;
}

.duration-grid strong {
  font-size: 19px;
  font-weight: 600;
}

.duration-grid span {
  font-size: 13px;
}

.buyable {
  display: grid;
  grid-template-columns: 1fr auto;
  margin-bottom: 22px;
  color: #8f929d;
  font-size: 14px;
}

.buyable strong {
  color: #fff;
  font-weight: 500;
}
</style>
