<script setup lang="ts">
import { computed } from 'vue'

import { apiGetTradeOptions } from '@/api/trade'
import { optionText, useOptions } from '@/composables/useOptions'
import type { ItickTenantProduct } from '@/types/itick'
import type {
  TradeSymbol,
  TradeSymbolContract,
  TradeSymbolLeverageConfig,
  TradeSymbolSpot,
} from '@/types/trade'

type SubmitSide = 'buy' | 'sell'
type TradeSymbolDetail = {
  symbol: TradeSymbol | null
  spot: TradeSymbolSpot | null
  contract: TradeSymbolContract | null
  leverageConfigs: TradeSymbolLeverageConfig[]
}

const props = defineProps<{
  selectedProduct: ItickTenantProduct | null
  orderMode: 'market' | 'limit'
  selectedTradeSymbol: TradeSymbol | null
  tradeSymbolDetail: TradeSymbolDetail | null
  tradeSymbolLoading: boolean
  isLoggedIn: boolean
  tradeAvailable: boolean
  tradePrice: string
  tradeQty: string
  tradePercent: number
  marginMode: number
  leverage: number
  maxLeverage: number
  leverageValues: number[]
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
  (e: 'update:marginMode', value: number): void
  (e: 'update:leverage', value: number): void
  (e: 'submit-order', side: SubmitSide): void
}>()

const MARKET_TYPE_SPOT = 1
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
const isSpotTrade = computed(() => props.selectedTradeSymbol?.marketType === MARKET_TYPE_SPOT)
const canSubmit = computed(
  () => props.isLoggedIn && props.tradeAvailable && !props.tradeSymbolLoading,
)
const submitDisabled = computed(() => !canSubmit.value || Boolean(props.submittingSide))
const baseAsset = computed(() => {
  return (
    props.selectedTradeSymbol?.baseAsset ||
    props.selectedProduct?.baseCoin ||
    props.selectedProduct?.symbol ||
    'BTC'
  )
})
const selectedMarginMode = computed(() => {
  return (
    marginModeOptions.value.find((option) => option.value === props.marginMode) ||
    marginModeOptions.value[0] || { value: 1, code: 'MARGIN_MODE_CROSS' }
  )
})
const conversionText = computed(() => {
  const contractSize = props.tradeSymbolDetail?.contract?.contractSize || ''
  if (isSpotTrade.value) return `1 ${baseAsset.value} = 1 ${baseAsset.value}`
  return contractSize ? `1 手 = ${contractSize} ${baseAsset.value}` : `1 手 = 1 ${baseAsset.value}`
})
const buyLabel = computed(() => (isSpotTrade.value ? '买入' : '买涨'))
const sellLabel = computed(() => (isSpotTrade.value ? '卖出' : '买跌'))
const unavailableText = computed(() => {
  if (!props.isLoggedIn) return '请先登录后再交易'
  if (props.tradeSymbolLoading) return '交易配置加载中'
  if (!props.tradeAvailable) return '当前品种暂未开放交易'
  return ''
})

function orderModeFromCode(code: string): 'market' | 'limit' {
  return code === 'ORDER_TYPE_LIMIT' ? 'limit' : 'market'
}

function inputValue(event: Event) {
  return (event.target as HTMLInputElement).value
}

