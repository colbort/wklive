<script setup lang="ts">
import { computed, ref } from 'vue'

import type { Interval } from '@/types/core'
import type { DepthPayload, ItickTenantProduct, KlinePayload, QuotePayload, TickPayload } from '@/types/itick'
import type { MarketRow } from './types'

type DetailTab = 'market' | 'depth' | 'trades'

const props = defineProps<{
  products: ItickTenantProduct[]
  rows: MarketRow[]
  categoryName: string
  selectedProductKey: string
  selectedQuote: QuotePayload | null
  klineSnapshot: KlinePayload[]
  depthSnapshot: DepthPayload | null
  tickSnapshot: TickPayload[]
  loadingKline: boolean
  intervals: Interval[]
  selectedIntervalName: string
}>()

const PRODUCT_PAGE_SIZE = 4

const emit = defineEmits<{
  selectProduct: [productKey: string]
  selectInterval: [interval: Interval]
  loadPreviousPage: []
}>()

const swipeStartX = ref<number | null>(null)
const swipeStartY = ref<number | null>(null)
const switcherOpen = ref(false)
const activeDetailTab = ref<DetailTab>('market')
const productPage = ref(0)
const sheetSwipeStartX = ref<number | null>(null)
const sheetSwipeStartY = ref<number | null>(null)

const selectedProduct = computed(() => {
  return props.products.find((item) => productKey(item) === props.selectedProductKey) ?? null
})
const selectedPriceChange = computed(() => getChangeRate(props.selectedQuote))
const selectedTrendClass = computed(() => trendClass(selectedPriceChange.value))
const selectedChangeValue = computed(() => {
  const quote = props.selectedQuote
  return quote ? quote.lastPrice - quote.open : 0
})

const chartCandles = computed(() => {
  const source = [...props.klineSnapshot].sort((left, right) => left.ts - right.ts).slice(-26)
  const prices = source.flatMap((item) => [item.high, item.low]).filter((item) => item > 0)
  const max = Math.max(...prices, 1)
  const min = Math.min(...prices, max * 0.98)
  const range = Math.max(max - min, max * 0.01)

  return source.map((item, index) => {
    const high = 12 + ((max - item.high) / range) * 164
    const low = 12 + ((max - item.low) / range) * 164
    const open = 12 + ((max - item.open) / range) * 164
    const close = 12 + ((max - item.close) / range) * 164
    const bodyTop = Math.min(open, close)
    const bodyHeight = Math.max(Math.abs(close - open), 4)

    return {
      key: `${item.ts}-${index}`,
      x: 14 + index * 18,
      high,
      low,
      bodyTop,
      bodyHeight,
      up: item.close >= item.open,
    }
  })
})

const chartPriceMarks = computed(() => {
  const quote = props.selectedQuote
  const high = quote?.high || props.klineSnapshot[0]?.high || 0
  const low = quote?.low || props.klineSnapshot[0]?.low || 0
  const last = quote?.lastPrice || props.klineSnapshot[0]?.close || 0

  return [high, last, low].filter((item, index, list) => item > 0 && list.indexOf(item) === index)
})

const chartStats = computed(() => {
  const quote = props.selectedQuote
  return [
    { label: '最高', value: formatNumber(quote?.high, 2), accent: true },
    { label: '最低', value: formatNumber(quote?.low, 2), accent: true },
    { label: '成交额', value: formatCompact(quote?.turnover) },
    { label: '成交量', value: formatCompact(quote?.volume) },
  ]
})
const productPageCount = computed(() => Math.max(1, Math.ceil(props.rows.length / PRODUCT_PAGE_SIZE)))
const visibleProductRows = computed(() => {
  const start = productPage.value * PRODUCT_PAGE_SIZE
  return props.rows.slice(start, start + PRODUCT_PAGE_SIZE)
})
const askRows = computed(() => props.depthSnapshot?.asks.slice(0, 6) ?? [])
const bidRows = computed(() => props.depthSnapshot?.bids.slice(0, 6) ?? [])
const maxAskVolume = computed(() => Math.max(...askRows.value.map((item) => item.volume), 1))
const maxBidVolume = computed(() => Math.max(...bidRows.value.map((item) => item.volume), 1))
const tradeRows = computed(() =>
  props.tickSnapshot.map((item, index, list) => {
    const next = list[index + 1]
    const direction = next && item.lastPrice < next.lastPrice ? 'down' : 'up'

    return {
      ...item,
      direction,
    }
  }),
)

