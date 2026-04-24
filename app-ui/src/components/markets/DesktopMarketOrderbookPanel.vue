<script setup lang="ts">
import { computed, ref } from 'vue'
import type { DepthPayload, QuotePayload, TickPayload } from '@/types/itick'

const props = defineProps<{
  selectedQuote: QuotePayload | null
  depthSnapshot: DepthPayload | null
  tickSnapshot: TickPayload[]
}>()

const activeBookTab = ref<'book' | 'trades'>('book')
const askRows = computed(() => props.depthSnapshot?.asks.slice(0, 10) ?? [])
const bidRows = computed(() => props.depthSnapshot?.bids.slice(0, 10) ?? [])
const maxAskVolume = computed(() => Math.max(...askRows.value.map((item) => item.volume), 1))
const maxBidVolume = computed(() => Math.max(...bidRows.value.map((item) => item.volume), 1))
const tradeRows = computed(() =>
  props.tickSnapshot.map((item, index, list) => {
    const next = list[index + 1]
    return {
      ...item,
      direction: next && item.lastPrice < next.lastPrice ? 'down' : 'up',
    }
  }),
)

function formatNumber(value?: number | null, digits = 2) {
  if (value === null || value === undefined || !Number.isFinite(value)) return '--'
  return new Intl.NumberFormat('zh-CN', {
    minimumFractionDigits: digits,
    maximumFractionDigits: digits,
  }).format(value)
}

