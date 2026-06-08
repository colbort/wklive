<script setup lang="ts">
import { computed, ref } from 'vue'

import { apiGetTradeOptions } from '@/api/trade'
import BottomDrawer from '@/components/common/BottomDrawer.vue'
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
  tradeKind: 'stock' | 'option' | 'forex' | 'commodity' | 'crypto'
  orderMode: 'market' | 'limit'
  selectedTradeSymbol: TradeSymbol | null
  tradeSymbolDetail: TradeSymbolDetail | null
  tradeSymbolLoading: boolean
  isLoggedIn: boolean
  tradeAvailable: boolean
  tradePrice: string
  tradeQty: string
  tradePercent: number
  referencePrice: string | number
  marginMode: number
  leverage: number
  maxLeverage: number
  leverageValues: number[]
  takeProfitPrice: string
  stopLossPrice: string
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
  (e: 'update:takeProfitPrice', value: string): void
  (e: 'update:stopLossPrice', value: string): void
  (e: 'submit-order', side: SubmitSide): void
}>()

const MARKET_TYPE_SPOT = 1
const percentSteps = [0, 25, 50, 75, 100]
type SelectionSheet = 'margin' | 'leverage' | 'risk' | null
type RiskEntrySide = 'long' | 'short'
const selectionSheet = ref<SelectionSheet>(null)
const riskEntrySide = ref<RiskEntrySide | null>(null)
const editingRiskEntrySide = ref<RiskEntrySide>('long')
const riskTakeProfitEnabled = ref(false)
const riskStopLossEnabled = ref(false)
const riskTakeProfitPrice = ref('')
const riskStopLossPrice = ref('')
const riskTakeProfitPercent = ref('')
const riskStopLossPercent = ref('')
const riskQty = ref('')
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
const buyLabel = computed(() => (isSpotTrade.value ? t('trade.buy') : t('trade.buyLong')))
const sellLabel = computed(() => (isSpotTrade.value ? t('trade.sell') : t('trade.sellShort')))
const riskSettingsActive = computed(() => Boolean(props.takeProfitPrice || props.stopLossPrice))
const riskTakeProfitReady = computed(
  () => riskTakeProfitEnabled.value && Boolean(validPositiveDecimal(riskTakeProfitPrice.value)),
)
const riskStopLossReady = computed(
  () => riskStopLossEnabled.value && Boolean(validPositiveDecimal(riskStopLossPrice.value)),
)
const riskQtyReady = computed(() => Boolean(validPositiveDecimal(riskQty.value)))
const riskCanConfirm = computed(
  () => (riskTakeProfitReady.value || riskStopLossReady.value) && riskQtyReady.value,
)
const riskReferencePrice = computed(() => {
  const orderPrice = parseDecimal(props.tradePrice)
  if (props.orderMode === 'limit' && orderPrice && orderPrice > 0) return orderPrice

  const latestPrice = parseDecimal(props.referencePrice)
  return latestPrice && latestPrice > 0 ? latestPrice : null
})
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

function updateTradePercent(value: number) {
  emit('update:tradePercent', value)
}

function handlePercentBarPointer(event: PointerEvent) {
  const rect = (event.currentTarget as HTMLElement).getBoundingClientRect()
  if (!rect.width) return

  const ratio = Math.min(Math.max((event.clientX - rect.left) / rect.width, 0), 1)
  updateTradePercent(Math.round((ratio * 100) / 25) * 25)
}

function parseDecimal(value: string | number | null | undefined) {
  const text = String(value ?? '')
    .trim()
    .replace(/,/g, '')
  if (!text) return null

  const numberValue = Number(text)
  return Number.isFinite(numberValue) ? numberValue : null
}

function validPositiveDecimal(value: string | number | null | undefined) {
  const numberValue = parseDecimal(value)
  return numberValue !== null && numberValue > 0 ? numberValue : null
}

