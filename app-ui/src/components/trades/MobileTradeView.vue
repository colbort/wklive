<script setup lang="ts">
import MobileMarketDepthPreview from '@/components/markets/MobileMarketDepthPreview.vue'
import MobileMarketTradeHeader from '@/components/markets/MobileMarketTradeHeader.vue'
import TradeOrdersPanel from '@/components/trades/TradeOrdersPanel.vue'
import MobileTradeSubmitPanel from '@/components/trades/MobileTradeSubmitPanel.vue'
import type {
  DepthPayload,
  ItickTenantCategory,
  ItickTenantProduct,
  QuotePayload,
  TickPayload,
} from '@/types/itick'
import type {
  TradeOrder,
  TradeSymbol,
  TradeSymbolContract,
  TradeSymbolLeverageConfig,
  TradeSymbolSpot,
} from '@/types/trade'

type ProductSheetRow = {
  key: string
  product: ItickTenantProduct
  price: string
  change: string
  direction: 'up' | 'down' | 'flat'
}
type SubmitSide = 'buy' | 'sell'
type TradeSymbolDetail = {
  symbol: TradeSymbol | null
  spot: TradeSymbolSpot | null
  contract: TradeSymbolContract | null
  leverageConfigs: TradeSymbolLeverageConfig[]
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
  tickSnapshot: TickPayload[]
  productMenuOpen: boolean
  productSheetRows: ProductSheetRow[]
  orderMode: 'market' | 'limit'
  selectedTradeSymbol: TradeSymbol | null
  tradeSymbolDetail: TradeSymbolDetail | null
  tradeSymbolLoading: boolean
  isLoggedIn: boolean
  tradeAvailable: boolean
  tradePrice: string
  tradeQty: string
  tradePercent: number
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
  tradeOrders: TradeOrder[]
  ordersLoading: boolean
  ordersError: string
  cancelingOrderId: number | null
  coinGlyph: (product: ItickTenantProduct) => string
}>()

const emit = defineEmits<{
  (e: 'select-category', categoryType: number): void
  (e: 'open-product-menu'): void
  (e: 'close-product-sheet'): void
  (e: 'select-product', product: ItickTenantProduct): void
  (e: 'update:orderMode', value: 'market' | 'limit'): void
  (e: 'update:tradePrice', value: string): void
  (e: 'update:tradeQty', value: string): void
  (e: 'update:tradePercent', value: number): void
  (e: 'update:marginMode', value: number): void
  (e: 'update:leverage', value: number): void
  (e: 'update:takeProfitPrice', value: string): void
  (e: 'update:stopLossPrice', value: string): void
  (e: 'submit-order', side: SubmitSide): void
  (e: 'cancel-order', order: TradeOrder): void
  (e: 'refresh-orders'): void
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
        :selected-trade-symbol="selectedTradeSymbol"
        :trade-symbol-detail="tradeSymbolDetail"
        :trade-symbol-loading="tradeSymbolLoading"
        :is-logged-in="isLoggedIn"
        :trade-available="tradeAvailable"
        :trade-price="tradePrice"
        :trade-qty="tradeQty"
        :trade-percent="tradePercent"
        :reference-price="selectedQuote?.lastPrice || placeholderPrice"
        :margin-mode="marginMode"
        :leverage="leverage"
        :max-leverage="maxLeverage"
        :leverage-values="leverageValues"
        :take-profit-price="takeProfitPrice"
        :stop-loss-price="stopLossPrice"
        :settle-asset="settleAsset"
        :available-balance="availableBalance"
        :long-position-qty="longPositionQty"
        :short-position-qty="shortPositionQty"
        :trade-message="tradeMessage"
        :trade-error="tradeError"
        :submitting-side="submittingSide"
        @update:order-mode="emit('update:orderMode', $event)"
        @update:trade-price="emit('update:tradePrice', $event)"
        @update:trade-qty="emit('update:tradeQty', $event)"
        @update:trade-percent="emit('update:tradePercent', $event)"
        @update:margin-mode="emit('update:marginMode', $event)"
        @update:leverage="emit('update:leverage', $event)"
        @update:take-profit-price="emit('update:takeProfitPrice', $event)"
        @update:stop-loss-price="emit('update:stopLossPrice', $event)"
        @submit-order="emit('submit-order', $event)"
      />
      <MobileMarketDepthPreview
        :selected-product="selectedProduct"
        :depth-snapshot="depthSnapshot"
        :selected-quote="selectedQuote"
        :tick-snapshot="tickSnapshot"
        :placeholder-price="placeholderPrice"
      />
    </section>

    <MobileTradeSubmitPanel
      v-else
      :selected-product="selectedProduct"
      :trade-kind="tradeKind"
      :order-mode="orderMode"
      :selected-trade-symbol="selectedTradeSymbol"
      :trade-symbol-detail="tradeSymbolDetail"
      :trade-symbol-loading="tradeSymbolLoading"
      :is-logged-in="isLoggedIn"
      :trade-available="tradeAvailable"
      :trade-price="tradePrice"
      :trade-qty="tradeQty"
      :trade-percent="tradePercent"
      :reference-price="selectedQuote?.lastPrice || placeholderPrice"
      :margin-mode="marginMode"
      :leverage="leverage"
      :max-leverage="maxLeverage"
      :leverage-values="leverageValues"
      :take-profit-price="takeProfitPrice"
      :stop-loss-price="stopLossPrice"
      :settle-asset="settleAsset"
      :available-balance="availableBalance"
      :long-position-qty="longPositionQty"
      :short-position-qty="shortPositionQty"
      :trade-message="tradeMessage"
      :trade-error="tradeError"
      :submitting-side="submittingSide"
      @update:order-mode="emit('update:orderMode', $event)"
      @update:trade-price="emit('update:tradePrice', $event)"
      @update:trade-qty="emit('update:tradeQty', $event)"
      @update:trade-percent="emit('update:tradePercent', $event)"
      @update:margin-mode="emit('update:marginMode', $event)"
      @update:leverage="emit('update:leverage', $event)"
      @update:take-profit-price="emit('update:takeProfitPrice', $event)"
      @update:stop-loss-price="emit('update:stopLossPrice', $event)"
      @submit-order="emit('submit-order', $event)"
    />

    <TradeOrdersPanel
      :show-premarket="tradeKind === 'stock'"
      :orders="tradeOrders"
      :loading="ordersLoading"
      :error="ordersError"
      :is-logged-in="isLoggedIn"
      :selected-trade-symbol="selectedTradeSymbol"
      :canceling-order-id="cancelingOrderId"
      @cancel-order="emit('cancel-order', $event)"
      @refresh="emit('refresh-orders')"
    />
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
