<script setup lang="ts">
import { computed } from 'vue'

import { apiGetCoreOptions } from '@/api/core'
import { optionText, useOptions } from '@/composables/useOptions'
import { useI18n } from '@/i18n'
import type { MarketTenantProduct } from '@/types/market'
import type { TradeSymbol } from '@/types/trade'

type SubmitSide = 'buy' | 'sell'

const props = defineProps<{
  selectedProduct: MarketTenantProduct | null
  orderMode: 'market' | 'limit'
  selectedTradeSymbol: TradeSymbol | null
  tradeSymbolLoading: boolean
  isLoggedIn: boolean
  tradeAvailable: boolean
  tradePrice: string
  tradeQty: string
  tradePercent: number
  referencePrice: string | number
  settleAsset: string
  availableBalance: string
  longPositionQty: string
  shortPositionQty: string
  tradeMessage: string
  tradeError: string
  submittingSide: SubmitSide | null
}>()

const emit = defineEmits<{
  (e: 'update:orderMode', value: 'market' | 'limit'): void
  (e: 'update:tradePrice', value: string): void
  (e: 'update:tradeQty', value: string): void
  (e: 'update:tradePercent', value: number): void
  (e: 'submit-order', side: SubmitSide): void
}>()

const { t } = useI18n()
const tradeOptions = useOptions(apiGetCoreOptions)
const percentSteps = [0, 25, 50, 75, 100]

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
const baseAsset = computed(
  () =>
    props.selectedTradeSymbol?.baseAsset ||
    props.selectedProduct?.baseCoin ||
    props.selectedProduct?.symbol ||
    '',
)
const quoteAsset = computed(
  () =>
    props.selectedTradeSymbol?.quoteAsset || props.selectedProduct?.quoteCoin || props.settleAsset,
)
const submitDisabled = computed(
  () =>
    !props.isLoggedIn ||
    !props.tradeAvailable ||
    props.tradeSymbolLoading ||
    Boolean(props.submittingSide),
)
const unavailableText = computed(() => {
  if (!props.isLoggedIn) return ''
  if (props.tradeSymbolLoading) return t('trade.configLoading')
  if (!props.tradeAvailable) return t('trade.unavailable')
  return ''
})
const conversionText = computed(() => {
  const price = Number(props.referencePrice)
  if (baseAsset.value && quoteAsset.value && Number.isFinite(price) && price > 0) {
    return `1 ${baseAsset.value} = ${props.referencePrice} ${quoteAsset.value}`
  }
  return `1 ${baseAsset.value || '--'} = -- ${quoteAsset.value || '--'}`
})

function orderModeFromCode(code: string): 'market' | 'limit' {
  return code === 'ORDER_TYPE_LIMIT' ? 'limit' : 'market'
}

function inputValue(event: Event) {
  return (event.target as HTMLInputElement).value
}

function updateTradePercent(value: number) {
  emit('update:tradePercent', value)
}

function handlePercentBarPointer(event: PointerEvent) {
  const rect = (event.currentTarget as HTMLElement).getBoundingClientRect()
  if (!rect.width) return
  const ratio = Math.min(Math.max((event.clientX - rect.left) / rect.width, 0), 1)
  updateTradePercent(Math.round((ratio * 100) / 25) * 25)
}
</script>