function trimDecimal(value: string) {
  return value.replace(/(\.\d*?)0+$/, '$1').replace(/\.$/, '')
}

function priceScale() {
  const scale = Number(props.selectedTradeSymbol?.priceScale)
  if (Number.isFinite(scale) && scale >= 0) return Math.min(scale, 12)

  const tick = props.selectedTradeSymbol?.priceTick || ''
  const decimalText = tick.includes('.') ? tick.split('.')[1]?.replace(/0+$/, '') || '' : ''
  return Math.min(decimalText.length || 8, 12)
}

function formatPriceValue(value: number) {
  if (!Number.isFinite(value) || value <= 0) return ''
  return trimDecimal(value.toFixed(priceScale()))
}

function formatPercentValue(value: number) {
  if (!Number.isFinite(value)) return ''
  return trimDecimal(value.toFixed(2))
}

function riskPercentFromPrice(priceText: string, direction: 'up' | 'down') {
  const price = parseDecimal(priceText)
  const referencePrice = riskReferencePrice.value
  if (!price || price <= 0 || !referencePrice) return ''

  const percent =
    direction === 'up'
      ? ((price - referencePrice) / referencePrice) * 100
      : ((referencePrice - price) / referencePrice) * 100
  return formatPercentValue(percent)
}

function riskPriceFromPercent(percentText: string, direction: 'up' | 'down') {
  const percent = parseDecimal(percentText)
  const referencePrice = riskReferencePrice.value
  if (percent === null || !referencePrice) return ''

  const price =
    direction === 'up' ? referencePrice * (1 + percent / 100) : referencePrice * (1 - percent / 100)
  return formatPriceValue(price)
}

function updateRiskTakeProfitPrice(value: string) {
  riskTakeProfitPrice.value = value
  riskTakeProfitPercent.value = riskPercentFromPrice(value, 'up')
}

function updateRiskTakeProfitPercent(value: string) {
  riskTakeProfitPercent.value = value
  riskTakeProfitPrice.value = riskPriceFromPercent(value, 'up')
}

function updateRiskStopLossPrice(value: string) {
  riskStopLossPrice.value = value
  riskStopLossPercent.value = riskPercentFromPrice(value, 'down')
}

function updateRiskStopLossPercent(value: string) {
  riskStopLossPercent.value = value
  riskStopLossPrice.value = riskPriceFromPercent(value, 'down')
}

function isRiskEntryActive(side: RiskEntrySide) {
  return riskSettingsActive.value && riskEntrySide.value === side
}

const selectionSheetTitle = computed(() => {
  if (selectionSheet.value === 'margin') return `${t('trade.margin')}${t('trade.mode')}`
  if (selectionSheet.value === 'leverage') return t('trade.leverage')
  return t('trade.tpSl')
})

function openSelectionSheet(type: Exclude<SelectionSheet, null>, side?: RiskEntrySide) {
  if (isSpotTrade.value) return
  if (type === 'risk' && submitDisabled.value) return
  if (type === 'risk') {
    editingRiskEntrySide.value = side || riskEntrySide.value || 'long'
    riskTakeProfitEnabled.value = Boolean(props.takeProfitPrice)
    riskStopLossEnabled.value = Boolean(props.stopLossPrice)
    riskTakeProfitPrice.value = props.takeProfitPrice
    riskStopLossPrice.value = props.stopLossPrice
    riskTakeProfitPercent.value = riskPercentFromPrice(props.takeProfitPrice, 'up')
    riskStopLossPercent.value = riskPercentFromPrice(props.stopLossPrice, 'down')
    riskQty.value = ''
  }
  selectionSheet.value = type
}

function closeSelectionSheet() {
  selectionSheet.value = null
}

function selectMarginMode(value: number) {
  emit('update:marginMode', value)
  closeSelectionSheet()
}

function selectLeverage(value: number) {
  emit('update:leverage', value)
  closeSelectionSheet()
}

