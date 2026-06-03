<script setup lang='ts'>
import { computed, ref, watch } from 'vue'
import { getLocale, useI18n } from '@/i18n'
import type { Interval } from '@/types/core'
import type { KlinePayload, QuotePayload } from '@/types/itick'

const props = defineProps<{
  selectedQuote: QuotePayload | null
  klineSnapshot: KlinePayload[]
  intervals: Interval[]
  selectedIntervalName: string
  loadingKline: boolean
}>()

const emit = defineEmits<{
  (e: 'select-interval', interval: Interval): void
}>()

const { t } = useI18n()
const CHART_VIEWBOX_WIDTH = 520
const CHART_VIEWBOX_HEIGHT = 240
const CHART_TOP = 12
const CHART_HEIGHT = 224
const CHART_LEFT = 14
const CHART_RIGHT = 18
const VOLUME_VIEWBOX_HEIGHT = 36
const MIN_VISIBLE_KLINES = 8

const minuteIntervals = computed(() => props.intervals.filter((item) => /m$/.test(item.name)))
const otherIntervals = computed(() => props.intervals.filter((item) => !/m$/.test(item.name)))
const selectedMinuteLabel = computed(() => {
  if (/m$/.test(props.selectedIntervalName)) return props.selectedIntervalName
  return minuteIntervals.value[0]?.name || '1m'
})

const sortedKlines = computed(() => [...props.klineSnapshot].sort((left, right) => left.ts - right.ts))
const windowSize = ref(26)
const windowOffset = ref(0)
const hoveredIndex = ref<number | null>(null)
const hoveredPrice = ref<number | null>(null)
const dragStartX = ref<number | null>(null)
const dragStartOffset = ref(0)
const hoverRatioX = ref(1)

const visibleKlines = computed(() => {
  const source = sortedKlines.value
  const size = Math.max(MIN_VISIBLE_KLINES, Math.min(windowSize.value, source.length || windowSize.value))
  const maxOffset = Math.max(0, source.length - size)
  const offset = Math.max(0, Math.min(windowOffset.value, maxOffset))
  const start = Math.max(0, source.length - size - offset)
  return source.slice(start, start + size)
})

watch(
  () => sortedKlines.value.length,
  (length) => {
    if (!length) {
      windowOffset.value = 0
      hoveredIndex.value = null
      hoveredPrice.value = null
      return
    }
    windowSize.value = Math.max(MIN_VISIBLE_KLINES, Math.min(windowSize.value, length))
    const maxOffset = Math.max(0, length - windowSize.value)
    windowOffset.value = Math.max(0, Math.min(windowOffset.value, maxOffset))
  },
  { immediate: true },
)

const candleGeometry = computed(() => {
  const count = Math.max(visibleKlines.value.length, 1)
  const innerWidth = CHART_VIEWBOX_WIDTH - CHART_LEFT - CHART_RIGHT
  const step = count > 1 ? innerWidth / (count - 1) : innerWidth
  const bodyWidth = Math.max(2, Math.min(24, step * 0.72))
  return { step, bodyWidth, innerWidth }
})

const verticalGridLines = computed(() => {
  const count = visibleKlines.value.length
  if (count <= 1) return []
  const slots = Math.min(4, Math.max(2, Math.floor(count / 6)))
  return Array.from({ length: slots - 1 }, (_, index) => {
    const ratio = (index + 1) / slots
    return CHART_LEFT + candleGeometry.value.innerWidth * ratio
  })
})

const chartDomain = computed(() => {
  const source = visibleKlines.value
  const prices = source.flatMap((item) => [item.high, item.low]).filter((item) => item > 0)
  const fallback = props.selectedQuote?.lastPrice || 1
  const rawMax = prices.length ? Math.max(...prices) : fallback
  const rawMin = prices.length ? Math.min(...prices) : fallback
  const padding = Math.max((rawMax - rawMin) * 0.12, rawMax * 0.002)
  const max = rawMax + padding
  const min = Math.max(0, rawMin - padding)
  const range = Math.max(max - min, rawMax * 0.01)

  return { min, max, range }
})