function productKey(product: Pick<ItickTenantProduct, 'market' | 'symbol'>) {
  return `${String(product.market || '').toUpperCase()}::${String(product.symbol || '').toUpperCase()}`
}

function openSwitcher() {
  const selectedIndex = props.rows.findIndex((item) => item.key === props.selectedProductKey)
  productPage.value = selectedIndex > -1 ? Math.floor(selectedIndex / PRODUCT_PAGE_SIZE) : 0
  switcherOpen.value = true
}

function closeSwitcher() {
  switcherOpen.value = false
}

function selectProductRow(row: MarketRow) {
  emit('selectProduct', row.key)
  closeSwitcher()
}

function handleSheetPointerStart(event: TouchEvent | MouseEvent) {
  const point = getClientPoint(event)
  sheetSwipeStartX.value = point.x
  sheetSwipeStartY.value = point.y
}

function handleSheetPointerEnd(event: TouchEvent | MouseEvent) {
  if (sheetSwipeStartX.value === null || sheetSwipeStartY.value === null) return

  const point = getClientPoint(event)
  const deltaX = point.x - sheetSwipeStartX.value
  const deltaY = point.y - sheetSwipeStartY.value

  sheetSwipeStartX.value = null
  sheetSwipeStartY.value = null

  if (Math.abs(deltaX) < 48 || Math.abs(deltaY) > 42) return

  if (deltaX < 0 && productPage.value < productPageCount.value - 1) {
    productPage.value += 1
  }

  if (deltaX > 0 && productPage.value > 0) {
    productPage.value -= 1
  }
}

function handlePointerStart(event: TouchEvent | MouseEvent) {
  const point = getClientPoint(event)
  swipeStartX.value = point.x
  swipeStartY.value = point.y
}

function handlePointerEnd(event: TouchEvent | MouseEvent) {
  if (swipeStartX.value === null || swipeStartY.value === null) return

  const point = getClientPoint(event)
  const deltaX = point.x - swipeStartX.value
  const deltaY = point.y - swipeStartY.value

  swipeStartX.value = null
  swipeStartY.value = null

  if (deltaX > 56 && Math.abs(deltaY) < 42) {
    emit('loadPreviousPage')
  }
}

function getClientPoint(event: TouchEvent | MouseEvent) {
  if ('changedTouches' in event && event.changedTouches.length > 0) {
    return {
      x: event.changedTouches[0].clientX,
      y: event.changedTouches[0].clientY,
    }
  }

  if ('touches' in event && event.touches.length > 0) {
    return {
      x: event.touches[0].clientX,
      y: event.touches[0].clientY,
    }
  }

  return {
    x: (event as MouseEvent).clientX,
    y: (event as MouseEvent).clientY,
  }
}

function getChangeRate(quote?: QuotePayload | null) {
  if (!quote || !quote.open) return 0
  return ((quote.lastPrice - quote.open) / quote.open) * 100
}

function trendClass(value: number) {
  if (value > 0) return 'up'
  if (value < 0) return 'down'
  return 'flat'
}

function formatNumber(value?: number | null, digits = 2) {
  if (value === null || value === undefined || !Number.isFinite(value)) return '--'

  return new Intl.NumberFormat('zh-CN', {
    minimumFractionDigits: digits,
    maximumFractionDigits: digits,
  }).format(value)
}

function formatCompact(value?: number | null) {
  if (value === null || value === undefined || !Number.isFinite(value)) return '--'

  return new Intl.NumberFormat('zh-CN', {
    notation: 'compact',
    maximumFractionDigits: 2,
  }).format(value)
}

function formatPercent(value: number) {
  return `${value >= 0 ? '+' : ''}${value.toFixed(2)}%`
}

function formatTime(ts: number) {
  if (!ts) return '--'

  return new Intl.DateTimeFormat('zh-CN', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false,
  }).format(ts)
}

function coinGlyph(product: ItickTenantProduct) {
  const coin = product.baseCoin || product.symbol.slice(0, 3)
  return coin.slice(0, 1).toUpperCase()
}
</script>

