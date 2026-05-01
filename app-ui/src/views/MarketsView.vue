<script setup lang="ts">
import { computed, ref } from 'vue'

import DesktopMarketTradeView from '@/components/markets/DesktopMarketTradeView.vue'
import MarketChartView from '@/components/markets/MarketChartView.vue'
import MarketQuotesView from '@/components/markets/MarketQuotesView.vue'
import MarketTopTabs from '@/components/markets/MarketTopTabs.vue'
import type { MarketTopTab, MarketTopTabItem } from '@/components/markets/types'
import { getAccessToken } from '@/api/http'
import { useDevice } from '@/composables/useDevice'
import { useTradingDesk } from '@/composables/useTradingDesk'
import type { ItickTenantProduct } from '@/types/itick'

// 市场页：手机端负责自选/行情/图表切换，桌面端承载市场与交易组合台。
const { isDesktop } = useDevice()

const topTabs: MarketTopTabItem[] = [
  { key: 'watchlist', label: '自选' },
  { key: 'markets', label: '行情' },
  { key: 'chart', label: '图表' },
]

const activeTopTab = ref<MarketTopTab>('markets')
const orderMode = ref<'market' | 'limit'>('market')
const desktopProductsExpanded = ref(false)
const desktopOrderbookExpanded = ref(true)

const showingDesktopDesk = computed(() => isDesktop.value || activeTopTab.value === 'chart')
const isLoggedIn = computed(() => Boolean(getAccessToken()))

const {
  selectedCategoryType,
  selectedProductKey,
  selectedIntervalName,
  loadingBootstrap,
  loadingKline,
  depthSnapshot,
  tickSnapshot,
  klineSnapshot,
  wsState,
  wsError,
  categories,
  products,
  intervals,
  selectedCategory,
  selectedCategoryCode,
  selectedProduct,
  selectedQuote,
  placeholderPrice,
  priceTrend,
  marketRows,
  desktopProductRows,
  desktopStats,
  loadPreviousKlinePage,
  selectCategory,
  selectProduct,
  selectInterval,
  coinGlyph,
} = useTradingDesk({
  detailVisible: showingDesktopDesk,
  tickLimit: 12,
})

function openProductChart(product: ItickTenantProduct) {
  selectProduct(product)
  activeTopTab.value = 'chart'
}

function toggleDesktopProducts() {
  desktopProductsExpanded.value = !desktopProductsExpanded.value
}

function toggleDesktopOrderbook() {
  desktopOrderbookExpanded.value = !desktopOrderbookExpanded.value
}
</script>

<template>
  <section class="markets-page" :aria-busy="loadingBootstrap">
    <DesktopMarketTradeView
      v-if="isDesktop"
      :selected-product="selectedProduct"
      :price-trend="priceTrend"
      :placeholder-price="placeholderPrice"
      :desktop-stats="desktopStats"
      :desktop-product-rows="desktopProductRows"
      :selected-product-key="selectedProductKey"
      :desktop-products-expanded="desktopProductsExpanded"
      :desktop-orderbook-expanded="desktopOrderbookExpanded"
      :selected-quote="selectedQuote"
      :depth-snapshot="depthSnapshot"
      :tick-snapshot="tickSnapshot"
      :kline-snapshot="klineSnapshot"
      :intervals="intervals"
      :selected-interval-name="selectedIntervalName"
      :loading-kline="loadingKline"
      :order-mode="orderMode"
      :coin-glyph="coinGlyph"
      @toggle-products="toggleDesktopProducts"
      @toggle-orderbook="toggleDesktopOrderbook"
      @select-product="selectProduct"
      @select-interval="selectInterval"
      @update:order-mode="orderMode = $event"
    />

    <template v-else>
      <MarketTopTabs :tabs="topTabs" :active-tab="activeTopTab" @change="activeTopTab = $event" />

      <div v-if="activeTopTab === 'markets'" class="markets-page__mobile">
        <MarketQuotesView
          :categories="categories"
          :selected-category-type="selectedCategoryType"
          :selected-category-name="selectedCategory?.categoryName || ''"
          :selected-category-code="selectedCategoryCode"
          :ws-state="wsState"
          :ws-error="wsError"
          :loading="loadingBootstrap"
          :rows="marketRows"
          :selected-product-key="selectedProductKey"
          @select-category="selectCategory"
          @select-product="openProductChart"
        />
      </div>

      <div v-else-if="activeTopTab === 'watchlist'" class="markets-page__watchlist">
        <div v-if="isLoggedIn" class="watchlist-empty">自选功能整理中</div>
        <div v-else class="watchlist-empty">登录后可查看自选产品</div>
      </div>

      <div v-else class="markets-page__mobile">
        <MarketChartView
          :products="products"
          :rows="marketRows"
          :category-name="selectedCategory?.categoryName || ''"
          :selected-product-key="selectedProductKey"
          :selected-quote="selectedQuote"
          :kline-snapshot="klineSnapshot"
          :depth-snapshot="depthSnapshot"
          :tick-snapshot="tickSnapshot"
          :loading-kline="loadingKline"
          :intervals="intervals"
          :selected-interval-name="selectedIntervalName"
          @select-product="selectedProductKey = $event"
          @select-interval="selectInterval"
          @load-previous-page="loadPreviousKlinePage"
        />
      </div>
    </template>
  </section>
</template>

<style scoped>
.markets-page {
  width: 100%;
  max-width: 100%;
  min-height: calc(100dvh - 72px);
  padding: 18px 22px 112px;
  overflow-x: hidden;
  background: #0b0c15;
  color: #f6f7fb;
}

.markets-page__mobile {
  min-height: calc(100dvh - 72px);
}

.markets-page__watchlist {
  display: grid;
  min-height: calc(100dvh - 160px);
  place-items: center;
  padding: 24px 18px;
}

.watchlist-empty {
  color: #8f929d;
  font-size: 15px;
}
</style>