const chartCandles = computed(() => {
  const source = visibleKlines.value
  const { max, range } = chartDomain.value
  const { step } = candleGeometry.value

  return source.map((item, index) => {
    const high = CHART_TOP + ((max - item.high) / range) * CHART_HEIGHT
    const low = CHART_TOP + ((max - item.low) / range) * CHART_HEIGHT
    const open = CHART_TOP + ((max - item.open) / range) * CHART_HEIGHT
    const close = CHART_TOP + ((max - item.close) / range) * CHART_HEIGHT
    const bodyTop = Math.min(open, close)
    const bodyHeight = Math.max(Math.abs(close - open), 4)

    return {
      key: `${item.ts}-${index}`,
      x: CHART_LEFT + index * step,
      high,
      low,
      bodyTop,
      bodyHeight,
      up: item.close >= item.open,
    }
  })
})

const hoveredKline = computed(() => {
  if (hoveredIndex.value === null) return null
  return visibleKlines.value[hoveredIndex.value] ?? null
})

function movingAverage(values: number[], period: number) {
  return values.map((_, index) => {
    if (index + 1 < period) return null
    const window = values.slice(index + 1 - period, index + 1)
    return window.reduce((sum, item) => sum + item, 0) / period
  })
}

function buildPricePath(values: Array<number | null>) {
  const { max, range } = chartDomain.value
  const { step } = candleGeometry.value
  const points = values
    .map((value, index) => {
      if (value === null || value <= 0) return null
      const x = CHART_LEFT + index * step
      const y = CHART_TOP + ((max - value) / range) * CHART_HEIGHT
      return `${x},${y}`
    })
    .filter((item): item is string => Boolean(item))

  if (!points.length) return ''
  return `M ${points.join(' L ')}`
}

const closeSeries = computed(() => visibleKlines.value.map((item) => item.close))
const priceMa5 = computed(() => movingAverage(closeSeries.value, 5))
const priceMa10 = computed(() => movingAverage(closeSeries.value, 10))
const priceMa30 = computed(() => movingAverage(closeSeries.value, 30))
const priceMa60 = computed(() => movingAverage(closeSeries.value, 60))
const priceMa5Path = computed(() => buildPricePath(priceMa5.value))
const priceMa10Path = computed(() => buildPricePath(priceMa10.value))
const priceMa30Path = computed(() => buildPricePath(priceMa30.value))
const priceMa60Path = computed(() => buildPricePath(priceMa60.value))
const hoveredMa5 = computed(() => hoveredIndex.value === null ? null : priceMa5.value[hoveredIndex.value] ?? null)
const hoveredMa10 = computed(() => hoveredIndex.value === null ? null : priceMa10.value[hoveredIndex.value] ?? null)
const hoveredMa30 = computed(() => hoveredIndex.value === null ? null : priceMa30.value[hoveredIndex.value] ?? null)
const hoveredMa60 = computed(() => hoveredIndex.value === null ? null : priceMa60.value[hoveredIndex.value] ?? null)

const chartPriceMarks = computed(() => {
  const { min, max } = chartDomain.value
  const middle = min + (max - min) / 2
  const lowerMid = min + (max - min) / 4
  return [max, middle, lowerMid, min].filter((item) => item > 0)
})

const highestCandle = computed(() => {
  const source = visibleKlines.value
  if (!source.length) return null
  const target = source.reduce(
    (best, item, index) => (item.high > best.item.high ? { item, index } : best),
    { item: source[0], index: 0 },
  )
  const { max, range } = chartDomain.value
  const { step } = candleGeometry.value
  return {
    x: CHART_LEFT + target.index * step,
    y: CHART_TOP + ((max - target.item.high) / range) * CHART_HEIGHT,
    value: target.item.high,
  }
})

const lowestCandle = computed(() => {
  const source = visibleKlines.value
  if (!source.length) return null
  const target = source.reduce(
    (best, item, index) => (item.low < best.item.low ? { item, index } : best),
    { item: source[0], index: 0 },
  )
  const { max, range } = chartDomain.value
  const { step } = candleGeometry.value
  return {
    x: CHART_LEFT + target.index * step,
    y: CHART_TOP + ((max - target.item.low) / range) * CHART_HEIGHT,
    value: target.item.low,
  }
})