<template>
  <section class="chart-view">
    <div class="chart-switcher">
      <button type="button" class="symbol-switch" @click="openSwitcher">
        <span>{{ selectedProduct?.symbol || '--' }}</span>
        <em />
      </button>

      <button type="button" class="star-button" aria-label="加入自选">☆</button>
    </div>

    <div class="chart-summary">
      <div>
        <strong class="chart-summary__price" :class="selectedTrendClass">
          {{ selectedQuote ? formatNumber(selectedQuote.lastPrice, 2) : '--' }}
        </strong>
        <span :class="selectedTrendClass">
          {{ selectedQuote ? `${selectedChangeValue >= 0 ? '+' : ''}${formatNumber(selectedChangeValue, 2)}  ${formatPercent(selectedPriceChange)}` : '--' }}
        </span>
        <small>今开 {{ formatNumber(selectedQuote?.open, 2) }}</small>
        <small>昨收 {{ formatNumber(selectedQuote?.open, 2) }}</small>
      </div>

      <div class="chart-stats">
        <span v-for="item in chartStats" :key="item.label">
          <em>{{ item.label }}</em>
          <strong :class="{ up: item.accent }">{{ item.value }}</strong>
        </span>
      </div>
    </div>

    <div class="chart-sticky-tabs">
      <div class="sub-tabs">
        <button
          type="button"
          class="sub-tab"
          :class="{ 'sub-tab--active': activeDetailTab === 'market' }"
          @click="activeDetailTab = 'market'"
        >
          行情
        </button>
        <button
          type="button"
          class="sub-tab"
          :class="{ 'sub-tab--active': activeDetailTab === 'depth' }"
          @click="activeDetailTab = 'depth'"
        >
          订单簿
        </button>
        <button
          type="button"
          class="sub-tab"
          :class="{ 'sub-tab--active': activeDetailTab === 'trades' }"
          @click="activeDetailTab = 'trades'"
        >
          最新成交
        </button>
      </div>
    </div>

    <template v-if="activeDetailTab === 'market'">
      <div class="interval-row">
        <button
          v-for="item in intervals"
          :key="`${item.name}-${item.kType}`"
          type="button"
          class="interval-pill"
          :class="{ 'interval-pill--active': item.name === selectedIntervalName }"
          @click="emit('selectInterval', item)"
        >
          {{ item.name }}
        </button>
      </div>

      <div class="tool-row">
        <button type="button">ƒ 指标</button>
        <button type="button">◉ 时区</button>
        <button type="button">⚙ 设置</button>
      </div>

      <div
        class="chart-board"
        @touchstart.passive="handlePointerStart"
        @touchend.passive="handlePointerEnd"
        @mousedown="handlePointerStart"
        @mouseup="handlePointerEnd"
      >
        <div class="chart-tools" aria-hidden="true">
          <span />
          <span />
          <span />
          <span />
          <span />
        </div>

        <svg class="candle-chart" viewBox="0 0 520 240" role="img" aria-label="K线图">
          <line x1="0" y1="58" x2="520" y2="58" class="grid-line" />
          <line x1="0" y1="120" x2="520" y2="120" class="grid-line" />
          <line x1="0" y1="182" x2="520" y2="182" class="grid-line" />
          <line x1="170" y1="0" x2="170" y2="240" class="grid-line" />
          <line x1="340" y1="0" x2="340" y2="240" class="grid-line" />

          <g v-for="candle in chartCandles" :key="candle.key">
            <line
              :x1="candle.x"
              :x2="candle.x"
              :y1="candle.high"
              :y2="candle.low"
              :class="candle.up ? 'candle-up' : 'candle-down'"
            />
            <rect
              :x="candle.x - 5"
              :y="candle.bodyTop"
              width="10"
              :height="candle.bodyHeight"
              rx="1"
              :class="candle.up ? 'candle-up' : 'candle-down'"
            />
          </g>
        </svg>

        <div class="price-axis">
          <span v-for="mark in chartPriceMarks" :key="mark">{{ formatNumber(mark, 2) }}</span>
        </div>

        <div v-if="loadingKline" class="chart-loading">加载中...</div>
      </div>

      <div class="volume-board" aria-hidden="true">
        <div class="volume-labels">
          <span>VOL(5,10,20)</span>
          <strong>VOLUME</strong>
        </div>
        <div class="volume-bars">
          <span
            v-for="candle in chartCandles"
            :key="`volume-${candle.key}`"
            :class="candle.up ? 'volume-up' : 'volume-down'"
            :style="{ height: `${24 + (candle.bodyHeight % 72)}px` }"
          />
        </div>
      </div>
    </template>

    <section v-else-if="activeDetailTab === 'depth'" class="depth-board">
      <header class="depth-board__head">
        <span>价格<br />({{ selectedProduct?.quoteCoin || 'USDT' }})</span>
        <span>数量<br />({{ selectedProduct?.baseCoin || selectedProduct?.symbol || '--' }})</span>
      </header>

      <div class="depth-list depth-list--asks">
        <div
          v-for="(item, index) in askRows"
          :key="`ask-${item.price}-${index}`"
          class="depth-row depth-row--ask"
        >
          <i :style="{ width: `${Math.max(8, (item.volume / maxAskVolume) * 100)}%` }" />
          <span>{{ formatNumber(item.price, 4) }}</span>
          <strong>{{ formatNumber(item.volume, 8) }}</strong>
        </div>
      </div>

      <div class="depth-mid" :class="selectedTrendClass">
        <strong>{{ selectedQuote ? formatNumber(selectedQuote.lastPrice, 4) : '--' }}</strong>
        <span>{{ formatNumber(selectedQuote?.open, 4) }}</span>
      </div>

      <div class="depth-list depth-list--bids">
        <div
          v-for="(item, index) in bidRows"
          :key="`bid-${item.price}-${index}`"
          class="depth-row depth-row--bid"
        >
          <i :style="{ width: `${Math.max(8, (item.volume / maxBidVolume) * 100)}%` }" />
          <span>{{ formatNumber(item.price, 4) }}</span>
          <strong>{{ formatNumber(item.volume, 8) }}</strong>
        </div>
      </div>

      <p v-if="!askRows.length && !bidRows.length" class="detail-empty">等待订单簿推送</p>
    </section>

    <section v-else class="trade-board">
      <header class="trade-board__head">
        <span>价格({{ selectedProduct?.quoteCoin || 'USDT' }})</span>
        <span>数量({{ selectedProduct?.baseCoin || selectedProduct?.symbol || '--' }})</span>
        <span>时间</span>
      </header>

      <div v-if="!tradeRows.length" class="detail-empty">等待最新成交推送</div>
      <div
        v-for="(item, index) in tradeRows"
        v-else
        :key="`${item.ts}-${index}`"
        class="trade-row"
        :class="{ 'trade-row--down': item.direction === 'down' }"
      >
        <span>{{ formatNumber(item.lastPrice, 4) }}</span>
        <strong>{{ formatNumber(item.volume, 8) }}</strong>
        <time>{{ formatTime(item.ts) }}</time>
      </div>
    </section>

    <div v-if="switcherOpen" class="product-sheet-overlay" @click.self="closeSwitcher">
      <section
        class="product-sheet"
        @touchstart.passive="handleSheetPointerStart"
        @touchend.passive="handleSheetPointerEnd"
        @mousedown="handleSheetPointerStart"
        @mouseup="handleSheetPointerEnd"
      >
        <span class="product-sheet__handle" />

        <header class="product-sheet__header">
          <h3>{{ categoryName || '产品' }}</h3>
          <button type="button" aria-label="关闭" @click="closeSwitcher">×</button>
        </header>

        <div class="product-sheet__rows">
          <button
            v-for="row in visibleProductRows"
            :key="row.key"
            type="button"
            class="product-sheet-row"
            :class="{
              'product-sheet-row--active': row.key === selectedProductKey,
              'product-sheet-row--down': row.direction === 'down',
            }"
            @click="selectProductRow(row)"
          >
            <span class="product-sheet-row__coin">{{ coinGlyph(row.product) }}</span>
            <span class="product-sheet-row__symbol">{{ row.product.symbol }}</span>
            <strong>{{ row.quote ? formatNumber(row.quote.lastPrice, 4) : '--' }}</strong>
            <span class="product-sheet-row__change">
              <em>{{ row.quote ? formatNumber(row.quote.lastPrice - row.quote.open, 4) : '--' }}</em>
              <small>{{ row.quote ? formatPercent(row.changeRate) : '等待' }}</small>
            </span>
          </button>
        </div>

        <div class="product-sheet__footer">
          <span v-if="productPageCount > 1">{{ productPage + 1 }} / {{ productPageCount }}</span>
          <span v-else>没有更多了</span>
        </div>
      </section>
    </div>
  </section>
