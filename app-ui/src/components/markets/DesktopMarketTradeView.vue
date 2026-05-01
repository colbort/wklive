<script setup lang='ts'>
import type { Interval } from '@/types/core'
import type { DepthPayload, ItickTenantProduct, KlinePayload, QuotePayload, TickPayload } from '@/types/itick'
import DesktopTradeOrdersPanel from '@/components/trades/DesktopTradeOrdersPanel.vue'
import DesktopTradeSubmitPanel from '@/components/trades/DesktopTradeSubmitPanel.vue'
import DesktopMarketChartPanel from './DesktopMarketChartPanel.vue'
import DesktopMarketHeader from './DesktopMarketHeader.vue'
import DesktopMarketOrderbookPanel from './DesktopMarketOrderbookPanel.vue'
import DesktopMarketProductsPanel from './DesktopMarketProductsPanel.vue'
import type { DesktopProductRow, DesktopStat } from './desktop-types'

defineProps<{
  selectedProduct: ItickTenantProduct | null
  priceTrend: 'up' | 'down'
  placeholderPrice: string
  desktopStats: DesktopStat[]
  desktopProductRows: DesktopProductRow[]
  selectedProductKey: string
  desktopProductsExpanded: boolean
  desktopOrderbookExpanded: boolean
  selectedQuote: QuotePayload | null
  depthSnapshot: DepthPayload | null
  tickSnapshot: TickPayload[]
  klineSnapshot: KlinePayload[]
  intervals: Interval[]
  selectedIntervalName: string
  loadingKline: boolean
  orderMode: 'market' | 'limit'
  coinGlyph: (product: ItickTenantProduct) => string
}>()

const emit = defineEmits<{
  (e: 'toggle-products'): void
  (e: 'toggle-orderbook'): void
  (e: 'select-product', product: ItickTenantProduct): void
  (e: 'select-interval', interval: Interval): void
  (e: 'update:orderMode', value: 'market' | 'limit'): void
}>()
</script>

<template>
  <section class="trade-desktop">
    <DesktopMarketHeader
      :selected-product="selectedProduct"
      :price-trend="priceTrend"
      :placeholder-price="placeholderPrice"
      :desktop-stats="desktopStats"
    />

    <section class="desktop-main">
      <div
        class="desktop-trade-workspace"
        :class="{
          'desktop-trade-workspace--products-open': desktopProductsExpanded,
          'desktop-trade-workspace--orderbook-open': desktopOrderbookExpanded,
        }"
      >
        <DesktopMarketProductsPanel
          v-if="desktopProductsExpanded"
          :rows="desktopProductRows"
          :selected-product-key="selectedProductKey"
          :coin-glyph="coinGlyph"
          @select-product="emit('select-product', $event)"
        />

        <div class="desktop-side-toggle">
          <button
            type="button"
            class="desktop-products-toggle"
            aria-label="展开或收起产品列表"
            @click="emit('toggle-products')"
          >
            ☰
          </button>
        </div>

        <DesktopMarketChartPanel
          :selected-quote="selectedQuote"
          :kline-snapshot="klineSnapshot"
          :intervals="intervals"
          :selected-interval-name="selectedIntervalName"
          :loading-kline="loadingKline"
          @select-interval="emit('select-interval', $event)"
        />

        <div class="desktop-side-toggle desktop-side-toggle--book">
          <button
            type="button"
            class="desktop-products-toggle"
            aria-label="展开或收起订单簿"
            @click="emit('toggle-orderbook')"
          >
            ☰
          </button>
        </div>

        <DesktopMarketOrderbookPanel
          v-if="desktopOrderbookExpanded"
          :selected-quote="selectedQuote"
          :depth-snapshot="depthSnapshot"
          :tick-snapshot="tickSnapshot"
        />

        <DesktopTradeSubmitPanel
          :selected-product="selectedProduct"
          :order-mode="orderMode"
          @update:order-mode="emit('update:orderMode', $event)"
        />
      </div>

      <DesktopTradeOrdersPanel />
    </section>
  </section>
</template>

<style scoped>
.trade-desktop {
  display: grid;
  grid-template-rows: auto minmax(0, 1fr);
  min-height: calc(100dvh - 72px);
  margin: -18px -22px -112px;
}

.desktop-main {
  display: grid;
  grid-template-rows: 540px minmax(220px, 1fr);
  min-height: 0;
}

.desktop-trade-workspace {
  display: grid;
  grid-template-columns: 52px minmax(0, 1fr) 52px 328px;
  min-height: 0;
  overflow: hidden;
}

.desktop-trade-workspace--products-open {
  grid-template-columns: 330px 52px minmax(0, 1fr) 52px 328px;
}

.desktop-trade-workspace--orderbook-open {
  grid-template-columns: 52px minmax(0, 1fr) 52px 252px 312px;
}

.desktop-trade-workspace--products-open.desktop-trade-workspace--orderbook-open {
  grid-template-columns: 282px 52px minmax(0, 1fr) 52px 246px 300px;
}

.desktop-side-toggle {
  display: grid;
  place-items: start center;
  border-right: 1px solid #242633;
}

.desktop-side-toggle--book {
  border-left: 1px solid #242633;
  border-right: 1px solid #242633;
}

.desktop-products-toggle {
  margin-top: 18px;
  border: 0;
  background: transparent;
  color: #fff;
  font-size: 26px;
}
</style>