const volumeSeries = computed(() => visibleKlines.value.map((item) => Math.max(item.volume ?? 0, 0)))
const volumeMax = computed(() => Math.max(...volumeSeries.value, 1))
const volumeMa5 = computed(() => movingAverage(volumeSeries.value, 5))
const volumeMa10 = computed(() => movingAverage(volumeSeries.value, 10))
const volumeMa20 = computed(() => movingAverage(volumeSeries.value, 20))

const volumeBars = computed(() => {
  const max = volumeMax.value
  const { step } = candleGeometry.value
  return visibleKlines.value.map((item, index) => {
    const volume = Math.max(item.volume ?? 0, 0)
    const height = Math.max((volume / max) * VOLUME_VIEWBOX_HEIGHT, volume > 0 ? 6 : 0)
    return {
      key: `${item.ts}-${index}`,
      x: CHART_LEFT + index * step,
      y: VOLUME_VIEWBOX_HEIGHT - height,
      height,
      up: item.close >= item.open,
    }
  })
})

function buildVolumePath(values: Array<number | null>) {
  const max = volumeMax.value
  const { step } = candleGeometry.value
  const points = values
    .map((value, index) => {
      if (value === null || value <= 0) return null
      const x = CHART_LEFT + index * step
      const y = VOLUME_VIEWBOX_HEIGHT - (value / max) * VOLUME_VIEWBOX_HEIGHT
      return `${x},${y}`
    })
    .filter((item): item is string => Boolean(item))
  if (!points.length) return ''
  return `M ${points.join(' L ')}`
}

const volumeMa5Path = computed(() => buildVolumePath(volumeMa5.value))
const volumeMa10Path = computed(() => buildVolumePath(volumeMa10.value))
const volumeMa20Path = computed(() => buildVolumePath(volumeMa20.value))
const hoveredVolumeMa5 = computed(() => hoveredIndex.value === null ? null : volumeMa5.value[hoveredIndex.value] ?? null)
const hoveredVolumeMa10 = computed(() => hoveredIndex.value === null ? null : volumeMa10.value[hoveredIndex.value] ?? null)
const hoveredVolumeMa20 = computed(() => hoveredIndex.value === null ? null : volumeMa20.value[hoveredIndex.value] ?? null)

const metricKline = computed(() => hoveredKline.value ?? visibleKlines.value.at(-1) ?? null)

const hoverCrosshair = computed(() => {
  if (hoveredIndex.value === null || hoveredPrice.value === null) return null
  const candle = chartCandles.value[hoveredIndex.value]
  if (!candle) return null
  const { max, range } = chartDomain.value
  const y = CHART_TOP + ((max - hoveredPrice.value) / range) * CHART_HEIGHT
  return {
    x: candle.x,
    y,
  }
})

const hoverTimeLabel = computed(() => {
  const candle = hoveredKline.value
  const crosshair = hoverCrosshair.value
  if (!candle || !crosshair) return null
  return {
    left: `${Math.max(0, Math.min(crosshair.x / CHART_VIEWBOX_WIDTH * 100, 100))}%`,
    label: formatKlineTime(candle.ts, true),
  }
})

const hoverPriceLabel = computed(() => {
  const crosshair = hoverCrosshair.value
  if (!crosshair || hoveredPrice.value === null) return null
  return {
    top: `${Math.max(0, Math.min(crosshair.y / CHART_VIEWBOX_HEIGHT * 100, 100))}%`,
    label: formatPrice(hoveredPrice.value),
  }
})

const timeAxisMarks = computed(() => {
  const source = visibleKlines.value
  if (!source.length) return []
  const slots = Math.min(5, Math.max(3, Math.floor(source.length / 5)))
  const points = Array.from({ length: slots }, (_, index) => {
    if (index === slots - 1) return source.length - 1
    return Math.round((source.length - 1) * (index / (slots - 1)))
  })
  return Array.from(new Set(points))
    .map((index) => {
      const candle = source[index]
      if (!candle) return null
      const x = chartCandles.value[index]?.x ?? CHART_LEFT
      return { candle, x }
    })
    .filter((item): item is { candle: KlinePayload; x: number } => Boolean(item))
})

function formatNumber(value?: number | null, digits = 2) {
  if (value === null || value === undefined || !Number.isFinite(value)) return '--'
  return new Intl.NumberFormat(getLocale(), {
    minimumFractionDigits: digits,
    maximumFractionDigits: digits,
  }).format(value)
}

