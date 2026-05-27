<script setup lang='ts'>
import MobileMarketDepthPreview from '@/components/markets/MobileMarketDepthPreview.vue'
import MobileMarketTradeHeader from '@/components/markets/MobileMarketTradeHeader.vue'
import TradeOrdersPanel from '@/components/trades/TradeOrdersPanel.vue'
import MobileTradeSubmitPanel from '@/components/trades/MobileTradeSubmitPanel.vue'
import type { DepthPayload, ItickTenantCategory, ItickTenantProduct, QuotePayload } from '@/types/itick'

type ProductSheetRow = {
  key: string
  product: ItickTenantProduct
  price: string
  change: string
  direction: 'up' | 'down' | 'flat'
}

defineProps<{
  categories: ItickTenantCategory[]
  selectedCategoryType: number | null
  selectedCategory: ItickTenantCategory | null
  selectedProduct: ItickTenantProduct | null
  selectedProductKey: string
  tradeKind: 'stock' | 'option' | 'forex' | 'commodity' | 'crypto'
  priceTrend: 'up' | 'down'
  placeholderPrice: string
  placeholderChange: string
  selectedQuote: QuotePayload | null
  depthSnapshot: DepthPayload | null
  productMenuOpen: boolean
  productSheetRows: ProductSheetRow[]
  orderMode: 'market' | 'limit'
  coinGlyph: (product: ItickTenantProduct) => string
}>()

const emit = defineEmits<{
  (e: 'select-category', categoryType: number): void
  (e: 'open-product-menu'): void
  (e: 'close-product-sheet'): void
  (e: 'select-product', product: ItickTenantProduct): void
  (e: 'update:orderMode', value: 'market' | 'limit'): void
}>()
</script>

<template>
  <div class="mobile-trade-view">
    <MobileMarketTradeHeader
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
      :coin-glyph="coinGlyph"
      @select-category="emit('select-category', $event)"
      @open-product-menu="emit('open-product-menu')"
      @close-product-sheet="emit('close-product-sheet')"
      @select-product="emit('select-product', $event)"
    />

    <section v-if="tradeKind === 'crypto'" class="mobile-contract-layout">
      <MobileTradeSubmitPanel
        :selected-product="selectedProduct"
        :trade-kind="tradeKind"
        :order-mode="orderMode"
        @update:order-mode="emit('update:orderMode', $event)"
      />
      <MobileMarketDepthPreview
        :selected-product="selectedProduct"
        :depth-snapshot="depthSnapshot"
        :placeholder-price="placeholderPrice"
      />
    </section>

    <MobileTradeSubmitPanel
      v-else
      :selected-product="selectedProduct"
      :trade-kind="tradeKind"
      :order-mode="orderMode"
      @update:order-mode="emit('update:orderMode', $event)"
    />

    <TradeOrdersPanel :show-premarket="tradeKind === 'stock'" />
  </div>
</template>

<style scoped>
.mobile-trade-view {
  min-width: 0;
  max-width: 100%;
  overflow-x: hidden;
}

@media (max-width: 767px) {
  .mobile-trade-view {
    padding-top: 50px;
  }
}

.mobile-contract-layout {
  display: grid;
  grid-template-columns: minmax(0, 1.1fr) minmax(112px, 0.9fr);
  gap: 12px;
  min-width: 0;
  align-items: start;
}

@media (max-width: 340px) {
  .mobile-contract-layout {
    grid-template-columns: 1fr;
  }
}
</style>
