<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'

import MarketChartView from '@/components/markets/ChartView.vue'
import MarketQuotesView from '@/components/markets/QuotesView.vue'
import MarketTopTabs from '@/components/markets/TopTabs.vue'
import type { MarketTopTab, MarketTopTabItem } from '@/components/markets/types'
import { getAccessToken } from '@/api/http'
import { useTradingDesk } from '@/composables/useTradingDesk'
import { t } from '@/i18n'
import type { ItickTenantProduct } from '@/types/itick'
import { marketCategoryLabel } from '@/utils/marketCategory'

const topTabs: MarketTopTabItem[] = [
  { key: 'watchlist', label: 'market.watchlist' },
  { key: 'markets', label: 'market.quotes' },
  { key: 'chart', label: 'market.chart' },
]

const activeTopTab = ref<MarketTopTab>('markets')
const marketsPageRef = ref<HTMLElement | null>(null)
const mobileTabsCollapsed = ref(false)
const mobileTabsCollapseProgress = ref(0)
let mobileScrollContainer: HTMLElement | null = null
let collapseRaf = 0

const mobileTopTabsHeight = 88

const detailVisible = computed(() => activeTopTab.value === 'chart')
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
  selectedQuote,
  marketRows,
  loadPreviousKlinePage,
  selectCategory,
  selectProduct,
  selectInterval,
} = useTradingDesk({
  detailVisible,
  tickLimit: 12,
})

function openProductChart(product: ItickTenantProduct) {
  selectProduct(product)
  activeTopTab.value = 'chart'
}

function updateMobileTabsCollapse() {
  collapseRaf = 0

  if (activeTopTab.value !== 'markets') {
    mobileTabsCollapseProgress.value = 0
    mobileTabsCollapsed.value = false
    return
  }

  const pageRect = marketsPageRef.value?.getBoundingClientRect()
  const scrollRect = mobileScrollContainer?.getBoundingClientRect()
  const scrollTop = mobileScrollContainer?.scrollTop || 0
  const pageOffset = pageRect && scrollRect ? pageRect.top - scrollRect.top : -scrollTop
  const pageScroll = Math.max(0, -pageOffset, scrollTop)
  const progress = Math.min(pageScroll / mobileTopTabsHeight, 1)

  mobileTabsCollapseProgress.value = progress
  mobileTabsCollapsed.value = progress >= 1
}

function requestMobileTabsCollapseUpdate() {
  if (collapseRaf) return
  collapseRaf = window.requestAnimationFrame(updateMobileTabsCollapse)
}

function bindMobileScrollContainer() {
  mobileScrollContainer?.removeEventListener('scroll', requestMobileTabsCollapseUpdate)
  mobileScrollContainer =
    (marketsPageRef.value?.closest('.page-content') as HTMLElement | null) ||
    document.querySelector<HTMLElement>('.page-content')
  mobileScrollContainer?.addEventListener('scroll', requestMobileTabsCollapseUpdate, {
    passive: true,
  })
  window.addEventListener('scroll', requestMobileTabsCollapseUpdate, { passive: true })
  window.addEventListener('resize', requestMobileTabsCollapseUpdate, { passive: true })
  updateMobileTabsCollapse()
}

onMounted(bindMobileScrollContainer)

onBeforeUnmount(() => {
  mobileScrollContainer?.removeEventListener('scroll', requestMobileTabsCollapseUpdate)
  window.removeEventListener('scroll', requestMobileTabsCollapseUpdate)
  window.removeEventListener('resize', requestMobileTabsCollapseUpdate)
  if (collapseRaf) window.cancelAnimationFrame(collapseRaf)
})

watch(activeTopTab, () => {
  updateMobileTabsCollapse()
})
</script>

<template>
  <section
    ref="marketsPageRef"
    class="markets-page"
    :class="{
      'markets-page--tabs-collapsed': mobileTabsCollapsed,
      'markets-page--chart': activeTopTab === 'chart',
    }"
    :aria-busy="loadingBootstrap"
  >
    <MarketTopTabs
      :tabs="topTabs"
      :active-tab="activeTopTab"
      :collapsed="mobileTabsCollapsed"
      :collapse-progress="mobileTabsCollapseProgress"
      @change="activeTopTab = $event"
    />

    <div v-if="activeTopTab === 'markets'" class="markets-page__mobile">
      <MarketQuotesView
        :categories="categories"
        :selected-category-type="selectedCategoryType"
        :selected-category-name="marketCategoryLabel(selectedCategory)"
        :selected-category-code="selectedCategoryCode"
        :ws-state="wsState"
        :ws-error="wsError"
        :loading="loadingBootstrap"
        :rows="marketRows"
        :selected-product-key="selectedProductKey"
        :category-pinned="mobileTabsCollapsed"
        @select-category="selectCategory"
        @select-product="openProductChart"
      />
    </div>

    <div v-else-if="activeTopTab === 'watchlist'" class="markets-page__watchlist">
      <div v-if="isLoggedIn" class="watchlist-empty">
        {{ t('market.watchlistPreparing') }}
      </div>
      <div v-else class="watchlist-empty">
        {{ t('market.watchlistLogin') }}
      </div>
    </div>

    <div v-else class="markets-page__mobile markets-page__mobile--chart">
      <MarketChartView
        :products="products"
        :rows="marketRows"
        :category-name="marketCategoryLabel(selectedCategory)"
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
  </section>
</template>

<style scoped>
.markets-page {
  width: 100%;
  max-width: 100%;
  min-height: calc(100dvh - 72px);
  padding: 18px 22px 112px;
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

@media (min-width: 0) {
  .markets-page {
    min-height: 100%;
    padding: 0 0 calc(96px + env(safe-area-inset-bottom));
  }

  .markets-page__mobile {
    min-height: 100%;
  }

  .markets-page--chart {
    padding-bottom: 0;
  }

  :global(.page-content:has(.markets-page--chart)) {
    padding-bottom: 0;
  }
}
</style>