function getMaxOffset(nextWindowSize = windowSize.value) {
  return Math.max(0, sortedKlines.value.length - nextWindowSize)
}

function clampOffset(nextOffset: number, nextWindowSize = windowSize.value) {
  return Math.max(0, Math.min(nextOffset, getMaxOffset(nextWindowSize)))
}

function formatPrice(value?: number | null) {
  if (value === null || value === undefined || !Number.isFinite(value)) return '--'
  return formatNumber(value, Math.abs(value) >= 1 ? 4 : 8)
}

function formatVolumeMetric(value?: number | null) {
  if (value === null || value === undefined || !Number.isFinite(value)) return '--'
  if (value >= 1_000_000_000) return `${formatNumber(value / 1_000_000_000, 2)}B`
  if (value >= 1_000_000) return `${formatNumber(value / 1_000_000, 3)}M`
  if (value >= 1_000) return `${formatNumber(value / 1_000, 2)}K`
  return formatNumber(value, 2)
}

function formatKlineTime(ts?: number | null, withDate = false) {
  if (!ts) return '--'
  return new Intl.DateTimeFormat(getLocale(), withDate ? {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  } : {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  }).format(ts)
}

function handleChartHover(event: MouseEvent) {
  if (!visibleKlines.value.length) return
  const rect = (event.currentTarget as HTMLElement).getBoundingClientRect()
  const relativeX = Math.max(0, Math.min(event.clientX - rect.left, rect.width))
  const relativeY = Math.max(0, Math.min(event.clientY - rect.top, rect.height))
  hoverRatioX.value = rect.width > 0 ? relativeX / rect.width : 1
  const approxIndex = Math.round((relativeX / rect.width) * (visibleKlines.value.length - 1))
  const nextIndex = Math.max(0, Math.min(approxIndex, visibleKlines.value.length - 1))
  const { max, range } = chartDomain.value

  hoveredIndex.value = nextIndex
  hoveredPrice.value = max - (relativeY / rect.height) * range
}

function handleChartPointerDown(event: MouseEvent) {
  dragStartX.value = event.clientX
  dragStartOffset.value = windowOffset.value
}

function handleChartPointerMove(event: MouseEvent) {
  if (dragStartX.value === null || !visibleKlines.value.length) return
  const rect = (event.currentTarget as HTMLElement).getBoundingClientRect()
  const innerWidth = rect.width || 1
  const candleWidth = innerWidth / Math.max(visibleKlines.value.length, 1)
  const deltaX = event.clientX - dragStartX.value
  const deltaCandles = Math.round(deltaX / candleWidth)
  windowOffset.value = clampOffset(dragStartOffset.value - deltaCandles)
  handleChartHover(event)
}

function handleChartPointerUp() {
  dragStartX.value = null
}

function handleChartWheel(event: WheelEvent) {
  if (!sortedKlines.value.length) return
  event.preventDefault()
  const currentSize = Math.max(MIN_VISIBLE_KLINES, Math.min(windowSize.value, sortedKlines.value.length))
  const currentOffset = clampOffset(windowOffset.value, currentSize)
  const anchorRatio = Math.max(0, Math.min(hoverRatioX.value, 1))
  const anchorIndexFromLeft = Math.round((currentSize - 1) * anchorRatio)
  const anchorAbsoluteIndex = Math.max(0, Math.min(sortedKlines.value.length - 1, sortedKlines.value.length - currentSize - currentOffset + anchorIndexFromLeft))
  const delta = event.deltaY > 0 ? 8 : -8
  const nextSize = Math.max(MIN_VISIBLE_KLINES, Math.min(sortedKlines.value.length, currentSize + delta))
  if (nextSize === currentSize) return
  windowSize.value = nextSize
  const nextAnchorIndexFromLeft = Math.round((nextSize - 1) * anchorRatio)
  const nextStart = Math.max(0, Math.min(sortedKlines.value.length - nextSize, anchorAbsoluteIndex - nextAnchorIndexFromLeft))
  const nextOffset = sortedKlines.value.length - nextSize - nextStart
  windowOffset.value = clampOffset(nextOffset, nextSize)
}

