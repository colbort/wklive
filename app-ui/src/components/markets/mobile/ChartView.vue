<script setup lang='ts'>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'

import { getLocale, useI18n } from '@/i18n'
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

const emit = defineEmits<{
  selectProduct: [productKey: string]
  selectInterval: [interval: Interval]
  loadPreviousPage: []
}>()

const swipeStartX = ref<number | null>(null)
const swipeStartY = ref<number | null>(null)
const switcherOpen = ref(false)
const timeSheetOpen = ref(false)
const activeDetailTab = ref<DetailTab>('market')
const { t } = useI18n()
const chartViewRef = ref<HTMLElement | null>(null)
const detailTabsSentinelRef = ref<HTMLElement | null>(null)
const detailTabsRef = ref<HTMLElement | null>(null)
const detailTabsPinned = ref(false)
let scrollContainer: HTMLElement | null = null
let pinRaf = 0
let detailTabsPinStart = 0

const minuteIntervals = computed(() => props.intervals.filter((item) => /m$/i.test(item.name)))
const majorIntervals = computed(() => props.intervals.filter((item) => !/m$/i.test(item.name)))
const selectedMinuteName = computed(() => {
  const minute = minuteIntervals.value.find((item) => item.name === props.selectedIntervalName)
  return minute ? minute.name : minuteIntervals.value[0]?.name || ''
})

const selectedProduct = computed(() => {
  return props.products.find((item) => productKey(item) === props.selectedProductKey) ?? null
})
const selectedDisplaySymbol = computed(() => {
  const product = selectedProduct.value
  if (!product) return '--'
  if (product.baseCoin && product.quoteCoin) return `${product.baseCoin}/${product.quoteCoin}`

  const quote = product.quoteCoin || 'USDT'
  if (product.symbol.toUpperCase().endsWith(quote.toUpperCase())) {
    return `${product.symbol.slice(0, -quote.length)}/${quote}`
  }

  return product.symbol
})
const selectedPriceChange = computed(() => getChangeRate(props.selectedQuote))
const selectedTrendClass = computed(() => trendClass(selectedPriceChange.value))
const selectedChangeValue = computed(() => {
  const quote = props.selectedQuote
  return quote ? quote.lastPrice - quote.open : 0
})

const chartCandles = computed(() => {
  const source = [...props.klineSnapshot].sort((left, right) => left.ts - right.ts).slice(-26)
  const prices = source.flatMap((item) => [item.open, item.close]).filter((item) => item > 0)
  const rawMax = Math.max(...prices, 1)
  const rawMin = Math.min(...prices, rawMax * 0.98)
  const padding = Math.max((rawMax - rawMin) * 0.08, rawMax * 0.00002)
  const max = rawMax + padding
  const min = Math.max(0, rawMin - padding)
  const range = Math.max(max - min, rawMax * 0.00004)
  const plotTop = 10
  const plotHeight = 210
  const toY = (value: number) => {
    const y = plotTop + ((max - value) / range) * plotHeight
    return Math.min(plotTop + plotHeight, Math.max(plotTop, y))
  }

  return source.map((item, index) => {
    const high = toY(Math.max(item.high, item.open, item.close))
    const low = toY(Math.min(item.low, item.open, item.close))
    const open = toY(item.open)
    const close = toY(item.close)
    const bodyTop = Math.min(open, close)
    const bodyHeight = Math.max(Math.abs(close - open), 7)

    return {
      key: `${item.ts}-${index}`,
      x: 14 + index * 19,
      high,
      low,
      bodyTop,
      bodyHeight,
      up: item.close >= item.open,
    }
  })
})

const chartLinePoints = computed(() => {
  if (!chartCandles.value.length) return ''

  return chartCandles.value
    .map((item) => `${item.x},${item.bodyTop + item.bodyHeight / 2}`)
    .join(' ')
})

const volumeBars = computed(() => {
  const source = [...props.klineSnapshot].sort((left, right) => left.ts - right.ts).slice(-26)
  const maxVolume = Math.max(...source.map((item) => item.volume || 0), 1)

  return source.map((item, index) => ({
    key: `${item.ts}-${index}`,
    up: item.close >= item.open,
    height: 12 + ((item.volume || 0) / maxVolume) * 78,
  }))
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
    { label: t('market.high'), value: formatChartPrice(quote?.high), accent: true },
    { label: t('market.openToday'), value: formatChartPrice(quote?.open) },
    { label: t('market.turnover'), value: formatCompact(quote?.turnover) },
    { label: t('market.low'), value: formatChartPrice(quote?.low), accent: true },
    { label: t('market.prevClose'), value: formatChartPrice(quote?.open) },
    { label: t('market.volume'), value: formatCompact(quote?.volume) },
  ]
})
const askRows = computed(() => props.depthSnapshot?.asks ?? [])
const bidRows = computed(() => props.depthSnapshot?.bids ?? [])
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
  switcherOpen.value = true
}