</template>

<style scoped>
.chart-view {
  padding: 20px 0 0;
}

.chart-switcher,
.chart-summary,
.interval-row,
.tool-row {
  padding-right: 18px;
  padding-left: 18px;
}

.chart-switcher {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
}

.symbol-switch {
  display: inline-flex;
  align-items: center;
  gap: 9px;
  min-width: 140px;
  border: 0;
  background: transparent;
  color: #fff;
  font: inherit;
  font-size: 18px;
  font-weight: 500;
  cursor: pointer;
}

.symbol-switch em {
  width: 10px;
  height: 10px;
  transform: rotate(45deg) translateY(-2px);
  border-right: 2px solid currentColor;
  border-bottom: 2px solid currentColor;
}

.star-button {
  border: 0;
  background: transparent;
  color: #fff;
  font-size: 24px;
  line-height: 1;
  cursor: pointer;
}

.chart-summary {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(155px, 0.9fr);
  gap: 18px;
  margin-top: 20px;
}

.chart-summary > div:first-child {
  display: grid;
  gap: 8px;
}

.chart-summary__price {
  color: #0cd977;
  font-size: 24px;
  font-weight: 600;
  line-height: 1;
}

.chart-summary span {
  font-size: 15px;
  font-weight: 500;
}

.chart-summary small {
  color: #9ca0aa;
  font-size: 13px;
  font-weight: 400;
}

