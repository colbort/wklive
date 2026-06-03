<script setup lang="ts">
import { computed } from 'vue'

import { apiGetTradeOptions } from '@/api/trade'
import { optionText, useOptions } from '@/composables/useOptions'
import { useI18n } from '@/i18n'
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
const percentSteps = [0, 25, 50, 75, 100]
const tradeOptions = useOptions(apiGetTradeOptions)
const { t } = useI18n()
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
  return contractSize
    ? `1 ${t('trade.lot')} = ${contractSize} ${baseAsset.value}`
    : `1 ${t('trade.lot')} = 1 ${baseAsset.value}`
})
const buyLabel = computed(() => (isSpotTrade.value ? t('trade.buy') : t('trade.buyUp')))
const sellLabel = computed(() => (isSpotTrade.value ? t('trade.sell') : t('trade.buyDown')))
const unavailableText = computed(() => {
  if (!props.isLoggedIn) return t('trade.loginFirst')
  if (props.tradeSymbolLoading) return t('trade.configLoading')
  if (!props.tradeAvailable) return t('trade.unavailable')
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
        <label>{{ t('trade.margin') }}{{ t('trade.mode') }}</label>
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
        <label>{{ t('trade.leverage') }}</label>
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
      {{ t('trade.price') }} ({{ settleAsset }})
    </label>
    <div v-if="orderMode === 'limit'" class="trade-input trade-input--field">
      <input
        :value="tradePrice"
        inputmode="decimal"
        :placeholder="`${selectedTradeSymbol?.priceTick || '--'}`"
        @input="emit('update:tradePrice', inputValue($event))"
      />
    </div>

    <label class="desktop-order-panel__label"> {{ t('trade.qty') }} ({{ baseAsset }}) </label>
    <div class="trade-input trade-input--field">
      <input
        :value="tradeQty"
        inputmode="decimal"
        :placeholder="`${selectedTradeSymbol?.minQty || '--'}`"
        @input="emit('update:tradeQty', inputValue($event))"
      />
    </div>
    <div
      class="percent-bar"
      :style="{ '--progress': `${tradePercent}%` }"
      @pointerdown="handlePercentBarPointer"
    >
      <button
        v-for="value in percentSteps"
        :key="value"
        class="percent-hit"
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
      <span>{{ t('trade.available') }}</span><strong>{{ availableBalance }} {{ settleAsset }}</strong> <span>{{ t('trade.conversion') }}</span
      ><strong>{{ conversionText }}</strong> <span>{{ t('trade.mode') }}</span
      ><strong
        >{{ optionText(selectedMarginMode) }} /
        {{ isSpotTrade ? t('trade.noLeverage') : `${leverage}X` }}</strong
      >
    </div>

    <label class="checkbox-line"><i />{{ t('trade.tpSl') }}</label>
    <div class="account-lines">
      <span>{{ isSpotTrade ? t('trade.canBuy') : t('trade.canOpenLong') }}</span
      ><strong>{{ longPositionQty }} {{ isSpotTrade ? baseAsset : t('trade.lot') }}</strong>
      <span>{{ t('trade.margin') }}</span><strong>{{ availableBalance }} {{ settleAsset }}</strong>
    </div>
    <button
      class="wide-action wide-action--buy"
      type="button"
      :disabled="submitDisabled"
      @click="emit('submit-order', 'buy')"
    >
      {{ submittingSide === 'buy' ? t('common.submitting') : buyLabel }}
    </button>

    <label class="checkbox-line"><i />{{ t('trade.tpSl') }}</label>
    <div class="account-lines">
      <span>{{ isSpotTrade ? t('trade.canSell') : t('trade.canOpenShort') }}</span
      ><strong>{{ shortPositionQty }} {{ isSpotTrade ? baseAsset : t('trade.lot') }}</strong>
      <span>{{ t('trade.margin') }}</span><strong>{{ availableBalance }} {{ settleAsset }}</strong>
    </div>
    <button
      class="wide-action wide-action--sell"
      type="button"
      :disabled="submitDisabled"
      @click="emit('submit-order', 'sell')"
    >
      {{ submittingSide === 'sell' ? t('common.submitting') : sellLabel }}
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
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  height: 10px;
  margin-bottom: 6px;
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
  width: 10px;
  height: 10px;
  border-radius: 999px;
  background: #02b904;
}

.percent-hit {
  position: relative;
  z-index: 1;
  height: 28px;
  margin: -9px 0;
  border: 0;
  background: transparent;
  cursor: pointer;
}

.percent-labels {
  display: flex;
  justify-content: space-between;
  margin-bottom: 14px;
  color: #8f929d;
  font-size: 12px;
}

.percent-labels button {
  border: 0;
  padding: 0;
  background: transparent;
  color: inherit;
  font: inherit;
  cursor: pointer;
}

.percent-labels button.active {
  color: #f6f7fb;
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