function closeSwitcher() {
  switcherOpen.value = false
}

function selectProductRow(row: MarketRow) {
  emit('selectProduct', row.key)
  closeSwitcher()
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

function openTimeSheet() {
  timeSheetOpen.value = true
}

function closeTimeSheet() {
  timeSheetOpen.value = false
}

function selectTimeInterval(interval: Interval) {
  emit('selectInterval', interval)
  closeTimeSheet()
}

function updateDetailTabsPin() {
  pinRaf = 0

  if (!scrollContainer || !detailTabsPinStart) {
    detailTabsPinned.value = false
    return
  }

  detailTabsPinned.value = scrollContainer.scrollTop >= detailTabsPinStart
}

function requestDetailTabsPinUpdate() {
  if (pinRaf) return
  pinRaf = window.requestAnimationFrame(updateDetailTabsPin)
}

function refreshDetailTabsPin() {
  detailTabsPinStart = getDetailTabsPinStart()
  requestDetailTabsPinUpdate()
}

function bindChartScroll() {
  scrollContainer =
    (chartViewRef.value?.closest('.page-content') as HTMLElement | null) ||
    document.querySelector<HTMLElement>('.page-content')
  detailTabsPinStart = getDetailTabsPinStart()
  scrollContainer?.addEventListener('scroll', requestDetailTabsPinUpdate, { passive: true })
  window.addEventListener('resize', refreshDetailTabsPin, { passive: true })
  updateDetailTabsPin()
}

function getDetailTabsPinStart() {
  const sentinelRect = detailTabsSentinelRef.value?.getBoundingClientRect()
  const scrollRect = scrollContainer?.getBoundingClientRect()
  if (!sentinelRect || !scrollRect || !scrollContainer) return 0

  return Math.max(0, scrollContainer.scrollTop + sentinelRect.top - scrollRect.top)
}

onMounted(bindChartScroll)

onBeforeUnmount(() => {
  scrollContainer?.removeEventListener('scroll', requestDetailTabsPinUpdate)
  window.removeEventListener('resize', refreshDetailTabsPin)
  if (pinRaf) window.cancelAnimationFrame(pinRaf)
})

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

  return new Intl.NumberFormat(getLocale(), {
    minimumFractionDigits: digits,
    maximumFractionDigits: digits,
  }).format(value)
}

function formatPrice(value?: number | null) {
  if (value === null || value === undefined || !Number.isFinite(value)) return '--'

  return formatNumber(value, Math.abs(value) >= 1 ? 4 : 8)
}

function formatChartPrice(value?: number | null) {
  if (value === null || value === undefined || !Number.isFinite(value)) return '--'

  return formatNumber(value, Math.abs(value) >= 1 ? 2 : 8)
}

function formatCompact(value?: number | null) {
  if (value === null || value === undefined || !Number.isFinite(value)) return '--'

  return new Intl.NumberFormat(getLocale(), {
    notation: 'compact',
    maximumFractionDigits: 2,
  }).format(value)
}

function formatPercent(value: number) {
  return `${value >= 0 ? '+' : ''}${value.toFixed(2)}%`
}