.chart-stats {
  display: grid;
  gap: 11px;
}

.chart-stats span {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  font-size: 13px;
}

.chart-stats em {
  color: #9ca0aa;
  font-style: normal;
}

.chart-stats strong {
  color: #fff;
  font-weight: 500;
}

.up {
  color: #0cd977 !important;
}

.down {
  color: #ff5959 !important;
}

.flat {
  color: #d7d9df !important;
}

.chart-sticky-tabs {
  position: sticky;
  top: var(--market-header-height, 68px);
  z-index: 19;
  margin-top: 26px;
  background: #0b0c15;
  border-bottom: 1px solid #242633;
  box-shadow: 0 8px 18px rgba(5, 6, 14, 0.5);
}

.sub-tabs {
  display: flex;
  gap: 28px;
  padding: 0 18px;
}

.sub-tab,
.interval-pill,
.tool-row button {
  border: 0;
  background: transparent;
  color: inherit;
  font: inherit;
  cursor: pointer;
}

.sub-tab {
  position: relative;
  padding: 0 0 15px;
  color: #8f929d;
  font-size: 17px;
  font-weight: 500;
}

.sub-tab--active {
  color: #fff;
  font-weight: 600;
}

.sub-tab--active::after {
  position: absolute;
  right: 3px;
  bottom: 0;
  left: 3px;
  height: 3px;
  border-radius: 999px 999px 0 0;
  background: #08c200;
  content: '';
}

.interval-row {
  display: flex;
  gap: 8px;
  overflow-x: auto;
  padding-top: 16px;
  padding-bottom: 15px;
}

.interval-pill {
  flex: 0 0 auto;
  min-width: 44px;
  padding: 7px 9px;
  border-radius: 999px;
  background: #191b25;
  color: #8d929d;
  font-size: 14px;
  font-weight: 500;
}

.interval-pill--active {
  background: #fff;
  color: #11131b;
}

.tool-row {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  border-top: 1px solid #242633;
  border-bottom: 1px solid #242633;
}

.tool-row button {
  min-height: 42px;
  border-right: 1px solid #242633;
  color: #8f929d;
  font-size: 14px;
  font-weight: 500;
}

.tool-row button:last-child {
  border-right: 0;
}

.chart-board {
  position: relative;
  display: grid;
  grid-template-columns: 58px minmax(0, 1fr) 70px;
  min-height: 560px;
  overflow: hidden;
  user-select: none;
}