<template>
  <section class="spot-trade-panel">
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

    <label v-if="orderMode === 'limit'" class="trade-input">
      <input
        :value="tradePrice"
        inputmode="decimal"
        :placeholder="`${t('trade.price')}(${quoteAsset}) / ${selectedTradeSymbol?.priceTick || '--'}`"
        @input="emit('update:tradePrice', inputValue($event))"
      >
    </label>
    <label class="trade-input">
      <input
        :value="tradeQty"
        inputmode="decimal"
        :placeholder="`${t('trade.qty')}(${baseAsset}) / ${selectedTradeSymbol?.minQty || '--'}`"
        @input="emit('update:tradeQty', inputValue($event))"
      >
    </label>

    <div
      class="percent-bar"
      :style="{ '--progress': `${tradePercent}%` }"
      @pointerdown="handlePercentBarPointer"
    >
      <button
        v-for="value in percentSteps"
        :key="value"
        type="button"
        :aria-label="`${value}%`"
        @pointerdown.stop="updateTradePercent(value)"
        @click.stop="updateTradePercent(value)"
      />
    </div>
    <div class="percent-labels">
      <button
        v-for="value in percentSteps"
        :key="value"
        type="button"
        :class="{ active: tradePercent === value }"
        @click="updateTradePercent(value)"
      >
        {{ value }}%
      </button>
    </div>

    <div class="account-lines">
      <span>{{ t('trade.available') }}</span><strong>{{ availableBalance }} {{ quoteAsset }}</strong>
      <span>{{ t('trade.conversion') }}</span><strong>{{ conversionText }}</strong> <span>{{ t('trade.canBuy') }}</span><strong>{{ longPositionQty }} {{ baseAsset }}</strong>
    </div>
    <button
      class="wide-action wide-action--buy"
      type="button"
      :disabled="submitDisabled"
      @click="emit('submit-order', 'buy')"
    >
      {{ submittingSide === 'buy' ? t('common.submitting') : t('trade.buy') }}
    </button>

    <div class="account-lines">
      <span>{{ t('trade.canSell') }}</span><strong>{{ shortPositionQty }} {{ baseAsset }}</strong>
    </div>
    <button
      class="wide-action wide-action--sell"
      type="button"
      :disabled="submitDisabled"
      @click="emit('submit-order', 'sell')"
    >
      {{ submittingSide === 'sell' ? t('common.submitting') : t('trade.sell') }}
    </button>

    <p v-if="tradeError || unavailableText" class="order-message order-message--error">
      {{ tradeError || unavailableText }}
    </p>
    <p v-else-if="tradeMessage" class="order-message">
      {{ tradeMessage }}
    </p>
  </section>
</template>

<style scoped>
.mode-switch {
  display: grid;
  grid-template-columns: 1fr 1fr;
  min-height: 42px;
  margin-bottom: 14px;
  overflow: hidden;
  border-radius: 999px;
  background: var(--control-bg);
}

button,
input {
  border: 0;
  background: transparent;
  color: inherit;
  font: inherit;
}

.mode-switch button {
  color: var(--muted);
  font-size: 0.8rem;
  font-weight: 600;
}

.mode-switch button.active {
  border-radius: 999px;
  background: var(--accent);
  color: var(--text);
}

.trade-input {
  display: flex;
  min-height: 48px;
  margin-bottom: 12px;
  padding: 0 12px;
  border-radius: 12px;
  background: var(--control-bg);
  align-items: center;
}

.trade-input input {
  width: 100%;
  min-width: 0;
  outline: 0;
}

.percent-bar {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  height: 8px;
  margin-bottom: 6px;
  border-radius: 999px;
  background: linear-gradient(
    90deg,
    var(--accent) 0 var(--progress),
    var(--panel-bg-alt) var(--progress)
  );
}

.percent-bar button {
  height: 28px;
  margin-top: -10px;
  cursor: pointer;
}

.percent-labels {
  display: flex;
  margin-bottom: 14px;
  color: var(--muted);
  font-size: 0.65rem;
  justify-content: space-between;
}

.percent-labels button.active {
  color: var(--text);
}

.account-lines {
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 7px 12px;
  margin-bottom: 14px;
  color: var(--muted);
  font-size: 0.65rem;
}

.account-lines strong {
  color: var(--text);
  font-weight: 500;
}

.wide-action {
  width: 100%;
  min-height: 46px;
  margin-bottom: 18px;
  border-radius: 12px;
  color: var(--text);
  font-size: 0.75rem;
  font-weight: 700;
}

.wide-action--buy {
  background: var(--success);
}
.wide-action--sell {
  background: var(--danger-strong);
}
.wide-action:disabled {
  opacity: 0.52;
}

.order-message {
  margin: -4px 0 12px;
  color: var(--success);
  font-size: 0.65rem;
}

.order-message--error {
  color: var(--danger);
}
</style>