function handleVolumeHover(event: MouseEvent) {
  if (!visibleKlines.value.length) return
  const rect = (event.currentTarget as HTMLElement).getBoundingClientRect()
  const relativeX = Math.max(0, Math.min(event.clientX - rect.left, rect.width))
  const approxIndex = Math.round((relativeX / rect.width) * (visibleKlines.value.length - 1))
  const nextIndex = Math.max(0, Math.min(approxIndex, visibleKlines.value.length - 1))
  const target = visibleKlines.value[nextIndex]
  hoveredIndex.value = nextIndex
  hoveredPrice.value = target?.close ?? null
}

function clearHover() {
  if (dragStartX.value !== null) return
  hoveredIndex.value = null
  hoveredPrice.value = null
}

function handleMinuteSelect(event: Event) {
  const nextName = (event.target as HTMLSelectElement).value
  const nextInterval = minuteIntervals.value.find((item) => item.name === nextName)
  if (nextInterval) emit('select-interval', nextInterval)
}
</script>

<template>
  <section class="desktop-chart-panel">
    <div class="desktop-chart-toolbar">
      <div class="desktop-chart-toolbar__times">
        <button type="button">Time</button>
        <div class="desktop-interval-menu">
          <span class="desktop-interval-menu__value">{{ selectedMinuteLabel }}</span>
          <select class="desktop-interval-menu__select" :value="selectedMinuteLabel" @change="handleMinuteSelect">
            <option v-for="item in minuteIntervals" :key="`${item.name}-${item.kType}`" :value="item.name">
              {{ item.name }}
            </option>
          </select>
        </div>

        <button
          v-for="item in otherIntervals"
          :key="`${item.name}-${item.kType}`"
          type="button"
          :class="{ active: item.name === selectedIntervalName }"
          @click="emit('select-interval', item)"
        >
          {{ item.name }}
        </button>
      </div>
      <div class="desktop-chart-toolbar__actions">
        <button type="button">{{ t('market.indicator') }}</button>
        <button type="button">{{ t('market.timezone') }}</button>
        <button type="button">{{ t('market.settings') }}</button>
        <button type="button">{{ t('market.screenshot') }}</button>
        <button type="button">{{ t('market.fullscreen') }}</button>
      </div>
    </div>

    <div class="desktop-chart-body">
      <div class="desktop-chart-body__tools">
        <button type="button">─</button>
        <button type="button">╱</button>
        <button type="button">◌</button>
        <button type="button">☷</button>
        <button type="button">⌘</button>
        <button type="button">⌂</button>
        <button type="button">◉</button>
        <button type="button">⌫</button>
      </div>

      <div class="desktop-chart-body__main">
        <div class="desktop-chart-metrics">
          <span>{{ t('market.open') }}: {{ formatPrice(metricKline?.open ?? selectedQuote?.open) }}</span>
          <span>{{ t('market.high') }}: {{ formatPrice(metricKline?.high ?? selectedQuote?.high) }}</span>
          <span>{{ t('market.low') }}: {{ formatPrice(metricKline?.low ?? selectedQuote?.low) }}</span>
          <span>{{ t('market.closePrice') }}: {{ formatPrice(metricKline?.close ?? selectedQuote?.lastPrice) }}</span>
          <span>{{ t('market.volume') }}: {{ formatVolumeMetric(metricKline?.volume ?? selectedQuote?.volume) }}</span>
          <span class="desktop-chart-metrics__ma desktop-chart-metrics__ma--muted">MA(5,10,30,60)</span>
          <span class="desktop-chart-metrics__ma desktop-chart-metrics__ma--orange">MA5: {{ formatPrice(hoveredMa5 ?? priceMa5.at(-1) ?? null) }}</span>
          <span class="desktop-chart-metrics__ma desktop-chart-metrics__ma--violet">MA10: {{ formatPrice(hoveredMa10 ?? priceMa10.at(-1) ?? null) }}</span>
          <span class="desktop-chart-metrics__ma desktop-chart-metrics__ma--blue">MA30: {{ formatPrice(hoveredMa30 ?? priceMa30.at(-1) ?? null) }}</span>
          <span class="desktop-chart-metrics__ma desktop-chart-metrics__ma--pink">MA60: {{ formatPrice(hoveredMa60 ?? priceMa60.at(-1) ?? null) }}</span>
        </div>
        <div
          class="desktop-chart-canvas"
          @mousemove="handleChartHover"
          @mousedown="handleChartPointerDown"
          @mousemove.capture="handleChartPointerMove"
          @mouseup="handleChartPointerUp"
          @mouseleave="handleChartPointerUp(); clearHover()"
          @wheel="handleChartWheel"
        >
          <svg class="desktop-candle-chart" viewBox="0 0 520 240" role="img" :aria-label="t('market.candleChart')">
            <line x1="0" y1="58" x2="520" y2="58" class="grid-line" />
            <line x1="0" y1="120" x2="520" y2="120" class="grid-line" />
            <line x1="0" y1="182" x2="520" y2="182" class="grid-line" />
            <line
              v-for="line in verticalGridLines"
              :key="`grid-${line}`"
              :x1="line"
              :x2="line"
              y1="0"
              y2="240"
              class="grid-line"
            />

            <g v-if="hoverCrosshair">
              <line :x1="hoverCrosshair.x" :x2="hoverCrosshair.x" y1="0" y2="240" class="crosshair-line" />
              <line x1="0" x2="520" :y1="hoverCrosshair.y" :y2="hoverCrosshair.y" class="crosshair-line" />
            </g>

            <path v-if="priceMa5Path" :d="priceMa5Path" class="price-ma price-ma--orange" />
            <path v-if="priceMa10Path" :d="priceMa10Path" class="price-ma price-ma--violet" />
            <path v-if="priceMa30Path" :d="priceMa30Path" class="price-ma price-ma--blue" />
            <path v-if="priceMa60Path" :d="priceMa60Path" class="price-ma price-ma--pink" />

            <g v-for="candle in chartCandles" :key="candle.key">
              <line :x1="candle.x" :x2="candle.x" :y1="candle.high" :y2="candle.low" :class="candle.up ? 'candle-up' : 'candle-down'" />
              <rect :x="candle.x - candleGeometry.bodyWidth / 2" :y="candle.bodyTop" :width="candleGeometry.bodyWidth" :height="candle.bodyHeight" rx="1" :class="candle.up ? 'candle-up' : 'candle-down'" />
            </g>

            <g v-if="highestCandle">
              <line :x1="highestCandle.x" :x2="highestCandle.x + 10" :y1="highestCandle.y" :y2="highestCandle.y" class="price-marker-line price-marker-line--high" />
              <text :x="highestCandle.x - 2" :y="highestCandle.y - 8" class="price-marker-text price-marker-text--high">{{ formatPrice(highestCandle.value) }}</text>
            </g>

            <g v-if="lowestCandle">
              <line :x1="lowestCandle.x - 10" :x2="lowestCandle.x" :y1="lowestCandle.y" :y2="lowestCandle.y" class="price-marker-line price-marker-line--low" />
              <text :x="lowestCandle.x - 34" :y="lowestCandle.y + 16" class="price-marker-text price-marker-text--low">{{ formatPrice(lowestCandle.value) }}</text>
            </g>
          </svg>

          <div class="desktop-price-axis">
            <span v-for="mark in chartPriceMarks" :key="mark">{{ formatPrice(mark) }}</span>
          </div>

          <div v-if="hoverPriceLabel" class="desktop-hover-price" :style="{ top: hoverPriceLabel.top }">
            {{ hoverPriceLabel.label }}
          </div>

          <div v-if="loadingKline" class="desktop-chart-loading">{{ t('common.loading') }}...</div>
        </div>

        <div class="desktop-chart-volume">
          <div class="desktop-chart-volume__legend">
            <span class="desktop-chart-volume__muted">VOL(5,10,20)</span>
            <span class="desktop-chart-volume__ma desktop-chart-volume__ma--orange">MA5: {{ formatVolumeMetric(hoveredVolumeMa5 ?? volumeMa5.at(-1) ?? null) }}</span>
            <span class="desktop-chart-volume__ma desktop-chart-volume__ma--violet">MA10: {{ formatVolumeMetric(hoveredVolumeMa10 ?? volumeMa10.at(-1) ?? null) }}</span>
            <span class="desktop-chart-volume__ma desktop-chart-volume__ma--blue">MA20: {{ formatVolumeMetric(hoveredVolumeMa20 ?? volumeMa20.at(-1) ?? null) }}</span>
            <span class="desktop-chart-volume__ma desktop-chart-volume__ma--green">VOLUME: {{ formatVolumeMetric(metricKline?.volume ?? selectedQuote?.volume ?? null) }}</span>
          </div>

        <div class="desktop-chart-volume__canvas" @mousemove="handleVolumeHover" @mouseleave="clearHover">
            <svg class="desktop-volume-chart" viewBox="0 0 520 36" role="img" :aria-label="t('market.volumeChart')">
              <line v-if="hoverCrosshair" :x1="hoverCrosshair.x" :x2="hoverCrosshair.x" y1="0" y2="36" class="crosshair-line" />
              <g v-for="bar in volumeBars" :key="bar.key">
                <rect :x="bar.x - candleGeometry.bodyWidth / 2" :y="bar.y" :width="candleGeometry.bodyWidth" :height="bar.height" rx="1" :class="bar.up ? 'candle-up' : 'candle-down'" />
              </g>
              <path v-if="volumeMa5Path" :d="volumeMa5Path" class="volume-ma volume-ma--orange" />
              <path v-if="volumeMa10Path" :d="volumeMa10Path" class="volume-ma volume-ma--violet" />
              <path v-if="volumeMa20Path" :d="volumeMa20Path" class="volume-ma volume-ma--blue" />
            </svg>

            <div class="desktop-chart-volume__axis">
              <span>{{ formatVolumeMetric(volumeMax) }}</span>
              <span>{{ formatVolumeMetric(volumeMax / 2) }}</span>
              <span>{{ formatVolumeMetric(volumeMax / 5) }}</span>
            </div>

            <div class="desktop-chart-time-axis">
              <span
                v-for="mark in timeAxisMarks"
                :key="mark.candle.ts"
                :style="{ left: `${Math.max(0, Math.min(mark.x / CHART_VIEWBOX_WIDTH * 100, 100))}%` }"
              >
                {{ formatKlineTime(mark.candle.ts) }}
              </span>
            </div>

            <div v-if="hoverTimeLabel" class="desktop-hover-time" :style="{ left: hoverTimeLabel.left }">
              {{ hoverTimeLabel.label }}
            </div>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.desktop-chart-panel { min-width: 0; }