function confirmRiskSettings() {
  if (!riskCanConfirm.value) return

  const nextTakeProfitPrice = riskTakeProfitReady.value ? riskTakeProfitPrice.value : ''
  const nextStopLossPrice = riskStopLossReady.value ? riskStopLossPrice.value : ''

  emit('update:takeProfitPrice', nextTakeProfitPrice)
  emit('update:stopLossPrice', nextStopLossPrice)
  riskEntrySide.value = nextTakeProfitPrice || nextStopLossPrice ? editingRiskEntrySide.value : null
  closeSelectionSheet()
}
</script>

<template>
  <section v-if="tradeKind === 'stock'" class="stock-panel">
    <div class="stock-alert">
      <span>!</span>
      <strong>{{ t('trade.marketClosed') }}</strong>
    </div>

    <div class="inner-tabs">
      <button class="active" type="button">{{ t('trade.normalTrade') }}</button>
      <button type="button">{{ t('trade.marginTrading') }}</button>
      <button type="button">{{ t('trade.premarketOrders') }}</button>
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

    <div class="trade-input">{{ t('trade.qty') }}</div>
    <div class="percent-bar"><i /></div>
    <div class="percent-labels">
      <span>0%</span><span>25%</span><span>50%</span><span>75%</span><span>100%</span>
    </div>
    <button class="wide-action" type="button">{{ t('trade.loginRegister') }}</button>
  </section>

  <section v-else-if="tradeKind === 'option'" class="option-panel">
    <div class="mini-chart">
      <svg viewBox="0 0 320 88" :aria-label="t('trade.trend')">
        <path
          d="M0 50 C28 48 26 40 54 42 C88 47 82 28 119 25 C152 20 143 32 176 28 C218 26 199 64 231 63 C266 62 252 49 320 53"
        />
      </svg>
    </div>

    <h3>{{ t('trade.time') }}</h3>
    <div class="duration-grid">
      <button class="active" type="button"><strong>30S</strong><span>30%</span></button>
      <button type="button"><strong>60S</strong><span>40%</span></button>
      <button type="button"><strong>90S</strong><span>50%</span></button>
      <button type="button"><strong>120S</strong><span>60%</span></button>
      <button type="button"><strong>180S</strong><span>70%</span></button>
      <button type="button"><strong>360S</strong><span>80%</span></button>
    </div>

    <h3>{{ t('trade.investmentAmount') }}</h3>
    <div class="trade-input split"><span>>=100</span><strong>USD</strong></div>
    <div class="percent-bar"><i /></div>
    <div class="percent-labels">
      <span>0%</span><span>25%</span><span>50%</span><span>75%</span><span>100%</span>
    </div>
    <div class="buyable">
      <span>{{ t('trade.canBuy') }}</span
      ><strong>0 USD</strong>
    </div>
    <div class="dual-actions">
      <button type="button">{{ t('auth.loginTitle') }}</button>
      <button type="button">{{ t('auth.goRegister') }}</button>
    </div>
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
        <button
          type="button"
          class="picker-trigger"
          :disabled="isSpotTrade"
          @click="openSelectionSheet('margin')"
        >
          {{ optionText(selectedMarginMode) }}
        </button>
        <button
          type="button"
          class="picker-trigger"
          :disabled="isSpotTrade"
          @click="openSelectionSheet('leverage')"
        >
          {{ leverage }}X
        </button>
      </div>

      <div v-if="orderMode === 'limit'" class="trade-input trade-input--field">
        <input
          :value="tradePrice"
          inputmode="decimal"
          :placeholder="`${t('trade.price')}(${settleAsset}) / ${selectedTradeSymbol?.priceTick || '--'}`"
          @input="emit('update:tradePrice', inputValue($event))"
        />
      </div>

      <div class="trade-input trade-input--field">
        <input
          :value="tradeQty"
          inputmode="decimal"
          :placeholder="`${t('trade.qty')}(${baseAsset}) / ${selectedTradeSymbol?.minQty || '--'}`"
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
        <span>{{ t('trade.available') }}</span>
        <strong>{{ availableBalance }} {{ settleAsset }}</strong>
        <span>{{ t('trade.conversion') }}</span
        ><strong>{{ conversionText }}</strong>
        <span>{{ t('trade.mode') }}</span>
        <strong
          >{{ optionText(selectedMarginMode) }} /
          {{ isSpotTrade ? t('trade.noLeverage') : `${leverage}X` }}</strong
        >
      </div>

      <button
        type="button"
        class="risk-trigger"
        :disabled="submitDisabled"
        @click="openSelectionSheet('risk', 'long')"
      >
        <span
          :class="{
            active: isRiskEntryActive('long'),
          }"
        />
        {{ t('trade.tpSl') }}
      </button>
      <div class="account-lines">
        <span>{{ isSpotTrade ? t('trade.canBuy') : t('trade.canOpenLong') }}</span>
        <strong>{{ longPositionQty }} {{ isSpotTrade ? baseAsset : t('trade.lot') }}</strong>
        <span>{{ t('trade.margin') }}</span
        ><strong>{{ availableBalance }} {{ settleAsset }}</strong>
      </div>
      <button
        class="wide-action wide-action--buy"
        type="button"
        :disabled="submitDisabled"
        @click="emit('submit-order', 'buy')"
      >
        {{ submittingSide === 'buy' ? t('common.submitting') : buyLabel }}
      </button>

      <button
        type="button"
        class="risk-trigger"
        :disabled="submitDisabled"
        @click="openSelectionSheet('risk', 'short')"
      >
        <span
          :class="{
            active: isRiskEntryActive('short'),
          }"
        />
        {{ t('trade.tpSl') }}
      </button>
      <div class="account-lines">
        <span>{{ isSpotTrade ? t('trade.canSell') : t('trade.canOpenShort') }}</span>
        <strong>{{ shortPositionQty }} {{ isSpotTrade ? baseAsset : t('trade.lot') }}</strong>
        <span>{{ t('trade.margin') }}</span
        ><strong>{{ availableBalance }} {{ settleAsset }}</strong>
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
    </div>

    <BottomDrawer
      :model-value="Boolean(selectionSheet)"
      :title="selectionSheetTitle"
      :close-label="t('common.close')"
      :max-height="selectionSheet === 'risk' ? '88dvh' : '72dvh'"
      :z-index="80"
      @update:model-value="
        (value) => {
          if (!value) closeSelectionSheet()
        }
      "
    >
      <div
        class="selection-sheet-content"
        :class="[
          `selection-sheet-content--${selectionSheet || 'none'}`,
          {
            'selection-sheet-content--risk-active':
              selectionSheet === 'risk' && (riskTakeProfitEnabled || riskStopLossEnabled),
            'selection-sheet-content--risk-full':
              selectionSheet === 'risk' && riskTakeProfitEnabled && riskStopLossEnabled,
          },
        ]"
      >
        <div v-if="selectionSheet === 'margin'" class="margin-mode-list">
          <button
            v-for="option in marginModeOptions"
            :key="option.value"
            type="button"
            :class="{ active: option.value === marginMode }"
            @click="selectMarginMode(option.value)"
          >
            {{ optionText(option) }}
          </button>
        </div>

        <div v-else-if="selectionSheet === 'leverage'" class="leverage-grid">
          <button
            v-for="value in leverageValues"
            :key="value"
            type="button"
            :class="{ active: value === leverage }"
            @click="selectLeverage(value)"
          >
            {{ value }}X
          </button>
        </div>

        <div v-else-if="selectionSheet === 'risk'" class="risk-sheet-body">
          <label class="risk-check-row">
            <input v-model="riskTakeProfitEnabled" type="checkbox" />
            <span />
            <strong>{{ t('options.TRIGGER_KIND_TAKE_PROFIT') }}</strong>
          </label>
          <div v-if="riskTakeProfitEnabled" class="risk-input-grid">
            <label class="risk-field risk-field--wide">
              <input
                :value="riskTakeProfitPrice"
                inputmode="decimal"
                :placeholder="t('trade.price')"
                @input="updateRiskTakeProfitPrice(inputValue($event))"
              />
              <strong>{{ settleAsset }}</strong>
            </label>
            <label class="risk-field">
              <input
                :value="riskTakeProfitPercent"
                inputmode="decimal"
                :placeholder="t('trade.risePercent')"
                @input="updateRiskTakeProfitPercent(inputValue($event))"
              />
              <strong>%</strong>
            </label>
          </div>

          <label class="risk-check-row">
            <input v-model="riskStopLossEnabled" type="checkbox" />
            <span />
            <strong>{{ t('options.TRIGGER_KIND_STOP_LOSS') }}</strong>
          </label>
          <div v-if="riskStopLossEnabled" class="risk-input-grid">
            <label class="risk-field risk-field--wide">
              <input
                :value="riskStopLossPrice"
                inputmode="decimal"
                :placeholder="t('trade.price')"
                @input="updateRiskStopLossPrice(inputValue($event))"
              />
              <strong>{{ settleAsset }}</strong>
            </label>
            <label class="risk-field">
              <input
                :value="riskStopLossPercent"
                inputmode="decimal"
                :placeholder="t('trade.fallPercent')"
                @input="updateRiskStopLossPercent(inputValue($event))"
              />
              <strong>%</strong>
            </label>
          </div>

          <div class="risk-divider" />
          <label v-if="riskTakeProfitEnabled || riskStopLossEnabled" class="risk-qty">
            <strong>{{ t('trade.qty') }}</strong>
            <input v-model="riskQty" inputmode="decimal" :placeholder="t('trade.qty')" />
          </label>

          <button
            type="button"
            class="risk-confirm"
            :disabled="!riskCanConfirm"
            @click="confirmRiskSettings"
          >
            {{ t('common.submit') }}
          </button>
        </div>
      </div>
    </BottomDrawer>
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
  background: #242631;
}