function formatTime(ts: number) {
  if (!ts) return '--'

  return new Intl.DateTimeFormat(getLocale(), {
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
  <section ref="chartViewRef" class="chart-view" :class="{ 'chart-view--tabs-pinned': detailTabsPinned }">
    <div class="chart-switcher">
      <button type="button" class="symbol-switch" @click="openSwitcher">
        <span>{{ selectedDisplaySymbol }}</span>
        <em />
      </button>

      <button type="button" class="star-button" :aria-label="t('market.addWatchlist')">☆</button>
    </div>

    <div class="chart-summary">
      <div>
        <strong class="chart-summary__price" :class="selectedTrendClass">
          {{ selectedQuote ? formatChartPrice(selectedQuote.lastPrice) : '--' }}
        </strong>
        <span :class="selectedTrendClass">
          {{ selectedQuote ? `${selectedChangeValue >= 0 ? '+' : ''}${formatChartPrice(selectedChangeValue)}  ${formatPercent(selectedPriceChange)}` : '--' }}
        </span>
      </div>

      <div class="chart-stats">
        <span v-for="item in chartStats" :key="item.label">
          <em>{{ item.label }}</em>
          <strong :class="{ up: item.accent }">{{ item.value }}</strong>
        </span>
      </div>
    </div>

    <span ref="detailTabsSentinelRef" class="chart-tabs-sentinel" aria-hidden="true" />
    <div ref="detailTabsRef" class="chart-sticky-tabs" :class="{ 'chart-sticky-tabs--pinned': detailTabsPinned }">
      <div class="sub-tabs">
        <button
          type="button"
          class="sub-tab"
          :class="{ 'sub-tab--active': activeDetailTab === 'market' }"
          @click="activeDetailTab = 'market'"
        >
          {{ t('market.quotes') }}
        </button>
        <button
          type="button"
          class="sub-tab"
          :class="{ 'sub-tab--active': activeDetailTab === 'depth' }"
          @click="activeDetailTab = 'depth'"
        >
          {{ t('market.orderbook') }}
        </button>
        <button
          type="button"
          class="sub-tab"
          :class="{ 'sub-tab--active': activeDetailTab === 'trades' }"
          @click="activeDetailTab = 'trades'"
        >
          {{ t('market.latestTrades') }}
        </button>
      </div>
    </div>

    <template v-if="activeDetailTab === 'market'">
      <div class="interval-row">
        <span class="time-selector" aria-hidden="true">
          <span class="time-selector__label">Time</span>
        </span>

        <button
          v-if="selectedMinuteName"
          type="button"
          class="interval-pill"
          :class="{ 'interval-pill--active': selectedIntervalName === selectedMinuteName }"
          @click="openTimeSheet"
        >
          {{ selectedMinuteName }}
          <span class="interval-pill__arrow">▾</span>
        </button>

        <button
          v-for="item in majorIntervals"
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
        <button type="button" class="tool-button">
          <span class="tool-button__icon">ƒx</span>
          <span class="tool-button__label">{{ t('market.indicator') }}</span>
        </button>
        <button type="button" class="tool-button">
          <span class="tool-button__icon">◉</span>
          <span class="tool-button__label">{{ t('market.timezone') }}</span>
        </button>
        <button type="button" class="tool-button">
          <span class="tool-button__icon">⚙</span>
          <span class="tool-button__label">{{ t('market.settings') }}</span>
        </button>
      </div>

      <div
        class="chart-board"
        @touchstart.passive="handlePointerStart"
        @touchend.passive="handlePointerEnd"
        @mousedown="handlePointerStart"
        @mouseup="handlePointerEnd"
      >
        <div class="chart-tools" aria-hidden="true">
          <span class="chart-tool chart-tool--line" />
          <span class="chart-tool chart-tool--trend" />
          <span class="chart-tool chart-tool--circle" />
          <span class="chart-tool chart-tool--rays" />
          <span class="chart-tool chart-tool--mesh" />
          <span class="chart-tool chart-tool--magnet" />
        </div>

        <div class="ma-legend" aria-hidden="true">
          <span>MA(5,10,30,60)</span>
          <button type="button">◉</button>
          <button type="button">⚙</button>
          <button type="button">×</button>
          <strong class="ma-legend__ma5">MA5: 75,900.94</strong>
          <strong class="ma-legend__ma10">MA10: 75,909.73</strong>
          <strong class="ma-legend__ma30">MA30: 76,533.68</strong>
          <strong class="ma-legend__ma60">MA60: 76,869.70</strong>
        </div>

        <svg class="candle-chart" viewBox="0 0 520 240" role="img" :aria-label="t('market.candleChart')">
          <line x1="0" y1="58" x2="520" y2="58" class="grid-line" />
          <line x1="0" y1="120" x2="520" y2="120" class="grid-line" />
          <line x1="0" y1="182" x2="520" y2="182" class="grid-line" />
          <line x1="170" y1="0" x2="170" y2="240" class="grid-line" />
          <line x1="340" y1="0" x2="340" y2="240" class="grid-line" />

          <polyline v-if="chartLinePoints" :points="chartLinePoints" class="ma-line ma-line--yellow" />
          <polyline v-if="chartLinePoints" :points="chartLinePoints" class="ma-line ma-line--blue" transform="translate(0, 12)" />
          <polyline v-if="chartLinePoints" :points="chartLinePoints" class="ma-line ma-line--pink" transform="translate(0, -20)" />

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
          <span v-for="mark in chartPriceMarks" :key="mark">{{ formatPrice(mark) }}</span>
        </div>

        <div v-if="loadingKline" class="chart-loading">{{ t('common.loading') }}...</div>
      </div>

      <div class="volume-board" aria-hidden="true">
        <div class="volume-tools" />
        <div class="volume-labels">
          <span>VOL(5,10,20)</span>
          <em>MA10: 181.723K</em>
          <em>MA20: 229.212K</em>
          <strong>VOLUME</strong>
        </div>
        <div class="volume-bars">
          <span
            v-for="bar in volumeBars"
            :key="`volume-${bar.key}`"
            :class="bar.up ? 'volume-up' : 'volume-down'"
            :style="{ height: `${bar.height}px` }"
          />
        </div>
      </div>
    </template>

    <section v-else-if="activeDetailTab === 'depth'" class="depth-board">
      <header class="depth-board__head">
        <span>{{ t('market.price') }}<br />({{ selectedProduct?.quoteCoin || 'USDT' }})</span>
        <span>{{ t('market.qty') }}<br />({{ selectedProduct?.baseCoin || selectedProduct?.symbol || '--' }})</span>
      </header>

      <div class="depth-list depth-list--asks">
        <div
          v-for="(item, index) in askRows"
          :key="`ask-${item.price}-${index}`"
          class="depth-row depth-row--ask"
        >
          <i :style="{ width: `${Math.max(8, (item.volume / maxAskVolume) * 100)}%` }" />
          <span>{{ formatPrice(item.price) }}</span>
          <strong>{{ formatNumber(item.volume, 8) }}</strong>
        </div>
      </div>

      <div class="depth-mid" :class="selectedTrendClass">
        <strong>{{ selectedQuote ? formatPrice(selectedQuote.lastPrice) : '--' }}</strong>
        <span>{{ formatPrice(selectedQuote?.open) }}</span>
      </div>

      <div class="depth-list depth-list--bids">
        <div
          v-for="(item, index) in bidRows"
          :key="`bid-${item.price}-${index}`"
          class="depth-row depth-row--bid"
        >
          <i :style="{ width: `${Math.max(8, (item.volume / maxBidVolume) * 100)}%` }" />
          <span>{{ formatPrice(item.price) }}</span>
          <strong>{{ formatNumber(item.volume, 8) }}</strong>
        </div>
      </div>

      <p v-if="!askRows.length && !bidRows.length" class="detail-empty">{{ t('market.waitingOrderbook') }}</p>
    </section>

    <section v-else class="trade-board">
      <header class="trade-board__head">
        <span>{{ t('market.price') }}({{ selectedProduct?.quoteCoin || 'USDT' }})</span>
        <span>{{ t('market.qty') }}({{ selectedProduct?.baseCoin || selectedProduct?.symbol || '--' }})</span>
        <span>{{ t('trade.time') }}</span>
      </header>

      <div v-if="!tradeRows.length" class="detail-empty">{{ t('market.waitingTrades') }}</div>
      <div
        v-for="(item, index) in tradeRows"
        v-else
        :key="`${item.ts}-${index}`"
        class="trade-row"
        :class="{ 'trade-row--down': item.direction === 'down' }"
      >
        <span>{{ formatPrice(item.lastPrice) }}</span>
        <strong>{{ formatNumber(item.volume, 8) }}</strong>
        <time>{{ formatTime(item.ts) }}</time>
      </div>
    </section>

    <div v-if="switcherOpen" class="product-sheet-overlay" @click.self="closeSwitcher">
      <section
        class="product-sheet"
      >
        <span class="product-sheet__handle" />

        <header class="product-sheet__header">
          <h3>{{ categoryName || t('market.product') }}</h3>
          <button type="button" :aria-label="t('common.close')" @click="closeSwitcher">×</button>
        </header>

        <div class="product-sheet__rows">
          <button
            v-for="row in rows"
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
            <strong>{{ row.quote ? formatPrice(row.quote.lastPrice) : '--' }}</strong>
            <span class="product-sheet-row__change">
              <em>{{ row.quote ? formatPrice(row.quote.lastPrice - row.quote.open) : '--' }}</em>
              <small>{{ row.quote ? formatPercent(row.changeRate) : t('market.waiting') }}</small>
            </span>
          </button>
        </div>

        <div class="product-sheet__footer">
          <span>{{ t('market.productCount', { count: rows.length }) }}</span>
        </div>
      </section>
    </div>

    <div v-if="timeSheetOpen" class="product-sheet-overlay" @click.self="closeTimeSheet">
      <section class="time-sheet">
        <span class="time-sheet__handle" />
        <div class="time-sheet__rows">
          <button
            v-for="item in minuteIntervals"
            :key="item.name"
            type="button"
            class="time-sheet__item"
            :class="{ 'time-sheet__item--active': item.name === selectedIntervalName }"
            @click="selectTimeInterval(item)"
          >
            {{ item.name }}
          </button>
        </div>
      </section>
    </div>
  </section>
</template>

<style scoped>
.chart-view {
  padding: 0;
  padding-bottom: calc(92px + env(safe-area-inset-bottom));
  overflow-x: clip;
}

.chart-view--tabs-pinned {
  padding-top: 58px;
}

.chart-tabs-sentinel {
  display: block;
  height: 0;
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
  padding-top: 10px;
  padding-bottom: 8px;
  background: #0b0c15;
}

.symbol-switch {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  min-width: 140px;
  border: 0;
  background: transparent;
  color: #fff;
  font: inherit;
  font-size: 19px;
  font-weight: 500;
  cursor: pointer;
}

.symbol-switch em {
  width: 14px;
  height: 14px;
  transform: rotate(45deg) translateY(-3px);
  border-right: 3px solid currentColor;
  border-bottom: 3px solid currentColor;
}

.star-button {
  border: 0;
  background: transparent;
  color: #fff;
  font-size: 30px;
  font-weight: 300;
  line-height: 1;
  cursor: pointer;
}

.chart-summary {
  display: grid;
  grid-template-columns: minmax(0, 0.88fr) minmax(0, 1fr);
  gap: 18px;
  padding: 14px 18px 16px;
  min-height: 96px;
  overflow: visible;
}

.chart-summary > div:first-child {
  display: grid;
  align-content: center;
  gap: 10px;
  min-width: 0;
}

.chart-summary__price {
  overflow: hidden;
  color: #0cd977;
  font-size: 30px;
  font-weight: 500;
  line-height: 1;
  display: block;
  letter-spacing: 0;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.chart-summary span {
  overflow: hidden;
  font-size: 17px;
  font-weight: 500;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.chart-stats {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px 10px;
  align-content: center;
  min-width: 0;
}

.chart-stats span {
  display: grid;
  gap: 5px;
  min-width: 0;
  font-size: 14px;
}

.chart-stats em {
  color: #9ca0aa;
  font-size: 12px;
  font-weight: 400;
  line-height: 1;
  white-space: nowrap;
}

.chart-stats strong {
  overflow: hidden;
  color: #9ca0aa;
  font-size: 13px;
  font-weight: 500;
  line-height: 1.08;
  text-overflow: ellipsis;
  white-space: nowrap;
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
  z-index: 60;
  background: #0b0c15;
  border-bottom: 1px solid #20222d;
  box-shadow: 0 8px 18px rgba(5, 6, 14, 0.36);
}

.chart-sticky-tabs--pinned {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 90;
  width: 100%;
  box-sizing: border-box;
}

.sub-tabs {
  display: flex;
  gap: 42px;
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
  padding: 12px 0 13px;
  color: #8f929d;
  font-size: 19px;
  font-weight: 500;
}

.sub-tab--active {
  color: #fff;
  font-weight: 500;
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
  flex-wrap: nowrap;
  align-items: center;
  gap: 8px;
  overflow-x: auto;
  padding-top: 8px;
  padding-bottom: 8px;
  border-bottom: 1px solid #20222d;
  scrollbar-width: none;
}

.interval-row::-webkit-scrollbar {
  display: none;
}

.interval-pill {
  flex: 0 0 auto;
  min-width: 40px;
  padding: 0 12px;
  height: 34px;
  line-height: 34px;
  border-radius: 999px;
  background: #161923;
  color: #8d929d;
  font-size: 14px;
  font-weight: 400;
  white-space: nowrap;
}

.interval-pill--time {
  color: #ffffff;
  background: rgba(255, 255, 255, 0.08);
}

.interval-pill--active {
  background: #fff;
  color: #11131b;
}

.tool-row {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  padding-right: 0;
  padding-left: 0;
  border-bottom: 1px solid #252733;
}

.tool-row button {
  min-height: 42px;
  display: flex;
  gap: 8px;
  justify-content: center;
  align-items: center;
  border-right: 1px solid #252733;
  background: #0b0c15;
  color: #8f929d;
  font-size: 14px;
  font-weight: 400;
  padding: 0;
}

.tool-row button:last-child {
  border-right: 0;
}

.tool-button__icon {
  font-size: 18px;
  line-height: 1;
}

.tool-button__label {
  font-size: 14px;
}

.time-selector {
  display: inline-flex;
  align-items: center;
  flex: 0 0 auto;
  padding: 0 14px;
  min-height: 34px;
  border-radius: 999px;
  background: #161923;
  color: #8f929d;
  font-size: 14px;
  font-weight: 400;
}

.time-selector__label {
  white-space: nowrap;
}

.interval-pill__arrow {
  margin-left: 6px;
  font-size: 16px;
  color: currentColor;
}

.time-sheet {
  position: relative;
  display: flex;
  flex-direction: column;
  width: min(100%, 640px);
  max-height: 68dvh;
  padding: 14px 16px 30px;
  overflow: hidden;
  border-radius: 28px 28px 0 0;
  background: #16171f;
  color: #f6f7fb;
  box-shadow: 0 -24px 70px rgba(0, 0, 0, 0.42);
  touch-action: pan-y;
  max-width: 100%;
}

.time-sheet__handle {
  display: block;
  width: 54px;
  height: 6px;
  margin: 0 auto 16px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.18);
}

.time-sheet__rows {
  display: grid;
  gap: 12px;
  overflow: hidden;
}

.time-sheet__item {
  width: 100%;
  min-height: 56px;
  border-radius: 16px;
  padding: 16px 20px;
  border: 1px solid transparent;
  background: #1b1d27;
  color: #d3d7df;
  text-align: center;
  font-size: 16px;
  font-weight: 500;
}

.time-sheet__item--active {
  background: #10131d;
  color: #08c200;
  border-color: rgba(8, 194, 0, 0.18);
}

.chart-board {
  position: relative;
  display: grid;
  grid-template-columns: 58px minmax(0, 1fr) 86px;
  min-height: 500px;
  overflow: hidden;
  border-bottom: 1px solid #20222d;
  user-select: none;
}

.chart-tools {
  display: grid;
  align-content: start;
  gap: 24px;
  padding-top: 34px;
  border-right: 1px solid #242633;
}

.chart-tool {
  position: relative;
  display: block;
  width: 34px;
  height: 34px;
  margin: 0 auto;
  color: #8f929d;
}

.chart-tool--line::before,
.chart-tool--rays::before,
.chart-tool--rays::after {
  position: absolute;
  left: 3px;
  right: 3px;
  height: 3px;
  border-radius: 999px;
  background: currentColor;
  content: '';
}

.chart-tool--line::before {
  top: 16px;
  box-shadow: 0 0 0 4px #0b0c15, 0 0 0 5px currentColor;
}

.chart-tool--trend::before {
  position: absolute;
  inset: 7px 4px;
  transform: rotate(-42deg);
  border-top: 3px solid currentColor;
  border-bottom: 3px solid currentColor;
  content: '';
}

.chart-tool--circle {
  border: 3px solid currentColor;
  border-radius: 999px;
}

.chart-tool--circle::after {
  position: absolute;
  top: 11px;
  left: 11px;
  width: 8px;
  height: 8px;
  border-radius: 999px;
  background: currentColor;
  box-shadow: 20px 2px 0 -1px currentColor;
  content: '';
}

.chart-tool--rays::before {
  top: 8px;
  box-shadow: 0 10px 0 currentColor, 0 20px 0 currentColor;
}

.chart-tool--mesh::before {
  position: absolute;
  inset: 5px;
  transform: rotate(-8deg);
  border: 3px solid currentColor;
  content: '';
}

.chart-tool--mesh::after {
  position: absolute;
  inset: 9px 3px;
  transform: rotate(34deg);
  border-top: 3px solid currentColor;
  border-bottom: 3px solid currentColor;
  content: '';
}

.chart-tool--magnet::before {
  position: absolute;
  inset: 5px;
  border: 4px solid currentColor;
  border-top-color: transparent;
  border-radius: 0 0 999px 999px;
  content: '';
}

.ma-legend {
  position: absolute;
  top: 16px;
  left: 82px;
  right: 92px;
  z-index: 2;
  display: flex;
  flex-wrap: wrap;
  gap: 6px 12px;
  align-items: center;
  color: #8f929d;
  font-size: 12px;
  font-weight: 400;
  line-height: 1.2;
  pointer-events: none;
}

.ma-legend button {
  border: 0;
  background: transparent;
  color: #8f929d;
  font: inherit;
  padding: 0;
}

.ma-legend strong {
  font-weight: 500;
}

.ma-legend__ma5 {
  color: #ffad16;
}

.ma-legend__ma10 {
  color: #b16adc;
}

.ma-legend__ma30 {
  color: #1aa9ff;
}

.ma-legend__ma60 {
  color: #ff1687;
}

.candle-chart {
  width: 100%;
  height: 500px;
  margin-top: 24px;
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

.ma-line {
  fill: none;
  stroke-width: 2;
  opacity: 0.96;
}

.ma-line--yellow {
  stroke: #ffad16;
}

.ma-line--blue {
  stroke: #1aa9ff;
}

.ma-line--pink {
  stroke: #ff1687;
}

.price-axis {
  display: grid;
  align-content: start;
  gap: 82px;
  padding: 58px 6px 0 8px;
  color: #8f929d;
  font-size: 12px;
  font-weight: 500;
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
  grid-template-columns: 58px minmax(0, 1fr) 86px;
  min-height: 150px;
  padding-bottom: 18px;
  border-top: 1px solid #242633;
}

.volume-tools {
  grid-row: 1 / 3;
  border-right: 1px solid #242633;
}

.volume-labels {
  grid-column: 2 / 4;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px 18px;
  padding: 12px 18px 0 16px;
  color: #8f929d;
  font-size: 13px;
}

.volume-labels strong {
  color: #ff574c;
  font-weight: 500;
  margin-left: auto;
}

.volume-labels em {
  color: #b16adc;
  font-style: normal;
}

.volume-bars {
  grid-column: 2 / 3;
  display: flex;
  align-items: end;
  gap: 5px;
  min-height: 86px;
  padding: 8px 8px 12px 16px;
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
  font-weight: 500;
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
  padding: 0;
  background: rgba(3, 4, 10, 0.68);
  backdrop-filter: blur(7px);
}

.product-sheet {
  position: relative;
  display: flex;
  flex-direction: column;
  width: min(100%, 640px);
  max-height: 68dvh;
  padding: 22px 22px 26px;
  overflow: hidden;
  border-radius: 28px 28px 0 0;
  background: #22232c;
  color: #f6f7fb;
  box-shadow: 0 -24px 70px rgba(0, 0, 0, 0.42);
  touch-action: pan-y;
  max-width: 100%;
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
  flex: 1 1 auto;
  min-height: 0;
  overflow-y: auto;
  overflow-x: hidden;
  overscroll-behavior: contain;
  -webkit-overflow-scrolling: touch;
  touch-action: pan-y;
}

.product-sheet-row {
  display: grid;
  grid-template-columns: 44px minmax(0, 1fr) minmax(0, 72px) minmax(0, 64px);
  align-items: center;
  column-gap: 10px;
  width: 100%;
  min-width: 0;
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
  font-size: 17px;
  font-weight: 500;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.product-sheet-row strong {
  min-width: 0;
  overflow: hidden;
  color: #09d676;
  font-size: 16px;
  font-weight: 500;
  text-align: right;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.product-sheet-row__change {
  display: grid;
  justify-items: end;
  gap: 6px;
  min-width: 0;
  overflow: hidden;
}

.product-sheet-row__change em {
  max-width: 100%;
  overflow: hidden;
  color: #09d676;
  font-size: 15px;
  font-style: normal;
  font-weight: 500;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.product-sheet-row__change small {
  width: 100%;
  max-width: 100%;
  min-width: 0;
  padding: 5px 9px;
  overflow: hidden;
  border-radius: 14px;
  background: #06d171;
  color: #fff;
  font-size: 13px;
  text-align: center;
  text-overflow: ellipsis;
  white-space: nowrap;
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