.desktop-chart-toolbar { display: grid; grid-template-columns: minmax(0, 1fr) auto; align-items: center; gap: 10px; padding: 10px 12px; border-bottom: 1px solid #242633; }
.desktop-chart-toolbar__times, .desktop-chart-toolbar__actions { display: flex; align-items: center; gap: 6px; flex-wrap: nowrap; min-width: 0; overflow: visible; }
.desktop-chart-toolbar__times { gap: 4px; }
.desktop-chart-toolbar__actions { gap: 4px; justify-self: end; }
.desktop-interval-menu { position: relative; flex: 0 0 auto; display: inline-flex; align-items: center; min-width: 74px; min-height: 32px; padding: 0 24px 0 12px; border-radius: 999px; background: #1d202a; }
.desktop-chart-toolbar button { min-height: 32px; padding: 0 10px; border: 0; border-radius: 999px; background: #1d202a; color: #f6f7fb; font: inherit; font-size: 11px; line-height: 1; white-space: nowrap; flex: 0 0 auto; }
.desktop-interval-menu__value { color: #f6f7fb; font-size: 11px; line-height: 1; pointer-events: none; }
.desktop-interval-menu::after { position: absolute; top: 50%; right: 10px; width: 0; height: 0; border-top: 5px solid #0b0c15; border-right: 4px solid transparent; border-left: 4px solid transparent; transform: translateY(-30%); content: ''; }
.desktop-interval-menu__select { position: absolute; inset: 0; width: 100%; height: 100%; opacity: 0; cursor: pointer; }
.desktop-chart-toolbar button.active { background: #fff; color: #0b0c15; }
.desktop-chart-body { display: grid; grid-template-columns: 56px minmax(0, 1fr); min-height: calc(100% - 63px); }
.desktop-chart-body__tools { display: grid; align-content: start; gap: 6px; padding: 12px 8px; border-right: 1px solid #242633; }
.desktop-chart-body__tools button { min-height: 30px; border: 0; background: transparent; color: #a8acb7; font: inherit; font-size: 18px; }
.desktop-chart-body__main { display: grid; grid-template-rows: auto minmax(0, 1fr) 40px; min-width: 0; min-height: 0; }
.desktop-chart-metrics { display: flex; flex-wrap: wrap; gap: 18px; padding: 12px 14px 0; color: #8f929d; font-size: 13px; }
.desktop-chart-metrics__ma { font-weight: 600; }
.desktop-chart-metrics__ma--muted { color: #8f929d; }
.desktop-chart-metrics__ma--orange { color: #ffab19; }
.desktop-chart-metrics__ma--violet { color: #a86de1; }
.desktop-chart-metrics__ma--blue { color: #297dff; }
.desktop-chart-metrics__ma--pink { color: #ff37a0; }
.desktop-chart-canvas { position: relative; margin: 10px 14px; border-radius: 12px; background: linear-gradient(rgba(120, 167, 216, 0.08) 1px, transparent 1px), linear-gradient(90deg, rgba(120, 167, 216, 0.08) 1px, transparent 1px), linear-gradient(180deg, rgba(12, 13, 23, 0.92), rgba(12, 13, 23, 0.92)); background-size: 100% 160px, 160px 100%, auto; cursor: crosshair; user-select: none; }
.desktop-candle-chart, .desktop-volume-chart { width: 100%; height: 100%; }
.grid-line { stroke: rgba(120, 167, 216, 0.12); stroke-width: 1; }
.crosshair-line { stroke: rgba(195, 203, 224, 0.6); stroke-width: 1; stroke-dasharray: 4 4; }
.price-ma, .volume-ma { fill: none; stroke-width: 2; }
.price-ma--orange, .volume-ma--orange { stroke: #ffab19; }
.price-ma--violet, .volume-ma--violet { stroke: #a86de1; }
.price-ma--blue, .volume-ma--blue { stroke: #297dff; }
.price-ma--pink { stroke: #ff37a0; }
.candle-up { fill: #0cd977; stroke: #0cd977; }
.candle-down { fill: #ff5959; stroke: #ff5959; }
.price-marker-line { stroke-width: 1.5; }
.price-marker-line--high { stroke: #14db82; }
.price-marker-line--low { stroke: #ff574c; }
.price-marker-text { font-size: 11px; font-weight: 600; }
.price-marker-text--high { fill: #14db82; }
.price-marker-text--low { fill: #ff574c; }
.desktop-price-axis { position: absolute; top: 12px; right: 10px; display: grid; gap: 42px; color: #8f929d; font-size: 12px; }
.desktop-hover-price {
  position: absolute;
  right: 0;
  transform: translateY(-50%);
  padding: 6px 10px;
  border-radius: 4px 0 0 4px;
  background: #7e848f;
  color: #fff;
  font-size: 12px;
  font-weight: 600;
}
.desktop-chart-loading { position: absolute; inset: 0; display: grid; place-items: center; color: #8f929d; font-size: 14px; background: rgba(11, 12, 21, 0.45); }
.desktop-chart-volume { display: grid; grid-template-rows: auto minmax(0, 1fr); margin: 0 14px 10px; border-top: 1px solid #2b2f3a; padding-top: 4px; }
.desktop-chart-volume__legend { display: flex; flex-wrap: wrap; align-items: center; gap: 8px 14px; min-height: 18px; color: #8f929d; font-size: 10px; }
.desktop-chart-volume__muted { color: #8f929d; }
.desktop-chart-volume__ma { font-weight: 600; }
.desktop-chart-volume__ma--green { color: #14db82; }
.desktop-chart-volume__canvas { position: relative; min-height: 0; padding-bottom: 24px; }
.desktop-chart-volume__axis { position: absolute; top: 0; right: 0; display: grid; gap: 6px; color: #8f929d; font-size: 10px; text-align: right; }
.desktop-chart-time-axis {
  position: absolute;
  right: 0;
  bottom: -2px;
  left: 0;
  height: 18px;
  color: #8f929d;
  font-size: 10px;
}
.desktop-chart-time-axis span {
  position: absolute;
  transform: translateX(-50%);
  white-space: nowrap;
}
.desktop-hover-time {
  position: absolute;
  bottom: 0;
  transform: translateX(-50%);
  padding: 6px 10px;
  border: 1px solid #ff5a49;
  background: #676d77;
  color: #fff;
  font-size: 10px;
  font-weight: 600;
  white-space: nowrap;
}
</style>