.mode-switch.compact {
  max-width: 260px;
}

.mode-switch button,
.form-row button,
.trade-input,
.risk-trigger,
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
  color: var(--muted);
  font-size: 0.8rem;
  font-weight: 600;
}

.mode-switch button.active {
  border-radius: 999px;
  background: var(--accent);
  color: var(--text);
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
  margin-bottom: 12px;
}

.form-row button,
.form-row select,
.form-row input,
.trade-input {
  min-height: 48px;
  padding: 0 12px;
  border-radius: 12px;
  background: #242631;
  color: var(--text);
  text-align: left;
}

.form-row button,
.form-row select,
.form-row input {
  border: 0;
  outline: 0;
  font-size: 0.8rem;
}

.picker-trigger {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
}

.picker-trigger::after {
  position: absolute;
  right: 16px;
  width: 8px;
  height: 8px;
  border-right: 2px solid #777c88;
  border-bottom: 2px solid #777c88;
  transform: rotate(45deg) translateY(-2px);
  content: '';
}

.picker-trigger:disabled::after {
  opacity: 0.45;
}

.margin-mode-list {
  display: grid;
  gap: clamp(24px, 4.6vw, 34px);
}

.margin-mode-list button,
.leverage-grid button {
  min-height: clamp(64px, 12.5vw, 90px);
  border: 0;
  border-radius: clamp(20px, 3.9vw, 28px);
  background: #3d3e47;
  color: var(--text);
  font: inherit;
  font-size: clamp(1.05rem, 3.9vw, 1.4rem);
  font-weight: 800;
}