function inputNumber(event: Event) {
  return Number((event.target as HTMLInputElement).value)
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
        <select
          :value="marginMode"
          :disabled="isSpotTrade"
          @change="emit('update:marginMode', inputNumber($event))"
        >
          <option v-for="option in marginModeOptions" :key="option.value" :value="option.value">
            {{ optionText(option) }}
          </option>
        </select>
      </div>
      <div>
        <label>杠杆</label>
        <select
          :value="leverage"
          :disabled="isSpotTrade"
          @change="emit('update:leverage', inputNumber($event))"
        >
          <option v-for="value in leverageValues" :key="value" :value="value">{{ value }}X</option>
        </select>
      </div>
    </div>

    <label v-if="orderMode === 'limit'" class="desktop-order-panel__label">
      价格 ({{ settleAsset }})
    </label>
    <div v-if="orderMode === 'limit'" class="trade-input trade-input--field">
      <input
        :value="tradePrice"
        inputmode="decimal"
        :placeholder="`最小变动 ${selectedTradeSymbol?.priceTick || '--'}`"
        @input="emit('update:tradePrice', inputValue($event))"
      />
    </div>

    <label class="desktop-order-panel__label"> 数量 ({{ baseAsset }}) </label>
    <div class="trade-input trade-input--field">
      <input
        :value="tradeQty"
        inputmode="decimal"
        :placeholder="`最小数量 ${selectedTradeSymbol?.minQty || '--'}`"
        @input="emit('update:tradeQty', inputValue($event))"
      />
    </div>
    <div class="percent-bar" :style="{ '--progress': `${tradePercent}%` }">
      <input
        class="percent-range"
        type="range"
        min="0"
        max="100"
        step="1"
        :value="tradePercent"
        @input="emit('update:tradePercent', inputNumber($event))"
      />
    </div>
    <div class="percent-labels">
      <span>0%</span><span>25%</span><span>50%</span><span>75%</span><span>100%</span>
    </div>

    <div class="account-lines">
      <span>可用</span><strong>{{ availableBalance }} {{ settleAsset }}</strong> <span>换算</span
      ><strong>{{ conversionText }}</strong> <span>模式</span
      ><strong
        >{{ optionText(selectedMarginMode) }} /
        {{ isSpotTrade ? '无杠杆' : `${leverage}X` }}</strong
      >
    </div>

    <label class="checkbox-line"><i />止盈/止损</label>
    <div class="account-lines">
      <span>{{ isSpotTrade ? '可买' : '可开多' }}</span
      ><strong>{{ longPositionQty }} {{ isSpotTrade ? baseAsset : '手' }}</strong>
      <span>保证金</span><strong>{{ availableBalance }} {{ settleAsset }}</strong>
    </div>
    <button
      class="wide-action wide-action--buy"
      type="button"
      :disabled="submitDisabled"
      @click="emit('submit-order', 'buy')"
    >
      {{ submittingSide === 'buy' ? '提交中' : buyLabel }}
    </button>

    <label class="checkbox-line"><i />止盈/止损</label>
    <div class="account-lines">
      <span>{{ isSpotTrade ? '可卖' : '可开空' }}</span
      ><strong>{{ shortPositionQty }} {{ isSpotTrade ? baseAsset : '手' }}</strong>
      <span>保证金</span><strong>{{ availableBalance }} {{ settleAsset }}</strong>
    </div>
    <button
      class="wide-action wide-action--sell"
      type="button"
      :disabled="submitDisabled"
      @click="emit('submit-order', 'sell')"
    >
      {{ submittingSide === 'sell' ? '提交中' : sellLabel }}
    </button>

    <p v-if="tradeError || unavailableText" class="order-message order-message--error">
      {{ tradeError || unavailableText }}
    </p>
    <p v-else-if="tradeMessage" class="order-message">{{ tradeMessage }}</p>
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

.desktop-order-panel__grid select,
.desktop-order-panel__grid input {
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

.desktop-order-panel__grid input:disabled,
.desktop-order-panel__grid select:disabled {
  color: #8f929d;
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

.trade-input--field input {
  width: 100%;
  min-width: 0;
  border: 0;
  outline: 0;
  background: transparent;
  color: #f6f7fb;
  font: inherit;
}

.trade-input--field input::placeholder {
  color: #7d828e;
}

.percent-bar {
  position: relative;
  height: 18px;
  margin-bottom: 10px;
  border-radius: 999px;
  background:
    linear-gradient(90deg, #02b904 0 var(--progress, 0%), transparent var(--progress, 0%)),
    linear-gradient(
      90deg,
      #1c1f2a 0 24%,
      transparent 24% 25%,
      #1c1f2a 25% 49%,
      transparent 49% 50%,
      #1c1f2a 50% 74%,
      transparent 74% 75%,
      #1c1f2a 75%
    );
}

.percent-bar i {
  display: block;
  width: 18px;
  height: 18px;
  border-radius: 999px;
  background: #02b904;
}

.percent-range {
  position: absolute;
  inset: -6px 0;
  width: 100%;
  opacity: 0;
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

.wide-action:disabled {
  cursor: not-allowed;
  opacity: 0.52;
}

.wide-action--buy {
  background: #02b904;
}

.wide-action--sell {
  background: #ff5a49;
}

.order-message {
  margin: 0;
  color: #10d27a;
  font-size: 13px;
  line-height: 1.5;
}

.order-message--error {
  color: #ff6b5f;
}
</style>