.chart-tools {
  display: grid;
  align-content: start;
  gap: 26px;
  padding-top: 36px;
  border-right: 1px solid #242633;
}

.chart-tools span {
  width: 30px;
  height: 3px;
  margin: 0 auto;
  border-radius: 999px;
  background: #8f929d;
  box-shadow: 0 11px 0 rgba(143, 146, 157, 0.72);
}

.candle-chart {
  width: 100%;
  height: 560px;
}

.grid-line {
  stroke: #252836;
  stroke-dasharray: 4 4;
  stroke-width: 1;
}

.candle-up {
  fill: #0cd977;
  stroke: #0cd977;
}

.candle-down {
  fill: #ff574c;
  stroke: #ff574c;
}

.price-axis {
  display: grid;
  align-content: start;
  gap: 92px;
  padding: 76px 6px 0 8px;
  color: #8f929d;
  font-size: 12px;
  font-weight: 400;
}

.chart-loading {
  position: absolute;
  top: 12px;
  left: 50%;
  padding: 5px 10px;
  transform: translateX(-50%);
  border-radius: 999px;
  background: rgba(25, 27, 37, 0.92);
  color: #d7d9df;
  font-size: 12px;
  font-weight: 500;
}

.volume-board {
  display: grid;
  grid-template-columns: 58px minmax(0, 1fr) 70px;
  min-height: 170px;
  padding-bottom: 120px;
  border-top: 1px solid #242633;
}

.volume-labels {
  grid-column: 2 / 4;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 14px 18px 0 16px;
  color: #8f929d;
  font-size: 13px;
}

.volume-labels strong {
  color: #ff574c;
  font-weight: 500;
}

.volume-bars {
  grid-column: 2 / 3;
  display: flex;
  align-items: end;
  gap: 5px;
  min-height: 112px;
  padding: 10px 8px 0 16px;
}

.volume-bars span {
  flex: 1 1 0;
  min-width: 4px;
}

.volume-up {
  background: #0cd977;
}

.volume-down {
  background: #ff574c;
}

.depth-board,
.trade-board {
  min-height: 620px;
  padding: 0 24px 120px;
}

.depth-board__head {
  display: grid;
  grid-template-columns: 1fr 1fr;
  padding: 10px 0 12px;
  border-bottom: 1px solid #242633;
  color: #8f929d;
  font-size: 14px;
  line-height: 1.25;
}

.depth-board__head span:last-child {
  text-align: right;
}

.depth-list {
  display: grid;
  gap: 6px;
  padding: 16px 0;
}

.depth-row {
  position: relative;
  display: grid;
  grid-template-columns: 1fr 1fr;
  align-items: center;
  min-height: 28px;
  overflow: hidden;
  font-size: 16px;
}

.depth-row i {
  position: absolute;
  top: 0;
  right: 0;
  bottom: 0;
  z-index: 0;
  opacity: 0.35;
}

.depth-row span,
.depth-row strong {
  position: relative;
  z-index: 1;
  font-weight: 500;
}

.depth-row strong {
  color: #f4f5f8;
  text-align: right;
}

.depth-row--ask span {
  color: #ff574c;
}

.depth-row--ask i {
  background: rgba(255, 74, 92, 0.28);
}

.depth-row--bid span {
  color: #0cd977;
}

.depth-row--bid i {
  background: rgba(12, 217, 119, 0.22);
}

.depth-mid {
  display: flex;
  align-items: baseline;
  gap: 16px;
  padding: 10px 0;
}

.depth-mid strong {
  font-size: 26px;
  font-weight: 600;
}

.depth-mid span {
  color: #8f929d;
  font-size: 18px;
}

.trade-board {
  display: grid;
  align-content: start;
  gap: 0;
  padding-top: 0;
}

.trade-board__head {
  display: grid;
  grid-template-columns: 1fr 1fr 88px;
  gap: 12px;
  padding: 14px 0 12px;
  color: #8f929d;
  font-size: 14px;
}

.trade-board__head span:nth-child(2),
.trade-board__head span:nth-child(3) {
  text-align: right;
}

