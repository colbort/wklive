<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute } from 'vue-router'

import DesktopMarketTradeView from '@/components/markets/DesktopMarketTradeView.vue'
import MobileTradeView from '@/components/trades/MobileTradeView.vue'
import { useDevice } from '@/composables/useDevice'
import { useTradingDesk } from '@/composables/useTradingDesk'

// 交易页：手机端展示交易表单与订单列表，桌面端复用市场交易组合台。
const route = useRoute()
const { isDesktop } = useDevice()
const detailVisible = computed(() => isDesktop.value)
const orderMode = ref<'market' | 'limit'>('market')
const productMenuOpen = ref(false)
const desktopProductsExpanded = ref(false)
const desktopOrderbookExpanded = ref(true)
const {
  selectedCategoryType,
  selectedProductKey,
  selectedIntervalName,
  loadingBootstrap,
  loadingKline,
  depthSnapshot,
  tickSnapshot,
  klineSnapshot,
  categories,
  intervals,
  selectedCategory,
  selectedProduct,
  selectedQuote,
  placeholderPrice,
  placeholderChange,
  priceTrend,
  desktopProductRows,
  productSheetRows,
  desktopStats,
  selectProduct,
  selectInterval,
  coinGlyph,
  productKey,
} = useTradingDesk({
  detailVisible,
  tickLimit: 24,
})
const tradeKind = computed(() => {
  const code = `${selectedCategory.value?.categoryCode || ''} ${selectedCategory.value?.categoryName || ''}`.toLowerCase()
  if (code.includes('stock') || code.includes('股票')) return 'stock'
  if (code.includes('option') || code.includes('期权')) return 'option'
  if (code.includes('forex') || code.includes('外汇')) return 'forex'
  if (code.includes('commodity') || code.includes('大宗')) return 'commodity'
  return 'crypto'
})

watch(
  () => route.query,
  (query) => {
    const categoryType = Number(query.categoryType)
    if (Number.isFinite(categoryType) && categoryType > 0) {
      selectedCategoryType.value = categoryType
    }

    const market = String(query.market || '')
    const symbol = String(query.symbol || '')
    if (market && symbol) {
      selectedProductKey.value = productKey({ market, symbol })
    }
  },
  { immediate: true },
)
watch(selectedProductKey, () => {
  productMenuOpen.value = false
})

function closeProductSheet() {
  productMenuOpen.value = false
}

function toggleDesktopProducts() {
  desktopProductsExpanded.value = !desktopProductsExpanded.value
}

function toggleDesktopOrderbook() {
  desktopOrderbookExpanded.value = !desktopOrderbookExpanded.value
}
</script>

<template>
  <section class="trade-page" :aria-busy="loadingBootstrap">
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
    <MobileTradeView
      v-else
      :categories="categories"
      :selected-category-type="selectedCategoryType"
      :selected-category="selectedCategory"
      :selected-product="selectedProduct"
      :selected-product-key="selectedProductKey"
      :trade-kind="tradeKind"
      :price-trend="priceTrend"
      :placeholder-price="placeholderPrice"
      :placeholder-change="placeholderChange"
      :selected-quote="selectedQuote"
      :product-menu-open="productMenuOpen"
      :product-sheet-rows="productSheetRows"
      :order-mode="orderMode"
      :coin-glyph="coinGlyph"
      @select-category="selectedCategoryType = $event"
      @open-product-menu="productMenuOpen = true"
      @close-product-sheet="closeProductSheet"
      @select-product="selectProduct"
      @update:order-mode="orderMode = $event"
    />
  </section>
</template>

<style scoped>
.trade-page {
  width: 100%;
  max-width: 100%;
  min-height: calc(100dvh - 72px);
  padding: 18px 22px 112px;
  overflow-x: hidden;
  background: #0b0c15;
  color: #f6f7fb;
}
</style>
