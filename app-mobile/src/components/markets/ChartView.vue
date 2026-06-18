<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import {
  dispose,
  init,
  type Chart,
  type DeepPartial,
  type KLineData,
  type Styles,
  LineType,
  TooltipShowRule,
  TooltipShowType,
} from 'klinecharts'

import BottomDrawer from '@/components/common/BottomDrawer.vue'
import { getLocale, useI18n } from '@/i18n'
import type { Interval } from '@/types/core'
import type {
  DepthPayload,
  ItickTenantProduct,
  KlinePayload,
  QuotePayload,
  TickPayload,
} from '@/types/itick'
import QuoteRow from './QuoteRow.vue'
import type { MarketRow } from './types'

type DetailTab = 'market' | 'depth' | 'trades'

const props = defineProps<{
  products: ItickTenantProduct[]
  rows: MarketRow[]
  categoryName: string
  selectedProductKey: string
  selectedQuote: QuotePayload | null
  klineSnapshot: KlinePayload[]
  viewingLatestKlinePage: boolean
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
const chartHostRef = ref<HTMLElement | null>(null)
const detailTabsSentinelRef = ref<HTMLElement | null>(null)
const detailTabsRef = ref<HTMLElement | null>(null)
const detailTabsPinned = ref(false)
let chart: Chart | null = null
let chartResizeObserver: ResizeObserver | null = null
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

const klineChartData = computed<KLineData[]>(() =>
  [...props.klineSnapshot]
    .filter((item) => item.ts && item.open > 0 && item.high > 0 && item.low > 0 && item.close > 0)
    .sort((left, right) => left.ts - right.ts)
    .map((item) => ({
      timestamp: normalizeChartTimestamp(item.ts),
      open: item.open,
      high: item.high,
      low: item.low,
      close: item.close,
      volume: item.volume,
      turnover: item.turnover,
    })),
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

function initKlineChart() {
  if (!chartHostRef.value || chart) return

  chart = init(chartHostRef.value, {
    locale: getLocale(),
    styles: chartStyles,
  })

  if (!chart) return

  chart.setPriceVolumePrecision(2, 3)
  chart.setBarSpace(9)
  chart.setOffsetRightDistance(8)
  chart.createIndicator('MA', true, { id: 'candle_pane' })
  chart.createIndicator('VOL', false, { id: 'volume_pane', height: 118, minHeight: 92 })
  chart.createIndicator('EMA', false, { id: 'ema_pane', height: 126, minHeight: 96 })
  syncKlineChartData()

  chartResizeObserver = new ResizeObserver(() => chart?.resize())
  chartResizeObserver.observe(chartHostRef.value)
}

function syncKlineChartData() {
  if (!chart) return

  if (!props.viewingLatestKlinePage) {
    const currentData = chart.getDataList()
    const firstTimestamp = currentData[0]?.timestamp ?? 0
    const previousData = firstTimestamp
      ? klineChartData.value.filter((item) => item.timestamp < firstTimestamp)
      : klineChartData.value

    if (previousData.length) {
      chart.applyMoreData(previousData, true)
    }
    return
  }

  chart.applyNewData(klineChartData.value)
  if (props.viewingLatestKlinePage && klineChartData.value.length) {
    chart.scrollToRealTime()
  }
}

function normalizeChartTimestamp(ts: number) {
  return ts > 0 && ts < 10_000_000_000 ? ts * 1000 : ts
}

function getDetailTabsPinStart() {
  const sentinelRect = detailTabsSentinelRef.value?.getBoundingClientRect()
  const scrollRect = scrollContainer?.getBoundingClientRect()
  if (!sentinelRect || !scrollRect || !scrollContainer) return 0

  return Math.max(0, scrollContainer.scrollTop + sentinelRect.top - scrollRect.top)
}

onMounted(() => {
  bindChartScroll()
  nextTick(initKlineChart)
})

onBeforeUnmount(() => {
  scrollContainer?.removeEventListener('scroll', requestDetailTabsPinUpdate)
  window.removeEventListener('resize', refreshDetailTabsPin)
  if (pinRaf) window.cancelAnimationFrame(pinRaf)
  chartResizeObserver?.disconnect()
  chartResizeObserver = null
  if (chart) {
    dispose(chart)
    chart = null
  }
})

watch(klineChartData, syncKlineChartData)

watch(
  () => getLocale(),
  (locale) => chart?.setLocale(locale),
)

const chartStyles: DeepPartial<Styles> = {
  grid: {
    horizontal: {
      color: '#242837',
      size: 1,
      style: LineType.Dashed,
      dashedValue: [4, 4],
    },
    vertical: {
      color: '#242837',
      size: 1,
      style: LineType.Dashed,
      dashedValue: [4, 4],
    },
  },
  candle: {
    bar: {
      upColor: '#08d88d',
      downColor: '#ff4f3f',
      noChangeColor: '#8b8e99',
      upBorderColor: '#08d88d',
      downBorderColor: '#ff4f3f',
      noChangeBorderColor: '#8b8e99',
      upWickColor: '#08d88d',
      downWickColor: '#ff4f3f',
      noChangeWickColor: '#8b8e99',
    },
    priceMark: {
      high: {
        color: '#f4f7ff',
      },
      low: {
        color: '#f4f7ff',
      },
      last: {
        upColor: '#08d88d',
        downColor: '#ff4f3f',
        noChangeColor: '#8b8e99',
        line: {
          show: true,
          style: LineType.Dashed,
          dashedValue: [4, 4],
          size: 1,
        },
        text: {
          show: true,
          color: '#ffffff',
          size: 12,
          paddingLeft: 4,
          paddingRight: 4,
          paddingTop: 2,
          paddingBottom: 2,
        },
      },
    },
    tooltip: {
      showRule: TooltipShowRule.FollowCross,
      showType: TooltipShowType.Standard,
      text: {
        color: '#d8dbe6',
        size: 11,
      },
    },
  },
  indicator: {
    tooltip: {
      showRule: TooltipShowRule.Always,
      showType: TooltipShowType.Standard,
      text: {
        color: '#a0a4af',
        size: 11,
      },
    },
    lines: [
      { color: '#ffad16', size: 1 },
      { color: '#8f5fd0', size: 1 },
      { color: '#1aa9ff', size: 1 },
      { color: '#ff1687', size: 1 },
    ],
    bars: [
      {
        upColor: '#08d88d',
        downColor: '#ff4f3f',
        noChangeColor: '#8b8e99',
      },
    ],
  },
  xAxis: {
    axisLine: {
      show: false,
    },
    tickText: {
      color: '#8b8e99',
      size: 11,
    },
  },
  yAxis: {
    axisLine: {
      show: false,
    },
    tickText: {
      color: '#8b8e99',
      size: 11,
    },
  },
  separator: {
    color: '#2b2f3f',
    size: 1,
  },
  crosshair: {
    horizontal: {
      line: {
        color: '#7f8491',
        size: 1,
        style: LineType.Dashed,
        dashedValue: [4, 4],
      },
      text: {
        backgroundColor: '#2b3141',
      },
    },
    vertical: {
      line: {
        color: '#7f8491',
        size: 1,
        style: LineType.Dashed,
        dashedValue: [4, 4],
      },
      text: {
        backgroundColor: '#2b3141',
      },
    },
  },
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

</script>

<template>
  <section
    ref="chartViewRef"
    class="chart-view"
    :class="{ 'chart-view--tabs-pinned': detailTabsPinned }"
  >
    <div class="chart-switcher">
      <button type="button" class="symbol-switch" @click="openSwitcher">
        <span>{{ selectedDisplaySymbol }}</span>
        <em />
      </button>

      <button type="button" class="star-button" :aria-label="t('market.addWatchlist')">
        <AppIcon name="star" class="star-icon-svg" />
      </button>
    </div>

    <div class="chart-summary">
      <div>
        <strong class="chart-summary__price" :class="selectedTrendClass">
          {{ selectedQuote ? formatChartPrice(selectedQuote.lastPrice) : '--' }}
        </strong>
        <span :class="selectedTrendClass">
          {{
            selectedQuote
              ? `${selectedChangeValue >= 0 ? '+' : ''}${formatChartPrice(selectedChangeValue)}  ${formatPercent(selectedPriceChange)}`
              : '--'
          }}
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
    <div
      ref="detailTabsRef"
      class="chart-sticky-tabs"
      :class="{ 'chart-sticky-tabs--pinned': detailTabsPinned }"
    >
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

      <div class="chart-board">
        <div class="chart-tools" aria-hidden="true">
          <span class="chart-tool chart-tool--line" />
          <span class="chart-tool chart-tool--trend" />
          <span class="chart-tool chart-tool--circle" />
          <span class="chart-tool chart-tool--rays" />
          <span class="chart-tool chart-tool--mesh" />
          <span class="chart-tool chart-tool--magnet" />
        </div>

        <div
          ref="chartHostRef"
          class="kline-chart-host"
          role="img"
          :aria-label="t('market.candleChart')"
          @touchstart.passive="handlePointerStart"
          @touchend.passive="handlePointerEnd"
          @mousedown="handlePointerStart"
          @mouseup="handlePointerEnd"
        >
          <div v-if="!klineChartData.length && !loadingKline" class="chart-empty">
            {{ t('market.waitingKline') }}
          </div>
        </div>

        <div v-if="loadingKline" class="chart-loading">
          {{ t('common.loading') }}...
        </div>
      </div>
    </template>

    <section v-else-if="activeDetailTab === 'depth'" class="depth-board">
      <header class="depth-board__head">
        <span>{{ t('market.price') }}<br>({{ selectedProduct?.quoteCoin || 'USDT' }})</span>
        <span>{{ t('market.qty') }}<br>({{
          selectedProduct?.baseCoin || selectedProduct?.symbol || '--'
        }})</span>
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

      <p v-if="!askRows.length && !bidRows.length" class="detail-empty">
        {{ t('market.waitingOrderbook') }}
      </p>
    </section>

    <section v-else class="trade-board">
      <header class="trade-board__head">
        <span>{{ t('market.price') }}({{ selectedProduct?.quoteCoin || 'USDT' }})</span>
        <span>{{ t('market.qty') }}({{
          selectedProduct?.baseCoin || selectedProduct?.symbol || '--'
        }})</span>
        <span>{{ t('trade.time') }}</span>
      </header>

      <div v-if="!tradeRows.length" class="detail-empty">
        {{ t('market.waitingTrades') }}
      </div>
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

    <BottomDrawer
      v-model="switcherOpen"
      :title="categoryName || t('market.product')"
      max-height="68dvh"
      @close="closeSwitcher"
    >
      <div class="product-sheet__rows">
        <QuoteRow
          v-for="row in rows"
          :key="row.key"
          :product="row.product"
          :quote="row.quote"
          :change-rate="row.changeRate"
          :direction="row.direction"
          :active="row.key === selectedProductKey"
          variant="sheet"
          @select="selectProductRow(row)"
        />
      </div>

      <template #footer>
        <div class="product-sheet__footer">
          <span>{{ t('market.productCount', { count: rows.length }) }}</span>
        </div>
      </template>
    </BottomDrawer>

    <BottomDrawer
      v-model="timeSheetOpen"
      :show-close="false"
      max-height="68dvh"
      @close="closeTimeSheet"
    >
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
    </BottomDrawer>
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
  background: var(--page-bg);
}

.symbol-switch {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  min-width: 140px;
  border: 0;
  background: transparent;
  color: var(--text);
  font: inherit;
  font-size: 0.95rem;
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
  color: var(--text);
  font-size: 1.5rem;
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
  color: var(--success);
  font-size: 1.5rem;
  font-weight: 500;
  line-height: 1;
  display: block;
  letter-spacing: 0;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.chart-summary span {
  overflow: hidden;
  font-size: 0.85rem;
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
  font-size: 0.7rem;
}

.chart-stats em {
  color: var(--muted);
  font-size: 0.6rem;
  font-weight: 400;
  line-height: 1;
  white-space: nowrap;
}

.chart-stats strong {
  overflow: hidden;
  color: var(--muted);
  font-size: 0.65rem;
  font-weight: 500;
  line-height: 1.08;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.up {
  color: var(--success) !important;
}

.down {
  color: #ff5959 !important;
}

.flat {
  color: var(--text) !important;
}

.chart-sticky-tabs {
  z-index: 60;
  background: var(--page-bg);
  border-bottom: 1px solid var(--divider);
  box-shadow: var(--shadow-floating);
}

.chart-sticky-tabs--pinned {
  position: fixed;
  top: 0;
  left: 50%;
  right: auto;
  z-index: 90;
  width: min(100%, var(--app-width, 100vw));
  box-sizing: border-box;
  transform: translateX(-50%);
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
  color: var(--muted);
  font-size: 0.95rem;
  font-weight: 500;
}

.sub-tab--active {
  color: var(--text);
  font-weight: 500;
}

.sub-tab--active::after {
  position: absolute;
  right: 3px;
  bottom: 0;
  left: 3px;
  height: 3px;
  border-radius: 999px 999px 0 0;
  background: var(--accent);
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
  border-bottom: 1px solid var(--divider);
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
  background: var(--panel-bg-deep);
  color: var(--muted);
  font-size: 0.7rem;
  font-weight: 400;
  white-space: nowrap;
}

.interval-pill--time {
  color: var(--text);
  background: var(--divider-soft);
}

.interval-pill--active {
  background: var(--text-strong);
  color: var(--text-dark);
}

.tool-row {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  padding-right: 0;
  padding-left: 0;
  border-bottom: 1px solid var(--control-bg-soft);
}

.tool-row button {
  min-height: 42px;
  display: flex;
  gap: 8px;
  justify-content: center;
  align-items: center;
  border-right: 1px solid var(--control-bg-soft);
  background: var(--page-bg);
  color: var(--muted);
  font-size: 0.7rem;
  font-weight: 400;
  padding: 0;
}

.tool-row button:last-child {
  border-right: 0;
}

.tool-button__icon {
  font-size: 0.9rem;
  line-height: 1;
}

.tool-button__label {
  font-size: 0.7rem;
}

.time-selector {
  display: inline-flex;
  align-items: center;
  flex: 0 0 auto;
  padding: 0 14px;
  min-height: 34px;
  border-radius: 999px;
  background: var(--panel-bg-deep);
  color: var(--muted);
  font-size: 0.7rem;
  font-weight: 400;
}

.time-selector__label {
  white-space: nowrap;
}

.interval-pill__arrow {
  margin-left: 6px;
  font-size: 0.8rem;
  color: currentColor;
}

.time-sheet__rows {
  display: grid;
  gap: 12px;
  padding: 0 0 12px;
}

.time-sheet__item {
  width: 100%;
  min-height: 56px;
  border-radius: 16px;
  padding: 16px 20px;
  border: 1px solid transparent;
  background: var(--panel-bg);
  color: var(--text-soft);
  text-align: center;
  font-size: 0.8rem;
  font-weight: 500;
}

.time-sheet__item--active {
  background: var(--page-bg-soft);
  color: var(--accent);
  border-color: var(--accent-border-soft);
}

.chart-board {
  position: relative;
  display: grid;
  grid-template-columns: 58px minmax(0, 1fr);
  min-height: 760px;
  overflow: hidden;
  border-bottom: 1px solid var(--divider);
  user-select: none;
}

.chart-tools {
  display: grid;
  align-content: start;
  gap: 24px;
  padding-top: 34px;
  border-right: 1px solid var(--border-soft);
}

.chart-tool {
  position: relative;
  display: block;
  width: 34px;
  height: 34px;
  margin: 0 auto;
  color: var(--muted);
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
  box-shadow:
    0 0 0 4px var(--page-bg),
    0 0 0 5px currentColor;
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
  box-shadow:
    0 10px 0 currentColor,
    0 20px 0 currentColor;
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

.kline-chart-host {
  position: relative;
  width: 100%;
  min-width: 0;
  height: 760px;
}

.chart-loading {
  position: absolute;
  top: 12px;
  left: 50%;
  padding: 5px 10px;
  transform: translateX(-50%);
  border-radius: 999px;
  background: var(--chart-toolbar-bg);
  color: var(--text);
  font-size: 0.6rem;
  font-weight: 500;
}

.chart-empty {
  position: absolute;
  inset: 0;
  display: grid;
  place-items: center;
  color: var(--muted);
  font-size: 0.72rem;
  pointer-events: none;
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
  border-bottom: 1px solid var(--border-soft);
  color: var(--muted);
  font-size: 0.7rem;
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
  font-size: 0.8rem;
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
  color: var(--text);
  text-align: right;
}

.depth-row--ask span {
  color: var(--danger-strong);
}

.depth-row--ask i {
  background: var(--danger-bg-emphasis);
}

.depth-row--bid span {
  color: var(--success);
}

.depth-row--bid i {
  background: var(--success-bg-emphasis);
}

.depth-mid {
  display: flex;
  align-items: baseline;
  gap: 16px;
  padding: 10px 0;
}

.depth-mid strong {
  font-size: 1.3rem;
  font-weight: 500;
}

.depth-mid span {
  color: var(--muted);
  font-size: 0.9rem;
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
  color: var(--muted);
  font-size: 0.7rem;
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
  color: var(--text);
  font-size: 0.75rem;
}

.trade-row span,
.trade-row strong,
.trade-row time {
  font-weight: 500;
}

.trade-row span {
  color: var(--success);
}

.trade-row--down span {
  color: var(--danger-strong);
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
  color: var(--muted);
  font-size: 0.7rem;
}

.product-sheet__rows {
  display: grid;
}

.product-sheet__footer {
  display: flex;
  justify-content: center;
  min-height: 26px;
  padding-top: 12px;
  color: var(--muted);
  font-size: 0.7rem;
}
</style>