.trade-row {
  display: grid;
  grid-template-columns: 1fr 1fr 88px;
  gap: 12px;
  align-items: center;
  min-height: 34px;
  color: #f4f5f8;
  font-size: 15px;
}

.trade-row span,
.trade-row strong,
.trade-row time {
  font-weight: 500;
}

.trade-row span {
  color: #0cd977;
}

.trade-row--down span {
  color: #ff574c;
}

.trade-row strong,
.trade-row time {
  text-align: right;
}

.detail-empty {
  display: grid;
  place-items: center;
  min-height: 240px;
  margin: 0;
  color: #8f929d;
  font-size: 14px;
}

.product-sheet-overlay {
  position: fixed;
  inset: 0;
  z-index: 80;
  display: flex;
  align-items: flex-end;
  justify-content: center;
  padding: 0 0 12px;
  background: rgba(3, 4, 10, 0.68);
  backdrop-filter: blur(7px);
}

.product-sheet {
  position: relative;
  width: min(100%, 640px);
  max-height: 68dvh;
  padding: 22px 22px 26px;
  overflow: hidden;
  border-radius: 28px 28px 8px 8px;
  background: #22232c;
  color: #f6f7fb;
  box-shadow: 0 -24px 70px rgba(0, 0, 0, 0.42);
}

.product-sheet__handle {
  display: block;
  width: 54px;
  height: 6px;
  margin: 0 auto 22px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.52);
}

.product-sheet__header {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 42px;
  margin-bottom: 10px;
}

.product-sheet__header h3 {
  margin: 0;
  color: #fff;
  font-size: 22px;
  font-weight: 500;
}

.product-sheet__header button {
  position: absolute;
  top: 42px;
  right: 24px;
  border: 0;
  background: transparent;
  color: #fff;
  font-size: 31px;
  line-height: 1;
  cursor: pointer;
}

.product-sheet__rows {
  min-height: 384px;
}

.product-sheet-row {
  display: grid;
  grid-template-columns: 54px minmax(100px, 1fr) minmax(88px, 0.78fr) minmax(82px, 0.7fr);
  align-items: center;
  width: 100%;
  min-height: 96px;
  padding: 12px 0;
  border: 0;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
  background: transparent;
  color: inherit;
  font: inherit;
  text-align: left;
  cursor: pointer;
}

.product-sheet-row__coin {
  display: grid;
  place-items: center;
  width: 44px;
  height: 44px;
  border-radius: 999px;
  background: linear-gradient(145deg, #4099ff, #67c2ff);
  color: #fff;
  font-size: 21px;
  font-weight: 500;
}

.product-sheet-row:nth-child(4n + 2) .product-sheet-row__coin {
  background: linear-gradient(145deg, #e9ddc9, #fff2d8);
  color: #b8346c;
}

.product-sheet-row:nth-child(4n + 3) .product-sheet-row__coin {
  background: linear-gradient(145deg, #2186dd, #2e9fff);
}

.product-sheet-row:nth-child(4n + 4) .product-sheet-row__coin {
  background: linear-gradient(145deg, #0e52ff, #3888ff);
}

.product-sheet-row__symbol {
  overflow: hidden;
  color: #fff;
  font-size: 18px;
  font-weight: 500;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.product-sheet-row strong {
  color: #09d676;
  font-size: 18px;
  font-weight: 500;
  text-align: right;
}

.product-sheet-row__change {
  display: grid;
  justify-items: end;
  gap: 6px;
}

.product-sheet-row__change em {
  color: #09d676;
  font-size: 17px;
  font-style: normal;
  font-weight: 500;
}

.product-sheet-row__change small {
  min-width: 76px;
  padding: 5px 9px;
  border-radius: 14px;
  background: #06d171;
  color: #fff;
  font-size: 14px;
  text-align: center;
}

.product-sheet-row--down strong,
.product-sheet-row--down .product-sheet-row__change em {
  color: #ff5148;
}

.product-sheet-row--down .product-sheet-row__change small {
  background: #ff4438;
}

.product-sheet-row--active {
  background: rgba(255, 255, 255, 0.025);
}

.product-sheet__footer {
  display: flex;
  justify-content: center;
  min-height: 26px;
  padding-top: 12px;
  color: #8f929d;
  font-size: 14px;
}
</style>