.margin-mode-list button.active,
.leverage-grid button.active {
  background: var(--accent);
}

.leverage-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: clamp(16px, 2.8vw, 20px) clamp(18px, 2.8vw, 20px);
}

.leverage-grid button {
  min-height: clamp(58px, 10.3vw, 74px);
  border-radius: clamp(18px, 3.2vw, 23px);
}

.risk-sheet-body {
  display: grid;
  gap: clamp(20px, 2.8vw, 22px);
  width: 100%;
}

.risk-check-row {
  display: flex;
  align-items: center;
  gap: 12px;
  color: var(--text);
  font-size: clamp(0.9rem, 2.9vw, 1rem);
  font-weight: 700;
}

.risk-check-row input {
  position: absolute;
  opacity: 0;
  pointer-events: none;
}

.risk-check-row span {
  position: relative;
  width: clamp(22px, 3.4vw, 25px);
  height: clamp(22px, 3.4vw, 25px);
  flex: 0 0 auto;
  border: 2px solid #d9ddea;
  border-radius: 7px;
}

.risk-check-row input:checked + span {
  border-color: var(--accent);
  background: var(--accent);
}

.risk-check-row input:checked + span::after {
  position: absolute;
  top: 4px;
  left: 7px;
  width: 6px;
  height: 11px;
  border-right: 3px solid #fff;
  border-bottom: 3px solid #fff;
  transform: rotate(45deg);
  content: '';
}