function formatPrice(value?: number | null) {
  if (value === null || value === undefined || !Number.isFinite(value)) return '--'
  return formatNumber(value, Math.abs(value) >= 1 ? 4 : 8)
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
</script>

<template>
  <aside class="desktop-orderbook-panel">
    <header class="desktop-orderbook-panel__header">
      <nav>
        <button type="button" :class="{ active: activeBookTab === 'book' }" @click="activeBookTab = 'book'">订单簿</button>
        <button type="button" :class="{ active: activeBookTab === 'trades' }" @click="activeBookTab = 'trades'">最新成交</button>
      </nav>
    </header>

    <div v-if="activeBookTab === 'book'" class="desktop-orderbook-panel__content">
      <div class="desktop-orderbook-panel__subtools">
        <span class="desktop-orderbook-panel__swatches">
          <i class="swatch swatch--red" />
          <i class="swatch swatch--green" />
          <i class="swatch swatch--mix" />
        </span>
      </div>

      <div class="desktop-orderbook-panel__columns">
        <span>价格(USDT)</span>
        <span>数量(BTC)</span>
      </div>

      <div class="desktop-orderbook-panel__rows desktop-orderbook-panel__rows--asks">
        <p v-for="(item, index) in askRows" :key="`ask-${item.price}-${index}`">
          <span>{{ formatPrice(item.price) }}</span>
          <strong>{{ formatNumber(item.volume, 8) }}</strong>
          <i :style="{ width: `${Math.max(8, (item.volume / maxAskVolume) * 100)}%` }" />
        </p>
      </div>

      <div class="desktop-orderbook-panel__mid">
        <strong>{{ formatPrice(selectedQuote?.lastPrice) }} ↑</strong>
        <span>{{ formatPrice(selectedQuote?.open) }}</span>
      </div>

      <div class="desktop-orderbook-panel__rows desktop-orderbook-panel__rows--bids">
        <p v-for="(item, index) in bidRows" :key="`bid-${item.price}-${index}`">
          <span>{{ formatPrice(item.price) }}</span>
          <strong>{{ formatNumber(item.volume, 8) }}</strong>
          <i :style="{ width: `${Math.max(8, (item.volume / maxBidVolume) * 100)}%` }" />
        </p>
      </div>
    </div>

    <div v-else class="desktop-trades-panel">
      <header class="desktop-orderbook-panel__columns">
        <span>价格(USDT)</span>
        <span>数量(BTC)</span>
        <span>时间</span>
      </header>
      <div class="desktop-trades-panel__rows">
        <p
          v-for="(item, index) in tradeRows"
          :key="`${item.ts}-${index}`"
          :class="{ 'desktop-trades-panel__row--down': item.direction === 'down' }"
        >
          <span>{{ formatPrice(item.lastPrice) }}</span>
          <strong>{{ formatNumber(item.volume, 8) }}</strong>
          <time>{{ formatTime(item.ts) }}</time>
        </p>
      </div>
    </div>
  </aside>
</template>

<style scoped>
.desktop-orderbook-panel {
  display: grid;
  grid-template-rows: auto minmax(0, 1fr);
  width: 100%;
  min-height: 0;
  border-right: 1px solid #242633;
  background: linear-gradient(180deg, rgba(11, 12, 21, 0.98), rgba(11, 12, 21, 0.92));
}

.desktop-orderbook-panel__header {
  padding: 14px 14px 12px;
  border-bottom: 1px solid #242633;
}

.desktop-orderbook-panel__header nav {
  display: flex;
  gap: 18px;
}

.desktop-orderbook-panel__header button {
  position: relative;
  border: 0;
  background: transparent;
  color: #8f929d;
  font: inherit;
  font-size: 13px;
  font-weight: 500;
}

.desktop-orderbook-panel__header button.active {
  color: #fff;
}

.desktop-orderbook-panel__content {
  display: grid;
  grid-template-rows: auto auto minmax(0, 1fr) auto minmax(0, 1fr);
  min-height: 0;
  padding: 12px 12px 14px;
}

.desktop-orderbook-panel__subtools {
  display: flex;
  justify-content: flex-start;
  margin-bottom: 10px;
}

.desktop-orderbook-panel__swatches {
  display: inline-flex;
  gap: 10px;
}

.swatch {
  width: 16px;
  height: 16px;
  border-radius: 3px;
  background: #364052;
}

.swatch--red {
  background: linear-gradient(180deg, #ff5a49 0 50%, #2d3443 50%);
}

.swatch--green {
  background: linear-gradient(180deg, #13db7b 0 50%, #2d3443 50%);
}

.swatch--mix {
  background: linear-gradient(90deg, #ff5a49 0 33%, #13db7b 33% 66%, #2d3443 66%);
}

.desktop-orderbook-panel__columns,
.desktop-orderbook-panel__rows p,
.desktop-orderbook-panel__mid {
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 18px;
}

.desktop-orderbook-panel__columns {
  margin-bottom: 10px;
  color: #8f929d;
  font-size: 11px;
}

.desktop-orderbook-panel__rows {
  min-height: 0;
  overflow-y: auto;
  overscroll-behavior: contain;
}

.desktop-orderbook-panel__rows p {
  position: relative;
  margin: 0 0 8px;
  font-size: 12px;
}

.desktop-orderbook-panel__rows p i {
  position: absolute;
  top: 0;
  right: 0;
  bottom: 0;
  z-index: 0;
  opacity: 0.14;
}

.desktop-orderbook-panel__rows span,
.desktop-orderbook-panel__mid span {
  color: #ff574c;
}

.desktop-orderbook-panel__rows p > span,
.desktop-orderbook-panel__rows p > strong,
.desktop-orderbook-panel__mid > strong,
.desktop-orderbook-panel__mid > span {
  position: relative;
  z-index: 1;
}

.desktop-orderbook-panel__rows strong {
  color: #fff;
  font-weight: 500;
}

.desktop-orderbook-panel__rows--bids span,
.desktop-orderbook-panel__mid strong {
  color: #13db7b;
}

.desktop-orderbook-panel__rows--asks p i {
  background: #ff5959;
}

.desktop-orderbook-panel__rows--bids p i {
  background: #0cd977;
}

.desktop-orderbook-panel__mid {
  align-items: center;
  padding: 12px 0;
}

.desktop-orderbook-panel__mid strong {
  font-size: 16px;
  font-weight: 700;
}

.desktop-orderbook-panel__mid span {
  color: #8f929d;
  font-size: 12px;
}

.desktop-trades-panel {
  display: grid;
  grid-template-rows: auto minmax(0, 1fr);
  min-height: 0;
  padding: 12px 12px 14px;
}

.desktop-trades-panel .desktop-orderbook-panel__columns {
  grid-template-columns: 1fr auto auto;
}

.desktop-trades-panel__rows {
  min-height: 0;
  overflow-y: auto;
  overscroll-behavior: contain;
}

.desktop-trades-panel__rows p {
  display: grid;
  grid-template-columns: 1fr auto auto;
  gap: 10px;
  margin: 0 0 8px;
  font-size: 12px;
}

.desktop-trades-panel__rows span {
  color: #13db7b;
}

.desktop-trades-panel__row--down span {
  color: #ff5959;
}

.desktop-trades-panel__rows strong,
.desktop-trades-panel__rows time {
  color: #fff;
}
</style>
