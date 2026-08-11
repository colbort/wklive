<script setup lang="ts">
import TradeDepthPreview from '@/components/trades/DepthPreview.vue'
import TradeHeader from '@/components/trades/Header.vue'
import TradeOrdersPanel from '@/components/trades/OrdersPanel.vue'
import ContractTradePanel from '@/components/trades/panels/ContractTradePanel.vue'
import SecondsTradePanel from '@/components/trades/panels/SecondsTradePanel.vue'
import SpotTradePanel from '@/components/trades/panels/SpotTradePanel.vue'
import type {
  TradeCategoryConfig,
  TradeExperience,
  TradeMarketMode,
  TradeSymbolDetail,
} from '@/features/trade/tradeModel'
import type {
  DepthPayload,
  MarketTenantCategory,
  MarketTenantProduct,
  QuotePayload,
  TickPayload,
} from '@/types/market'
import type { TradeOrder, TradeSymbol } from '@/types/trade'

type ProductSheetRow = {
  key: string
  product: MarketTenantProduct
  price: string
  change: string
  direction: 'up' | 'down' | 'flat'
}
type SubmitSide = 'buy' | 'sell'
defineProps<{
  selectedCategory: MarketTenantCategory | null
  selectedProduct: MarketTenantProduct | null
  selectedProductKey: string
  tradeMarketMode: TradeMarketMode
  tradeMarketModeLabel: string
  tradeMarketModeOptions: Array<{ value: TradeMarketMode; label: string; available: boolean }>
  categoryConfig: TradeCategoryConfig
  tradeExperience: TradeExperience
  priceTrend: 'up' | 'down' | 'flat'
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
  secondsDuration: number
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
}>()

const emit = defineEmits<{
  (e: 'open-product-menu'): void
  (e: 'close-product-sheet'): void
  (e: 'select-product', product: MarketTenantProduct): void
  (e: 'select-trade-market-mode', mode: TradeMarketMode): void
  (e: 'update:orderMode', value: 'market' | 'limit'): void
  (e: 'update:tradePrice', value: string): void
  (e: 'update:tradeQty', value: string): void
  (e: 'update:tradePercent', value: number): void
  (e: 'update:marginMode', value: number): void
  (e: 'update:leverage', value: number): void
  (e: 'update:seconds-duration', value: number): void
  (e: 'update:takeProfitPrice', value: string): void
  (e: 'update:stopLossPrice', value: string): void
  (e: 'submit-order', side: SubmitSide): void
  (e: 'cancel-order', order: TradeOrder): void
  (e: 'refresh-orders'): void
}>()
</script>

<template>
  <div class="trade-view">
    <TradeHeader
      :selected-category="selectedCategory"
      :selected-product="selectedProduct"
      :selected-trade-symbol="selectedTradeSymbol"
      :selected-product-key="selectedProductKey"
      :trade-market-mode="tradeMarketMode"
      :trade-market-mode-label="tradeMarketModeLabel"
      :trade-market-mode-options="tradeMarketModeOptions"
      :category-config="categoryConfig"
      :price-trend="priceTrend"
      :placeholder-price="placeholderPrice"
      :placeholder-change="placeholderChange"
      :product-menu-open="productMenuOpen"
      :product-sheet-rows="productSheetRows"
      @open-product-menu="emit('open-product-menu')"
      @close-product-sheet="emit('close-product-sheet')"
      @select-product="emit('select-product', $event)"
      @select-trade-market-mode="emit('select-trade-market-mode', $event)"
    />

    <section
      :class="{
        'contract-layout': categoryConfig.showDepthPreview && tradeExperience !== 'seconds',
      }"
    >
      <SpotTradePanel
        v-if="tradeExperience === 'spot'"
        :selected-product="selectedProduct"
        :order-mode="orderMode"
        :selected-trade-symbol="selectedTradeSymbol"
        :trade-symbol-loading="tradeSymbolLoading"
        :is-logged-in="isLoggedIn"
        :trade-available="tradeAvailable"
        :trade-price="tradePrice"
        :trade-qty="tradeQty"
        :trade-percent="tradePercent"
        :reference-price="selectedQuote?.lastPrice || placeholderPrice"
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
        @submit-order="emit('submit-order', $event)"
      />

      <SecondsTradePanel
        v-else-if="tradeExperience === 'seconds'"
        :seconds-configs="tradeSymbolDetail?.secondsConfigs || []"
        :seconds-duration="secondsDuration"
        :trade-qty="tradeQty"
        :settle-asset="settleAsset"
        :available-balance="availableBalance"
        :trade-symbol-loading="tradeSymbolLoading"
        :is-logged-in="isLoggedIn"
        :trade-available="tradeAvailable"
        :trade-message="tradeMessage"
        :trade-error="tradeError"
        :submitting-side="submittingSide"
        @update:seconds-duration="emit('update:seconds-duration', $event)"
        @update:trade-qty="emit('update:tradeQty', $event)"
        @submit-order="emit('submit-order', $event)"
      />

      <ContractTradePanel
        v-else
        :selected-product="selectedProduct"
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

      <TradeDepthPreview
        v-if="categoryConfig.showDepthPreview && tradeExperience !== 'seconds'"
        :selected-product="selectedProduct"
        :depth-snapshot="depthSnapshot"
        :selected-quote="selectedQuote"
        :tick-snapshot="tickSnapshot"
        :placeholder-price="placeholderPrice"
        :quote-asset="selectedTradeSymbol?.quoteAsset || selectedProduct?.quoteCoin || ''"
      />
    </section>

    <TradeOrdersPanel
      :show-premarket="categoryConfig.showPremarket"
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
.trade-view {
  min-width: 0;
  max-width: 100%;
  overflow-x: hidden;
}

.contract-layout {
  display: grid;
  grid-template-columns: minmax(0, 1.1fr) minmax(112px, 0.9fr);
  gap: 12px;
  min-width: 0;
  align-items: start;
}

@media (max-width: 340px) {
  .contract-layout {
    grid-template-columns: 1fr;
  }
}
</style>