.risk-input-grid {
  display: grid;
  width: 100%;
  grid-template-columns: minmax(0, 1fr) minmax(0, 0.55fr);
  gap: clamp(10px, 1.7vw, 12px);
}

.risk-field {
  display: flex;
  align-items: center;
  min-height: clamp(52px, 9.8vw, 64px);
  padding: 0 clamp(18px, 3.7vw, 24px);
  border-radius: clamp(16px, 3vw, 20px);
  background: #3d3e47;
}

.risk-field input,
.risk-qty input {
  width: 100%;
  min-width: 0;
  border: 0;
  outline: 0;
  background: transparent;
  color: var(--text);
  font: inherit;
  font-size: clamp(0.85rem, 2.8vw, 1rem);
  font-weight: 600;
}

.risk-field input::placeholder,
.risk-qty input::placeholder {
  color: #8d8f97;
}

.risk-field strong {
  margin-left: clamp(8px, 1.8vw, 12px);
  color: var(--text);
  font-size: clamp(0.85rem, 2.8vw, 1rem);
  font-weight: 700;
}

.risk-divider {
  height: 1px;
  background: #3d3e47;
}

.risk-qty {
  display: grid;
  gap: clamp(16px, 2.6vw, 18px);
  color: var(--text);
  font-size: clamp(0.9rem, 2.9vw, 1rem);
  font-weight: 700;
}

.risk-qty input {
  min-height: clamp(52px, 9.8vw, 64px);
  padding: 0 clamp(20px, 3.7vw, 24px);
  border-radius: clamp(16px, 3vw, 20px);
  background: #3d3e47;
}

.risk-confirm {
  min-height: clamp(60px, 11.8vw, 78px);
  margin-top: clamp(14px, 2.4vw, 18px);
  border: 0;
  border-radius: clamp(24px, 5.5vw, 36px);
  background: var(--accent);
  color: var(--text);
  font: inherit;
  font-size: clamp(0.95rem, 2.9vw, 1.05rem);
  font-weight: 700;
}

.risk-confirm:disabled {
  cursor: default;
  opacity: 0.45;
}

.form-row button:disabled,
.form-row select:disabled,
.form-row input:disabled {
  color: var(--muted);
}

.trade-input {
  display: flex;
  align-items: center;
  margin-bottom: 12px;
  color: var(--muted);
  font-size: 0.8rem;
}

.trade-input--field input {
  width: 100%;
  min-width: 0;
  border: 0;
  outline: 0;
  background: transparent;
  color: var(--text);
  font: inherit;
}

.trade-input--field input::placeholder {
  color: var(--muted);
}

.trade-input.split {
  justify-content: space-between;
}

.trade-input.split strong {
  color: var(--text);
  font-weight: 500;
}

.percent-bar {
  position: relative;
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  height: 8px;
  margin-bottom: 6px;
  border-radius: 999px;
  background:
    linear-gradient(90deg, var(--accent) 0 var(--progress, 0%), transparent var(--progress, 0%)),
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
  width: 8px;
  height: 8px;
  border-radius: 999px;
  background: var(--accent);
}

.percent-hit {
  position: relative;
  z-index: 1;
  height: 28px;
  margin: -10px 0;
  border: 0;
  background: transparent;
  cursor: pointer;
}

.percent-labels {
  display: flex;
  justify-content: space-between;
  margin-bottom: 14px;
  color: var(--muted);
  font-size: 0.65rem;
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

.risk-trigger {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 7px 0 14px;
  padding: 0;
  color: var(--text);
  font-size: 0.75rem;
  font-weight: 600;
}

.risk-trigger span {
  position: relative;
  width: 18px;
  height: 18px;
  border: 1px solid var(--text);
  border-radius: 4px;
}

.risk-trigger span.active {
  border-color: var(--accent);
  background: var(--accent);
}

.risk-trigger span.active::after {
  position: absolute;
  top: 3px;
  left: 6px;
  width: 5px;
  height: 9px;
  border-right: 2px solid #fff;
  border-bottom: 2px solid #fff;
  transform: rotate(45deg);
  content: '';
}

.risk-trigger:disabled {
  cursor: not-allowed;
  color: var(--muted);
}

.risk-trigger:disabled span {
  border-color: var(--muted);
}

.wide-action,
.dual-actions button {
  min-height: 46px;
  border-radius: 12px;
  background: #181b25;
  color: var(--text);
  font-size: 0.75rem;
  font-weight: 700;
}

.wide-action {
  width: 100%;
  margin-bottom: 18px;
}

.wide-action--buy {
  background: #10d27a;
}

.wide-action--sell {
  background: #ff4438;
}

.wide-action:disabled {
  cursor: not-allowed;
  opacity: 0.52;
}

.order-message {
  margin: -4px 0 12px;
  color: #10d27a;
  font-size: 0.65rem;
  line-height: 1.45;
}

.order-message--error {
  color: #ff6b5f;
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
  color: var(--page-bg);
}

.stock-alert strong {
  font-size: 0.75rem;
  font-weight: 500;
}

.inner-tabs {
  display: flex;
  gap: 22px;
  margin-bottom: 22px;
}

.inner-tabs button {
  position: relative;
  color: var(--muted);
  font-size: 0.85rem;
}

.inner-tabs button.active {
  color: var(--text);
}

.inner-tabs button.active::after {
  position: absolute;
  right: 0;
  bottom: -12px;
  left: 0;
  height: 3px;
  border-radius: 999px;
  background: var(--accent);
  content: '';
}

.option-panel h3 {
  margin: 24px 0 14px;
  font-size: 0.85rem;
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
  color: var(--muted);
  text-align: left;
}

.duration-grid button.active {
  background: var(--accent);
  color: var(--text);
}

.duration-grid strong {
  font-size: 0.95rem;
  font-weight: 600;
}

.duration-grid span {
  font-size: 0.65rem;
}

.buyable {
  display: grid;
  grid-template-columns: 1fr auto;
  margin-bottom: 22px;
  color: var(--muted);
  font-size: 0.7rem;
}

.buyable strong {
  color: var(--text);
  font-weight: 500;
}

@media (max-width: 390px) {
  .mode-switch {
    min-height: 38px;
  }

  .mode-switch button,
  .form-row button,
  .form-row select,
  .form-row input,
  .trade-input {
    font-size: 0.7rem;
  }

  .form-row {
    gap: 8px;
  }

  .form-row button,
  .trade-input {
    min-height: 44px;
    padding: 0 10px;
  }

  .percent-labels,
  .account-lines {
    font-size: 0.6rem;
  }

  .risk-trigger {
    font-size: 0.7rem;
  }
}
</style>
